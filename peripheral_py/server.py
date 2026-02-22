"""blerpc Python Peripheral Server using bless.

Acts as a BLE peripheral (GATT server) on macOS, handling echo and flash_read
RPCs. For role-reversal testing with nRF54L15 as Central.
"""

import asyncio
import logging
import os
import struct
import sys
import threading
import time

from blerpc_protocol.command import CommandPacket, CommandType
from blerpc_protocol.container import (
    BLERPC_ERROR_RESPONSE_TOO_LARGE,
    CAPABILITY_FLAG_ENCRYPTION_SUPPORTED,
    Container,
    ContainerAssembler,
    ContainerSplitter,
    ContainerType,
    ControlCmd,
    make_stream_end_p2c,
)
from blerpc_protocol.crypto import (
    BlerpcCrypto,
    BlerpcCryptoSession,
    PeripheralKeyExchange,
)

# Import protobuf definitions from central_py/blerpc/
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "central_py"))
from blerpc.generated import blerpc_pb2
from bless import (
    BlessGATTCharacteristic,
    BlessServer,
    GATTAttributePermissions,
    GATTCharacteristicProperties,
)
from generated_handlers import HANDLERS as _GENERATED_HANDLERS

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("blerpc-peripheral")

SERVICE_UUID = "12340001-0000-1000-8000-00805f9b34fb"
CHAR_UUID = "12340002-0000-1000-8000-00805f9b34fb"
TIMEOUT_MS = 100
MTU = 247
MAX_RESPONSE_PAYLOAD_SIZE = 65535
NOTIFY_MAX_RETRIES = 50
NOTIFY_RETRY_DELAY_S = 0.005


HANDLERS = dict(_GENERATED_HANDLERS)


def handle_echo(req_data: bytes) -> bytes:
    req = blerpc_pb2.EchoRequest()
    req.ParseFromString(req_data)
    logger.info("Echo: '%s'", req.message)
    resp = blerpc_pb2.EchoResponse(message=req.message)
    return resp.SerializeToString()


def handle_flash_read(req_data: bytes) -> bytes:
    req = blerpc_pb2.FlashReadRequest()
    req.ParseFromString(req_data)
    logger.info("FlashRead: addr=0x%08x len=%d", req.address, req.length)
    data = os.urandom(req.length)
    resp = blerpc_pb2.FlashReadResponse(address=req.address, data=data)
    return resp.SerializeToString()


def handle_data_write(req_data: bytes) -> bytes:
    req = blerpc_pb2.DataWriteRequest()
    req.ParseFromString(req_data)
    logger.info("DataWrite: received %d bytes", len(req.data))
    return blerpc_pb2.DataWriteResponse(length=len(req.data)).SerializeToString()


def handle_counter_upload(req_data: bytes) -> bytes:
    """Accumulate counter_upload requests (called per message)."""
    req = blerpc_pb2.CounterUploadRequest()
    req.ParseFromString(req_data)
    logger.debug("CounterUpload: seq=%d value=%d", req.seq, req.value)
    # Accumulation handled in BlerpcPeripheral._upload_count
    return None  # Signal: no response for this message


HANDLERS["echo"] = handle_echo
HANDLERS["flash_read"] = handle_flash_read
HANDLERS["data_write"] = handle_data_write
HANDLERS["counter_upload"] = handle_counter_upload


