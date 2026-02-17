"""Unit tests for BlerpcClient with mock transport.

Tests the full protocol stack (protobuf → command → container)
without requiring BLE hardware.
"""

import asyncio

import pytest
from blerpc.client import BlerpcClient, PayloadTooLargeError, ResponseTooLargeError
from blerpc.generated import blerpc_pb2
from blerpc_protocol.command import CommandPacket, CommandType
from blerpc_protocol.container import (
    BLERPC_ERROR_RESPONSE_TOO_LARGE,
    Container,
    ContainerSplitter,
    ContainerType,
    ControlCmd,
)


class MockTransport:
    """Mock transport that simulates a peripheral."""

    def __init__(self, mtu: int = 247):
        self._mtu = mtu
        self._written: list[bytes] = []
        self._notify_queue: asyncio.Queue[bytes] = asyncio.Queue()
        self._handler = None  # Callable to process requests

    @property
    def mtu(self) -> int:
        return self._mtu

    @property
    def is_connected(self) -> bool:
        return True

    async def scan(self, **kwargs):
        return []

    async def connect(self, device):
        pass

    async def write(self, data: bytes):
        self._written.append(data)

    async def read_notify(self, timeout: float = 5.0) -> bytes:
        return await asyncio.wait_for(self._notify_queue.get(), timeout=timeout)

    async def disconnect(self):
        pass

    def inject_response(self, cmd_name: str, resp_data: bytes, transaction_id: int):
        """Build and enqueue a full response (command → containers)."""
        cmd = CommandPacket(
            cmd_type=CommandType.RESPONSE,
            cmd_name=cmd_name,
            data=resp_data,
        )
        payload = cmd.serialize()
        splitter = ContainerSplitter(mtu=self._mtu)
        containers = splitter.split(payload, transaction_id=transaction_id)
        for c in containers:
            self._notify_queue.put_nowait(c.serialize())


def make_client(transport: MockTransport) -> BlerpcClient:
    """Create a BlerpcClient wired to a mock transport."""
    client = BlerpcClient()
    client._transport = transport
    client._splitter = ContainerSplitter(mtu=transport.mtu)
    client._timeout_s = 2.0
    return client


# ── Echo tests ───────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_echo_roundtrip():
    """Test full echo protocol: encode request → decode response."""
    transport = MockTransport()
    client = make_client(transport)

    # Pre-enqueue echo response
    resp = blerpc_pb2.EchoResponse(message="hello")
    transport.inject_response("echo", resp.SerializeToString(), transaction_id=0)

    result = await client.echo(message="hello")
    assert result.message == "hello"

    # Verify the request was correctly encoded
    assert len(transport._written) == 1
    container = Container.deserialize(transport._written[0])
    assert container.container_type == ContainerType.FIRST
    cmd = CommandPacket.deserialize(container.payload[: container.total_length])
    assert cmd.cmd_type == CommandType.REQUEST
    assert cmd.cmd_name == "echo"

    req = blerpc_pb2.EchoRequest()
    req.ParseFromString(cmd.data)
    assert req.message == "hello"


@pytest.mark.asyncio
async def test_echo_empty():
    transport = MockTransport()
    client = make_client(transport)
    resp = blerpc_pb2.EchoResponse(message="")
    transport.inject_response("echo", resp.SerializeToString(), transaction_id=0)
    result = await client.echo(message="")
    assert result.message == ""


@pytest.mark.asyncio
async def test_echo_max_length():
    transport = MockTransport()
    client = make_client(transport)
    msg = "A" * 256
    resp = blerpc_pb2.EchoResponse(message=msg)
    transport.inject_response("echo", resp.SerializeToString(), transaction_id=0)
    result = await client.echo(message=msg)
    assert result.message == msg


# ── Flash read tests ─────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_flash_read_roundtrip():
    transport = MockTransport()
    client = make_client(transport)
    data = bytes(range(256)) * 4  # 1024 bytes
    resp = blerpc_pb2.FlashReadResponse(address=0x1000, data=data)
    transport.inject_response("flash_read", resp.SerializeToString(), transaction_id=0)
    result = await client.flash_read(address=0x1000, length=1024)
    assert result.data == data


@pytest.mark.asyncio
async def test_flash_read_large():
    """8KB response requires many containers."""
    transport = MockTransport()
    client = make_client(transport)
    data = bytes([0xAB] * 8192)
    resp = blerpc_pb2.FlashReadResponse(address=0, data=data)
    transport.inject_response("flash_read", resp.SerializeToString(), transaction_id=0)
    result = await client.flash_read(address=0, length=8192)
    assert len(result.data) == 8192
    assert result.data == data


