"""High-level RPC client for blerpc."""

from __future__ import annotations

import asyncio
import logging
from collections.abc import AsyncIterator

from blerpc_protocol.command import CommandPacket, CommandType
from blerpc_protocol.container import (
    BLERPC_ERROR_RESPONSE_TOO_LARGE,
    CAPABILITY_FLAG_ENCRYPTION_SUPPORTED,
    Container,
    ContainerAssembler,
    ContainerSplitter,
    ContainerType,
    ControlCmd,
    make_capabilities_request,
    make_key_exchange,
    make_stream_end_c2p,
    make_timeout_request,
)
from blerpc_protocol.crypto import BlerpcCryptoSession, central_perform_key_exchange

from .generated import blerpc_pb2
from .generated.generated_client import GeneratedClientMixin
from .transport import SERVICE_UUID, BleTransport, ScannedDevice

logger = logging.getLogger(__name__)


class PayloadTooLargeError(Exception):
    """Raised when a request payload exceeds the peripheral's max_payload_size."""

    def __init__(self, actual: int, limit: int):
        self.actual = actual
        self.limit = limit
        super().__init__(
            f"Request payload ({actual} bytes) exceeds peripheral limit ({limit} bytes)"
        )


class ResponseTooLargeError(Exception):
    """Raised when the response exceeds max_response_payload_size."""


