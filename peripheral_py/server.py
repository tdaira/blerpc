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

# Import blerpc protocol layers from central/blerpc/
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "central"))

from blerpc.command import CommandPacket, CommandType
from blerpc.container import (
    Container,
    ContainerAssembler,
    ContainerSplitter,
    ContainerType,
    ControlCmd,
)
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


HANDLERS["echo"] = handle_echo
HANDLERS["flash_read"] = handle_flash_read
HANDLERS["data_write"] = handle_data_write


class BlerpcPeripheral:
    def __init__(self):
        self.server: BlessServer | None = None
        self.assembler = ContainerAssembler()
        self.splitter = ContainerSplitter(mtu=MTU)
        self._loop: asyncio.AbstractEventLoop | None = None
        self._send_queue: list[tuple[bytes, int]] = []
        self._send_lock = threading.Lock()

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

    def _on_write(
        self, characteristic: BlessGATTCharacteristic, value: bytearray, **kwargs
    ):
        data = bytes(value)
        logger.debug("Write received: %d bytes", len(data))

        container = Container.deserialize(data)

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

    def _process_request_thread(self, payload: bytes, transaction_id: int):
        try:
            self._process_request(payload, transaction_id)
        except Exception:
            logger.exception("Error processing request")

    def _process_request(self, payload: bytes, transaction_id: int):
        cmd = CommandPacket.deserialize(payload)
        if cmd.cmd_type != CommandType.REQUEST:
            logger.error("Expected request, got type=%d", cmd.cmd_type)
            return

        handler = HANDLERS.get(cmd.cmd_name)
        if not handler:
            logger.error("Unknown command: '%s'", cmd.cmd_name)
            return

        resp_data = handler(cmd.data)

        resp_cmd = CommandPacket(
            cmd_type=CommandType.RESPONSE,
            cmd_name=cmd.cmd_name,
            data=resp_data,
        )
        resp_payload = resp_cmd.serialize()

        containers = self.splitter.split(resp_payload, transaction_id=transaction_id)
        logger.info(
            "Sending %d containers (%d bytes payload)",
            len(containers),
            len(resp_payload),
        )
        for c in containers:
            self._send_container_sync(c)

    def _send_container_sync(self, container: Container):
        data = container.serialize()
        char = self.server.get_characteristic(CHAR_UUID)
        char.value = data
        for attempt in range(50):
            result = self.server.update_value(SERVICE_UUID, CHAR_UUID)
            if result:
                break
            time.sleep(0.005)
        else:
            logger.error("update_value failed after 50 retries")

    async def stop(self):
        if self.server:
            await self.server.stop()


async def main():
    peripheral = BlerpcPeripheral()
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
