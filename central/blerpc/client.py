"""High-level RPC client for blerpc."""

from __future__ import annotations

import asyncio
import logging

from blerpc_protocol.command import CommandPacket, CommandType
from blerpc_protocol.container import (
    BLERPC_ERROR_RESPONSE_TOO_LARGE,
    Container,
    ContainerAssembler,
    ContainerSplitter,
    ContainerType,
    ControlCmd,
    make_capabilities_request,
    make_stream_end_c2p,
    make_stream_end_p2c,
    make_timeout_request,
)

from .generated import blerpc_pb2
from .generated.generated_client import GeneratedClientMixin
from .transport import BleTransport

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
    """Raised when the peripheral reports that the response exceeds its max_response_payload_size."""


class BlerpcClient(GeneratedClientMixin):
    """High-level RPC client that communicates over BLE."""

    def __init__(self):
        self._transport = BleTransport()
        self._splitter: ContainerSplitter | None = None
        self._assembler = ContainerAssembler()
        self._timeout_s = 0.1  # Default 100ms
        self._max_request_payload_size: int | None = None
        self._max_response_payload_size: int | None = None

    @property
    def mtu(self) -> int:
        return self._transport.mtu

    @property
    def max_request_payload_size(self) -> int | None:
        return self._max_request_payload_size

    @property
    def max_response_payload_size(self) -> int | None:
        return self._max_response_payload_size

    async def connect(self, device_name: str = "blerpc", timeout: float = 10.0):
        """Connect to the blerpc peripheral."""
        await self._transport.connect(device_name=device_name, timeout=timeout)
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

    async def _request_timeout(self):
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

    async def _request_capabilities(self):
        """Request capabilities from peripheral."""
        tid = self._splitter.next_transaction_id()
        req = make_capabilities_request(transaction_id=tid)
        await self._transport.write(req.serialize())
        data = await self._transport.read_notify(timeout=1.0)
        resp = Container.deserialize(data)
        if (
            resp.container_type == ContainerType.CONTROL
            and resp.control_cmd == ControlCmd.CAPABILITIES
            and len(resp.payload) == 4
        ):
            self._max_request_payload_size = int.from_bytes(resp.payload[0:2], "little")
            self._max_response_payload_size = int.from_bytes(
                resp.payload[2:4], "little"
            )
            logger.info(
                "Peripheral capabilities: max_request=%d, max_response=%d",
                self._max_request_payload_size,
                self._max_response_payload_size,
            )

    async def _call(self, cmd_name: str, request_data: bytes) -> bytes:
        """Execute an RPC call and return response data."""
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

        # Split into containers and send
        containers = self._splitter.split(payload)
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

        # Decode command response
        resp = CommandPacket.deserialize(result)
        if resp.cmd_type != CommandType.RESPONSE:
            raise RuntimeError(f"Expected response, got type={resp.cmd_type}")
        if resp.cmd_name != cmd_name:
            raise RuntimeError(
                f"Command name mismatch: expected '{cmd_name}', got '{resp.cmd_name}'"
            )

        return resp.data

    async def stream_receive(self, cmd_name: str, request_data: bytes):
        """P->C stream: send request, yield response data until STREAM_END_P2C.

        Each yielded bytes object is the protobuf-encoded data portion
        of a single CommandPacket response.
        """
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

        containers = self._splitter.split(payload)
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
                resp = CommandPacket.deserialize(result)
                if resp.cmd_type != CommandType.RESPONSE:
                    raise RuntimeError(
                        f"Expected response, got type={resp.cmd_type}"
                    )
                yield resp.data

    async def stream_send(
        self,
        cmd_name: str,
        messages: list[bytes],
        final_cmd_name: str,
    ) -> bytes:
        """C->P stream: send multiple requests, then STREAM_END_C2P, return final response data.

        Each item in messages is protobuf-encoded request data.
        After sending all messages + STREAM_END_C2P, waits for a final response
        with cmd_name=final_cmd_name and returns its data.
        """
        # Send each message as an independent request
        for msg_data in messages:
            cmd = CommandPacket(
                cmd_type=CommandType.REQUEST,
                cmd_name=cmd_name,
                data=msg_data,
            )
            payload = cmd.serialize()
            containers = self._splitter.split(payload)
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

        resp = CommandPacket.deserialize(result)
        if resp.cmd_type != CommandType.RESPONSE:
            raise RuntimeError(f"Expected response, got type={resp.cmd_type}")
        if resp.cmd_name != final_cmd_name:
            raise RuntimeError(
                f"Command name mismatch: expected '{final_cmd_name}', got '{resp.cmd_name}'"
            )
        return resp.data

    async def counter_stream(self, count: int) -> list:
        """P->C stream: request count counter values, return list of (seq, value) tuples."""
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
        resp_data = await self.stream_send(
            "counter_upload", messages, "counter_upload"
        )
        resp = blerpc_pb2.CounterUploadResponse()
        resp.ParseFromString(resp_data)
        return resp.received_count

    async def echo(self, message: str) -> str:
        """Call the echo command."""
        req = blerpc_pb2.EchoRequest(message=message)
        resp_data = await self._call("echo", req.SerializeToString())
        resp = blerpc_pb2.EchoResponse()
        resp.ParseFromString(resp_data)
        return resp.message

    async def flash_read(self, address: int, length: int) -> bytes:
        """Call the flash_read command."""
        req = blerpc_pb2.FlashReadRequest(address=address, length=length)
        resp_data = await self._call("flash_read", req.SerializeToString())
        resp = blerpc_pb2.FlashReadResponse()
        resp.ParseFromString(resp_data)
        return resp.data

    async def data_write(self, data: bytes) -> int:
        """Call the data_write command. Returns confirmed length."""
        req = blerpc_pb2.DataWriteRequest(data=data)
        resp_data = await self._call("data_write", req.SerializeToString())
        resp = blerpc_pb2.DataWriteResponse()
        resp.ParseFromString(resp_data)
        return resp.length

    async def disconnect(self):
        """Disconnect from the peripheral."""
        await self._transport.disconnect()
