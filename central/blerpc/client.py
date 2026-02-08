"""High-level RPC client for blerpc."""

from __future__ import annotations

import asyncio
import logging

from blerpc_protocol.command import CommandPacket, CommandType
from blerpc_protocol.container import (
    Container,
    ContainerAssembler,
    ContainerSplitter,
    ContainerType,
    ControlCmd,
    make_timeout_request,
)

from .generated import blerpc_pb2
from .generated.generated_client import GeneratedClientMixin
from .transport import BleTransport

logger = logging.getLogger(__name__)


class BlerpcClient(GeneratedClientMixin):
    """High-level RPC client that communicates over BLE."""

    def __init__(self):
        self._transport = BleTransport()
        self._splitter: ContainerSplitter | None = None
        self._assembler = ContainerAssembler()
        self._timeout_s = 0.1  # Default 100ms

    @property
    def mtu(self) -> int:
        return self._transport.mtu

    async def connect(self, device_name: str = "blerpc", timeout: float = 10.0):
        """Connect to the blerpc peripheral."""
        await self._transport.connect(device_name=device_name, timeout=timeout)
        self._splitter = ContainerSplitter(mtu=self._transport.mtu)

        # Optionally request timeout from peripheral
        try:
            await self._request_timeout()
        except asyncio.TimeoutError:
            logger.debug("Peripheral did not respond to timeout request, using default")

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

    async def _call(self, cmd_name: str, request_data: bytes) -> bytes:
        """Execute an RPC call and return response data."""
        # Encode command
        cmd = CommandPacket(
            cmd_type=CommandType.REQUEST,
            cmd_name=cmd_name,
            data=request_data,
        )
        payload = cmd.serialize()

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
                continue  # Skip control containers

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