class BlerpcPeripheral:
    def __init__(
        self,
        ed25519_private_key_hex: str | None = None,
    ):
        self.server: BlessServer | None = None
        self.assembler = ContainerAssembler()
        self.splitter = ContainerSplitter(mtu=MTU)
        self._loop: asyncio.AbstractEventLoop | None = None
        self._send_queue: list[tuple[bytes, int]] = []
        self._send_lock = threading.Lock()
        self._upload_count = 0

        # Encryption state
        self._encryption_supported = False
        self._session: BlerpcCryptoSession | None = None
        self._kx: PeripheralKeyExchange | None = None
        self._ed25519_privkey = None  # Store for KX recreation on disconnect
        self._connected = False

        if ed25519_private_key_hex:
            ed25519_priv_bytes = bytes.fromhex(ed25519_private_key_hex)
            ed25519_privkey = BlerpcCrypto.ed25519_private_from_bytes(
                ed25519_priv_bytes
            )
            self._ed25519_privkey = ed25519_privkey
            self._kx = PeripheralKeyExchange(ed25519_privkey)
            self._encryption_supported = True
            logger.info("Encryption key loaded (X25519 generated per session)")

    async def start(self):
        self._loop = asyncio.get_event_loop()
        self.server = BlessServer(name="blerpc", loop=self._loop)
        self.server.write_request_func = self._on_write

        await self.server.add_new_service(SERVICE_UUID)

        char_flags = (
            GATTCharacteristicProperties.write_without_response
            | GATTCharacteristicProperties.notify
        )
        permissions = (
            GATTAttributePermissions.readable | GATTAttributePermissions.writeable
        )
        await self.server.add_new_characteristic(
            SERVICE_UUID,
            CHAR_UUID,
            char_flags,
            None,
            permissions,
        )

        await self.server.start()
        logger.info("Advertising as 'blerpc' — waiting for connections...")

    def _reset_connection_state(self):
        """Reset session state for new connection."""
        logger.info("Resetting connection state")
        self._session = None
        self._upload_count = 0
        self.assembler = ContainerAssembler()
        if self._ed25519_privkey is not None:
            self._kx = PeripheralKeyExchange(self._ed25519_privkey)

    def _on_write(
        self, characteristic: BlessGATTCharacteristic, value: bytearray, **kwargs
    ):
        data = bytes(value)
        logger.debug("Write received: %d bytes", len(data))

        # Detect new connection by tracking connection state.
        # bless doesn't provide disconnect callbacks, so we detect new
        # connections by watching for a CAPABILITIES request (always the first
        # thing a central sends). If we were previously connected and get a new
        # CAPABILITIES, reset state.
        container = Container.deserialize(data)
        if (
            container.container_type == ContainerType.CONTROL
            and container.control_cmd == ControlCmd.CAPABILITIES
            and self._connected
        ):
            self._reset_connection_state()
        self._connected = True

        # Handle control containers
        if container.container_type == ContainerType.CONTROL:
            if container.control_cmd == ControlCmd.TIMEOUT:
                resp = Container(
                    transaction_id=container.transaction_id,
                    sequence_number=0,
                    container_type=ContainerType.CONTROL,
                    control_cmd=ControlCmd.TIMEOUT,
                    payload=struct.pack("<H", TIMEOUT_MS),
                )
                self._send_container_sync(resp)
            elif container.control_cmd == ControlCmd.STREAM_END_C2P:
                threading.Thread(
                    target=self._handle_stream_end_c2p,
                    args=(container.transaction_id,),
                    daemon=True,
                ).start()
            elif container.control_cmd == ControlCmd.CAPABILITIES:
                flags = 0
                if self._encryption_supported:
                    flags |= CAPABILITY_FLAG_ENCRYPTION_SUPPORTED
                logger.info(
                    "Capabilities request, max_req=65535 max_resp=%d flags=0x%04x",
                    MAX_RESPONSE_PAYLOAD_SIZE,
                    flags,
                )
                resp = Container(
                    transaction_id=container.transaction_id,
                    sequence_number=0,
                    container_type=ContainerType.CONTROL,
                    control_cmd=ControlCmd.CAPABILITIES,
                    payload=struct.pack(
                        "<HHH", 65535, MAX_RESPONSE_PAYLOAD_SIZE, flags
                    ),
                )
                self._send_container_sync(resp)
            elif container.control_cmd == ControlCmd.KEY_EXCHANGE:
                self._handle_key_exchange(container)
            return

        # Feed into assembler
        result = self.assembler.feed(container)
        if result is not None:
            # Process in a separate thread to avoid blocking CoreBluetooth callback
            tid = container.transaction_id
            threading.Thread(
                target=self._process_request_thread,
                args=(result, tid),
                daemon=True,
            ).start()

    def _handle_key_exchange(self, container: Container):
        """Handle KEY_EXCHANGE control containers."""
        if not self._encryption_supported or self._kx is None:
            logger.warning("KEY_EXCHANGE received but encryption not supported")
            return

        # Block KX re-initiation when session already exists
        if self._session is not None:
            logger.warning("KEY_EXCHANGE rejected: encryption already active")
            return

        try:
            response, session = self._kx.handle_step(container.payload)
        except ValueError as e:
            logger.error("Key exchange failed: %s", e)
            return

        resp = Container(
            transaction_id=container.transaction_id,
            sequence_number=0,
            container_type=ContainerType.CONTROL,
            control_cmd=ControlCmd.KEY_EXCHANGE,
            payload=response,
        )
        self._send_container_sync(resp)

        if session is not None:
            self._session = session
            logger.info("E2E encryption established")

    def _process_request_thread(self, payload: bytes, transaction_id: int):
        try:
            self._process_request(payload, transaction_id)
        except Exception:
            logger.exception("Error processing request")

    def _process_request(self, payload: bytes, transaction_id: int):
        # Decrypt if encryption is active
        if self._session is not None:
            try:
                payload = self._session.decrypt(payload)
            except RuntimeError as e:
                logger.error("Decryption/replay error: %s", e)
                return
        elif self._encryption_supported:
            # Reject unencrypted data when encryption is supported
            logger.warning(
                "Rejecting unencrypted payload"
                " (encryption supported but not active)"
            )
            return

        cmd = CommandPacket.deserialize(payload)
        if cmd.cmd_type != CommandType.REQUEST:
            logger.error("Expected request, got type=%d", cmd.cmd_type)
            return

        # Handle counter_stream specially (P→C stream)
        if cmd.cmd_name == "counter_stream":
            self._handle_counter_stream(cmd.data)
            return

        handler = HANDLERS.get(cmd.cmd_name)
        if not handler:
            logger.error("Unknown command: '%s'", cmd.cmd_name)
            return

        resp_data = handler(cmd.data)

        # counter_upload returns None (no individual response)
        if resp_data is None:
            self._upload_count += 1
            return

        resp_cmd = CommandPacket(
            cmd_type=CommandType.RESPONSE,
            cmd_name=cmd.cmd_name,
            data=resp_data,
        )
        resp_payload = resp_cmd.serialize()

        if len(resp_payload) > MAX_RESPONSE_PAYLOAD_SIZE:
            err = Container(
                transaction_id=transaction_id,
                sequence_number=0,
                container_type=ContainerType.CONTROL,
                control_cmd=ControlCmd.ERROR,
                payload=bytes([BLERPC_ERROR_RESPONSE_TOO_LARGE]),
            )
            self._send_container_sync(err)
            logger.warning(
                "Response too large: %d > %d",
                len(resp_payload),
                MAX_RESPONSE_PAYLOAD_SIZE,
            )
            return

        send_payload = self._maybe_encrypt(resp_payload)
        containers = self.splitter.split(send_payload, transaction_id=transaction_id)
        logger.info(
            "Sending %d containers (%d bytes payload)",
            len(containers),
            len(send_payload),
        )
        for c in containers:
            self._send_container_sync(c)

    def _maybe_encrypt(self, payload: bytes) -> bytes:
        """Encrypt payload if encryption is active, otherwise return as-is."""
        if self._session is None:
            return payload
        return self._session.encrypt(payload)

    def _handle_counter_stream(self, req_data: bytes):
        """Handle counter_stream: send N responses + STREAM_END_P2C."""
        req = blerpc_pb2.CounterStreamRequest()
        req.ParseFromString(req_data)
        logger.info("CounterStream: count=%d", req.count)

        for i in range(req.count):
            resp = blerpc_pb2.CounterStreamResponse(seq=i, value=i * 10)
            resp_cmd = CommandPacket(
                cmd_type=CommandType.RESPONSE,
                cmd_name="counter_stream",
                data=resp.SerializeToString(),
            )
            resp_payload = resp_cmd.serialize()
            send_payload = self._maybe_encrypt(resp_payload)
            tid = self.splitter.next_transaction_id()
            containers = self.splitter.split(send_payload, transaction_id=tid)
            for c in containers:
                self._send_container_sync(c)

        # Send STREAM_END_P2C
        tid = self.splitter.next_transaction_id()
        stream_end = make_stream_end_p2c(transaction_id=tid)
        self._send_container_sync(stream_end)
        logger.info("CounterStream: sent %d responses + STREAM_END_P2C", req.count)

    def _handle_stream_end_c2p(self, transaction_id: int):
        """Handle STREAM_END_C2P: send final counter_upload response."""
        count = self._upload_count
        self._upload_count = 0
        logger.info(
            "STREAM_END_C2P: sending counter_upload response, received_count=%d",
            count,
        )

        resp = blerpc_pb2.CounterUploadResponse(received_count=count)
        resp_cmd = CommandPacket(
            cmd_type=CommandType.RESPONSE,
            cmd_name="counter_upload",
            data=resp.SerializeToString(),
        )
        resp_payload = resp_cmd.serialize()
        send_payload = self._maybe_encrypt(resp_payload)
        tid = self.splitter.next_transaction_id()
        containers = self.splitter.split(send_payload, transaction_id=tid)
        for c in containers:
            self._send_container_sync(c)

    def _send_container_sync(self, container: Container):
        data = container.serialize()
        char = self.server.get_characteristic(CHAR_UUID)
        char.value = data
        for attempt in range(NOTIFY_MAX_RETRIES):
            result = self.server.update_value(SERVICE_UUID, CHAR_UUID)
            if result:
                break
            time.sleep(NOTIFY_RETRY_DELAY_S)
        else:
            logger.error("update_value failed after %d retries", NOTIFY_MAX_RETRIES)

    async def stop(self):
        if self.server:
            await self.server.stop()


async def main():
    ed25519_key = os.environ.get("BLERPC_ED25519_KEY")
    peripheral = BlerpcPeripheral(
        ed25519_private_key_hex=ed25519_key,
    )
    await peripheral.start()

    try:
        while True:
            await asyncio.sleep(1)
    except KeyboardInterrupt:
        pass
    finally:
        await peripheral.stop()


if __name__ == "__main__":
    asyncio.run(main())