# ── Multi-container tests ────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_multi_container_request():
    """Request that spans multiple containers."""
    transport = MockTransport(mtu=50)  # Small MTU forces many containers
    client = make_client(transport)
    msg = "X" * 200
    resp = blerpc_pb2.EchoResponse(message=msg)
    transport.inject_response("echo", resp.SerializeToString(), transaction_id=0)
    result = await client.echo(message=msg)
    assert result.message == msg
    # Multiple write calls due to small MTU
    assert len(transport._written) > 1


@pytest.mark.asyncio
async def test_multi_container_response():
    """Response that spans multiple containers."""
    transport = MockTransport(mtu=50)
    client = make_client(transport)
    data = bytes(range(256))
    resp = blerpc_pb2.FlashReadResponse(address=0, data=data)
    transport.inject_response("flash_read", resp.SerializeToString(), transaction_id=0)
    result = await client.flash_read(address=0, length=256)
    assert result.data == data


# ── Error handling tests ─────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_response_timeout():
    """No response → TimeoutError."""
    transport = MockTransport()
    client = make_client(transport)
    client._timeout_s = 0.1
    with pytest.raises(asyncio.TimeoutError):
        await client.echo(message="hello")


@pytest.mark.asyncio
async def test_command_name_mismatch():
    """Response with wrong command name raises RuntimeError."""
    transport = MockTransport()
    client = make_client(transport)
    resp = blerpc_pb2.EchoResponse(message="hello")
    transport.inject_response("wrong_cmd", resp.SerializeToString(), transaction_id=0)
    with pytest.raises(RuntimeError, match="Command name mismatch"):
        await client.echo(message="hello")


@pytest.mark.asyncio
async def test_control_containers_skipped():
    """Control containers in response stream are skipped."""
    transport = MockTransport()
    client = make_client(transport)

    # Enqueue a control container before the actual response
    ctrl = Container(
        transaction_id=0,
        sequence_number=0,
        container_type=ContainerType.CONTROL,
        control_cmd=ControlCmd.TIMEOUT,
        payload=b"\x64\x00",
    )
    transport._notify_queue.put_nowait(ctrl.serialize())

    resp = blerpc_pb2.EchoResponse(message="hello")
    transport.inject_response("echo", resp.SerializeToString(), transaction_id=0)

    result = await client.echo(message="hello")
    assert result.message == "hello"


# ── Transaction ID tests ────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_sequential_calls_increment_transaction_id():
    """Each call uses a different transaction ID."""
    transport = MockTransport()
    client = make_client(transport)

    for i in range(3):
        resp = blerpc_pb2.EchoResponse(message=f"msg{i}")
        transport.inject_response("echo", resp.SerializeToString(), transaction_id=i)
        result = await client.echo(message=f"msg{i}")
        assert result.message == f"msg{i}"

    assert len(transport._written) == 3
    tids = set()
    for raw in transport._written:
        c = Container.deserialize(raw)
        tids.add(c.transaction_id)
    assert len(tids) == 3


# ── Payload size limit tests ─────────────────────────────────────────────


@pytest.mark.asyncio
async def test_payload_too_large_raises():
    """Request exceeding max_request_payload_size raises PayloadTooLargeError."""
    transport = MockTransport()
    client = make_client(transport)
    client._max_request_payload_size = 50

    with pytest.raises(PayloadTooLargeError):
        await client.echo(message="A" * 256)


@pytest.mark.asyncio
async def test_no_max_payload_allows_large():
    """Without max_request_payload_size, large payloads are allowed."""
    transport = MockTransport()
    client = make_client(transport)
    client._max_request_payload_size = None

    msg = "A" * 256
    resp = blerpc_pb2.EchoResponse(message=msg)
    transport.inject_response("echo", resp.SerializeToString(), transaction_id=0)
    result = await client.echo(message=msg)
    assert result.message == msg


# ── Response too large error tests ────────────────────────────────────────


@pytest.mark.asyncio
async def test_response_too_large_error():
    """ERROR control container with RESPONSE_TOO_LARGE raises ResponseTooLargeError."""
    transport = MockTransport()
    client = make_client(transport)

    # Enqueue an ERROR control container
    err_container = Container(
        transaction_id=0,
        sequence_number=0,
        container_type=ContainerType.CONTROL,
        control_cmd=ControlCmd.ERROR,
        payload=bytes([BLERPC_ERROR_RESPONSE_TOO_LARGE]),
    )
    transport._notify_queue.put_nowait(err_container.serialize())

    with pytest.raises(ResponseTooLargeError):
        await client.echo(message="hello")