class BlerpcClient(GeneratedClientMixin):
    """High-level RPC client that communicates over BLE."""

    def __init__(
        self,
        known_keys_path: str | None = None,
        require_encryption: bool = True,
    ):
        self._transport = BleTransport()
        self._splitter: ContainerSplitter | None = None
        self._assembler = ContainerAssembler()
        self._timeout_s = 0.1  # Default 100ms
        self._max_request_payload_size: int | None = None
        self._max_response_payload_size: int | None = None

        # Encryption state
        self._session: BlerpcCryptoSession | None = None
        self._known_keys_path = known_keys_path
        self._require_encryption = require_encryption

    @property
    def mtu(self) -> int:
        return self._transport.mtu

    @property
    def max_request_payload_size(self) -> int | None:
        return self._max_request_payload_size

    @property
    def max_response_payload_size(self) -> int | None:
        return self._max_response_payload_size

    @property
    def is_encrypted(self) -> bool:
        return self._session is not None

    async def scan(
        self,
        timeout: float = 5.0,
        service_uuid: str | None = SERVICE_UUID,
    ) -> list[ScannedDevice]:
        """Scan for BLE devices."""
        return await self._transport.scan(timeout=timeout, service_uuid=service_uuid)

    async def connect(self, device: ScannedDevice) -> None:
        """Connect to a previously scanned device."""
        await self._transport.connect(device)
        self._splitter = ContainerSplitter(mtu=self._transport.mtu)

        # Optionally request timeout from peripheral
        try:
            await self._request_timeout()
        except asyncio.TimeoutError:
            logger.debug("Peripheral did not respond to timeout request, using default")

        # Optionally request capabilities from peripheral
        try:
            await self._request_capabilities()
        except asyncio.TimeoutError:
            logger.debug("Peripheral did not respond to capabilities request")

        if self._require_encryption and self._session is None:
            raise RuntimeError(
                "Encryption required but key exchange was not completed. "
                "The peripheral may not support encryption or a MitM may "
                "have stripped the encryption capability flag."
            )

    async def _request_timeout(self) -> None:
        """Request timeout value from peripheral."""
        tid = self._splitter.next_transaction_id()
        req = make_timeout_request(transaction_id=tid)
        await self._transport.write(req.serialize())
        data = await self._transport.read_notify(timeout=1.0)
        resp = Container.deserialize(data)
        if (
            resp.container_type == ContainerType.CONTROL
            and resp.control_cmd == ControlCmd.TIMEOUT
            and len(resp.payload) == 2
        ):
            timeout_ms = int.from_bytes(resp.payload, "little")
            self._timeout_s = timeout_ms / 1000.0
            logger.info("Peripheral timeout: %dms", timeout_ms)
        else:
            logger.warning(
                "Unexpected timeout response: type=%s, cmd=%s, payload_len=%d",
                resp.container_type,
                resp.control_cmd,
                len(resp.payload),
            )

    async def _request_capabilities(self) -> None:
        """Request capabilities from peripheral (6-byte format)."""
        tid = self._splitter.next_transaction_id()
        req = make_capabilities_request(transaction_id=tid)
        await self._transport.write(req.serialize())
        data = await self._transport.read_notify(timeout=1.0)
        resp = Container.deserialize(data)
        if (
            resp.container_type == ContainerType.CONTROL
            and resp.control_cmd == ControlCmd.CAPABILITIES
            and len(resp.payload) >= 6
        ):
            max_req = int.from_bytes(resp.payload[0:2], "little")
            max_resp = int.from_bytes(resp.payload[2:4], "little")
            flags = int.from_bytes(resp.payload[4:6], "little")
            if max_req == 0 or max_resp == 0:
                logger.warning(
                    "Peripheral reported zero capability:"
                    " max_request=%d, max_response=%d",
                    max_req,
                    max_resp,
                )
            self._max_request_payload_size = max_req
            self._max_response_payload_size = max_resp
            logger.info(
                "Peripheral capabilities: max_request=%d, "
                "max_response=%d, flags=0x%04x",
                self._max_request_payload_size,
                self._max_response_payload_size,
                flags,
            )

            # Initiate key exchange if peripheral supports encryption
            if flags & CAPABILITY_FLAG_ENCRYPTION_SUPPORTED:
                await self._perform_key_exchange()
        else:
            logger.warning(
                "Unexpected capabilities response: type=%s, cmd=%s, payload_len=%d",
                resp.container_type,
                resp.control_cmd,
                len(resp.payload),
            )

    async def _perform_key_exchange(self) -> None:
        """Perform the 4-step key exchange handshake."""

        async def send(payload: bytes) -> None:
            tid = self._splitter.next_transaction_id()
            req = make_key_exchange(transaction_id=tid, payload=payload)
            await self._transport.write(req.serialize())

        async def receive() -> bytes:
            data = await self._transport.read_notify(timeout=2.0)
            resp = Container.deserialize(data)
            if (
                resp.container_type != ContainerType.CONTROL
                or resp.control_cmd != ControlCmd.KEY_EXCHANGE
            ):
                raise ValueError("Expected KEY_EXCHANGE response, got something else")
            return resp.payload

        verify_cb = None
        if self._known_keys_path:
            from .known_keys import check_or_store_key

            def verify_cb(ed25519_pub: bytes) -> bool:
                return check_or_store_key(
                    self._known_keys_path, self._transport.address, ed25519_pub
                )

        try:
            self._session = await central_perform_key_exchange(
                send, receive, verify_key_cb=verify_cb
            )
        except ValueError as e:
            logger.error("Key exchange failed: %s", e)
            if self._require_encryption:
                raise
            return

        logger.info("E2E encryption established")

    def _encrypt_payload(self, payload: bytes) -> bytes:
        """Encrypt payload if encryption is active."""
        if self._session is None:
            if self._require_encryption:
                raise RuntimeError("Encryption required but no session established")
            return payload
        return self._session.encrypt(payload)

    def _decrypt_payload(self, payload: bytes) -> bytes:
        """Decrypt payload if encryption is active."""
        if self._session is None:
            if self._require_encryption:
                raise RuntimeError("Encryption required but no session established")
            return payload
        return self._session.decrypt(payload)

    async def _call(self, cmd_name: str, request_data: bytes) -> bytes:
        """Execute an RPC call and return response data."""
        if self._splitter is None:
            raise RuntimeError("Not connected: call connect() first")

        # Encode command
        cmd = CommandPacket(
            cmd_type=CommandType.REQUEST,
            cmd_name=cmd_name,
            data=request_data,
        )
        payload = cmd.serialize()

        if (
            self._max_request_payload_size is not None
            and len(payload) > self._max_request_payload_size
        ):
            raise PayloadTooLargeError(len(payload), self._max_request_payload_size)

        # Encrypt if active, then split into containers and send
        send_payload = self._encrypt_payload(payload)
        containers = self._splitter.split(send_payload)
        for c in containers:
            await self._transport.write(c.serialize())

        # Receive response containers
        self._assembler.reset()
        while True:
            notify_data = await self._transport.read_notify(timeout=self._timeout_s)
            container = Container.deserialize(notify_data)

            if container.container_type == ContainerType.CONTROL:
                if (
                    container.control_cmd == ControlCmd.ERROR
                    and len(container.payload) >= 1
                ):
                    error_code = container.payload[0]
                    if error_code == BLERPC_ERROR_RESPONSE_TOO_LARGE:
                        raise ResponseTooLargeError(
                            "Response exceeds peripheral's max_response_payload_size"
                        )
                    raise RuntimeError(f"Peripheral error: 0x{error_code:02x}")
                continue  # Skip other control containers

            result = self._assembler.feed(container)
            if result is not None:
                break

        # Decrypt if active
        result = self._decrypt_payload(result)

        # Decode command response
        resp = CommandPacket.deserialize(result)
        if resp.cmd_type != CommandType.RESPONSE:
            raise RuntimeError(f"Expected response, got type={resp.cmd_type}")
        if resp.cmd_name != cmd_name:
            raise RuntimeError(
                f"Command name mismatch: expected '{cmd_name}', got '{resp.cmd_name}'"
            )

        return resp.data

    async def stream_receive(
        self, cmd_name: str, request_data: bytes
    ) -> AsyncIterator[bytes]:
        """P->C stream: send request, yield response data until STREAM_END_P2C.

        Each yielded bytes object is the protobuf-encoded data portion
        of a single CommandPacket response.
        """
        if self._splitter is None:
            raise RuntimeError("Not connected: call connect() first")

        # Send initial request (same as _call send path)
        cmd = CommandPacket(
            cmd_type=CommandType.REQUEST,
            cmd_name=cmd_name,
            data=request_data,
        )
        payload = cmd.serialize()

        if (
            self._max_request_payload_size is not None
            and len(payload) > self._max_request_payload_size
        ):
            raise PayloadTooLargeError(len(payload), self._max_request_payload_size)

        send_payload = self._encrypt_payload(payload)
        containers = self._splitter.split(send_payload)
        for c in containers:
            await self._transport.write(c.serialize())

        # Receive stream responses until STREAM_END_P2C
        self._assembler.reset()
        while True:
            notify_data = await self._transport.read_notify(timeout=self._timeout_s)
            container = Container.deserialize(notify_data)

            if container.container_type == ContainerType.CONTROL:
                if container.control_cmd == ControlCmd.STREAM_END_P2C:
                    break
                if (
                    container.control_cmd == ControlCmd.ERROR
                    and len(container.payload) >= 1
                ):
                    error_code = container.payload[0]
                    if error_code == BLERPC_ERROR_RESPONSE_TOO_LARGE:
                        raise ResponseTooLargeError(
                            "Response exceeds peripheral's max_response_payload_size"
                        )
                    raise RuntimeError(f"Peripheral error: 0x{error_code:02x}")
                continue

            result = self._assembler.feed(container)
            if result is not None:
                result = self._decrypt_payload(result)
                resp = CommandPacket.deserialize(result)
                if resp.cmd_type != CommandType.RESPONSE:
                    raise RuntimeError(f"Expected response, got type={resp.cmd_type}")
                yield resp.data

    async def stream_send(
        self,
        cmd_name: str,
        messages: list[bytes],
        final_cmd_name: str,
    ) -> bytes:
        """C->P stream: send requests, STREAM_END_C2P, return response.

        Each item in messages is protobuf-encoded request data.
        After sending all messages + STREAM_END_C2P, waits for a final response
        with cmd_name=final_cmd_name and returns its data.
        """
        if self._splitter is None:
            raise RuntimeError("Not connected: call connect() first")

        # Send each message as an independent request
        for msg_data in messages:
            cmd = CommandPacket(
                cmd_type=CommandType.REQUEST,
                cmd_name=cmd_name,
                data=msg_data,
            )
            payload = cmd.serialize()
            send_payload = self._encrypt_payload(payload)
            containers = self._splitter.split(send_payload)
            for c in containers:
                await self._transport.write(c.serialize())

        # Send STREAM_END_C2P
        tid = self._splitter.next_transaction_id()
        stream_end = make_stream_end_c2p(transaction_id=tid)
        await self._transport.write(stream_end.serialize())

        # Wait for final response
        self._assembler.reset()
        while True:
            notify_data = await self._transport.read_notify(timeout=self._timeout_s)
            container = Container.deserialize(notify_data)

            if container.container_type == ContainerType.CONTROL:
                if (
                    container.control_cmd == ControlCmd.ERROR
                    and len(container.payload) >= 1
                ):
                    error_code = container.payload[0]
                    if error_code == BLERPC_ERROR_RESPONSE_TOO_LARGE:
                        raise ResponseTooLargeError(
                            "Response exceeds peripheral's max_response_payload_size"
                        )
                    raise RuntimeError(f"Peripheral error: 0x{error_code:02x}")
                continue

            result = self._assembler.feed(container)
            if result is not None:
                break

        result = self._decrypt_payload(result)
        resp = CommandPacket.deserialize(result)
        if resp.cmd_type != CommandType.RESPONSE:
            raise RuntimeError(f"Expected response, got type={resp.cmd_type}")
        if resp.cmd_name != final_cmd_name:
            raise RuntimeError(
                f"Command name mismatch: expected "
                f"'{final_cmd_name}', got '{resp.cmd_name}'"
            )
        return resp.data

    async def counter_stream(self, count: int) -> list[tuple[int, int]]:
        """P->C stream: request counter values, return (seq, value) list."""
        req = blerpc_pb2.CounterStreamRequest(count=count)
        results = []
        async for data in self.stream_receive(
            "counter_stream", req.SerializeToString()
        ):
            resp = blerpc_pb2.CounterStreamResponse()
            resp.ParseFromString(data)
            results.append((resp.seq, resp.value))
        return results

    async def counter_upload(self, count: int) -> int:
        """C->P stream: upload count counter values, return received_count."""
        messages = []
        for i in range(count):
            req = blerpc_pb2.CounterUploadRequest(seq=i, value=i * 10)
            messages.append(req.SerializeToString())
        resp_data = await self.stream_send("counter_upload", messages, "counter_upload")
        resp = blerpc_pb2.CounterUploadResponse()
        resp.ParseFromString(resp_data)
        return resp.received_count

    async def disconnect(self) -> None:
        """Disconnect from the peripheral."""
        await self._transport.disconnect()