@pytest.mark.asyncio
async def test_unknown_error_code_raises_runtime_error():
    """ERROR control container with unknown code raises RuntimeError."""
    transport = MockTransport()
    client = make_client(transport)

    err_container = Container(
        transaction_id=0,
        sequence_number=0,
        container_type=ContainerType.CONTROL,
        control_cmd=ControlCmd.ERROR,
        payload=bytes([0xFF]),
    )
    transport._notify_queue.put_nowait(err_container.serialize())

    with pytest.raises(RuntimeError, match="Peripheral error: 0xff"):
        await client.echo(message="hello")


# ── Stream tests ──────────────────────────────────────────────────────────


def inject_stream_end_p2c(transport: MockTransport, transaction_id: int):
    """Enqueue a STREAM_END_P2C control container."""
    ctrl = Container(
        transaction_id=transaction_id,
        sequence_number=0,
        container_type=ContainerType.CONTROL,
        control_cmd=ControlCmd.STREAM_END_P2C,
        payload=b"",
    )
    transport._notify_queue.put_nowait(ctrl.serialize())


@pytest.mark.asyncio
async def test_counter_stream():
    """Test P→C stream: receive N counter responses + STREAM_END_P2C."""
    transport = MockTransport()
    client = make_client(transport)

    count = 5
    # Pre-enqueue N responses followed by STREAM_END_P2C
    for i in range(count):
        resp = blerpc_pb2.CounterStreamResponse(seq=i, value=i * 10)
        transport.inject_response(
            "counter_stream", resp.SerializeToString(), transaction_id=i + 10
        )
    inject_stream_end_p2c(transport, transaction_id=100)

    results = await client.counter_stream(count)

    assert len(results) == count
    for i, (seq, value) in enumerate(results):
        assert seq == i
        assert value == i * 10


@pytest.mark.asyncio
async def test_counter_stream_empty():
    """Test P→C stream with count=0: only STREAM_END_P2C."""
    transport = MockTransport()
    client = make_client(transport)

    inject_stream_end_p2c(transport, transaction_id=100)

    results = await client.counter_stream(0)
    assert results == []


@pytest.mark.asyncio
async def test_counter_upload():
    """Test C->P stream: send N requests, STREAM_END_C2P, get response."""
    transport = MockTransport()
    client = make_client(transport)

    count = 3
    # Pre-enqueue final response
    resp = blerpc_pb2.CounterUploadResponse(received_count=count)
    transport.inject_response(
        "counter_upload", resp.SerializeToString(), transaction_id=50
    )

    result = await client.counter_upload(count)
    assert result == count

    # Verify N requests + STREAM_END_C2P were sent
    # N messages each produce container(s), plus one STREAM_END_C2P control container
    written_containers = [Container.deserialize(w) for w in transport._written]
    stream_ends = [
        c
        for c in written_containers
        if c.container_type == ContainerType.CONTROL
        and c.control_cmd == ControlCmd.STREAM_END_C2P
    ]
    assert len(stream_ends) == 1

    # All non-control containers should be FIRST containers (requests are small)
    data_containers = [
        c for c in written_containers if c.container_type != ContainerType.CONTROL
    ]
    assert len(data_containers) == count


@pytest.mark.asyncio
async def test_stream_receive_error_during_stream():
    """ERROR control during P→C stream raises."""
    transport = MockTransport()
    client = make_client(transport)

    # Enqueue one response, then an error
    resp = blerpc_pb2.CounterStreamResponse(seq=0, value=0)
    transport.inject_response(
        "counter_stream", resp.SerializeToString(), transaction_id=10
    )
    err_container = Container(
        transaction_id=0,
        sequence_number=0,
        container_type=ContainerType.CONTROL,
        control_cmd=ControlCmd.ERROR,
        payload=bytes([BLERPC_ERROR_RESPONSE_TOO_LARGE]),
    )
    transport._notify_queue.put_nowait(err_container.serialize())

    with pytest.raises(ResponseTooLargeError):
        async for _ in client.stream_receive(
            "counter_stream",
            blerpc_pb2.CounterStreamRequest(count=5).SerializeToString(),
        ):
            pass
