"""Integration tests requiring real BLE hardware.

Run with: pytest tests/test_integration.py -v -s
Requires an nRF54L15 DK running the blerpc peripheral firmware.
"""

import time

import pytest
import pytest_asyncio
from blerpc.client import BlerpcClient, PayloadTooLargeError, ResponseTooLargeError

# Skip all tests if no BLE hardware is available
pytestmark = pytest.mark.skipif(
    not pytest.importorskip("bleak"),
    reason="bleak not available",
)


@pytest_asyncio.fixture
async def client():
    c = BlerpcClient()
    try:
        devices = await c.scan(timeout=15.0)
        if not devices:
            pytest.skip("No blerpc peripheral found")
        await c.connect(devices[0])
        c._timeout_s = 5.0  # Use generous timeout for integration tests
        yield c
    except Exception:
        pytest.skip("Could not connect to blerpc peripheral")
    finally:
        try:
            await c.disconnect()
        except Exception:
            pass  # Ignore disconnect errors


@pytest.mark.asyncio
async def test_capabilities(client):
    """Verify capabilities were received during connect()."""
    print(
        f"\nCapabilities: max_request={client.max_request_payload_size}, "
        f"max_response={client.max_response_payload_size}, "
        f"encrypted={client.is_encrypted}"
    )
    assert client.max_request_payload_size is not None
    assert client.max_response_payload_size is not None
    assert client.max_request_payload_size > 0
    assert client.max_response_payload_size > 0
    assert client.is_encrypted, "E2E encryption should be established"


@pytest.mark.asyncio
async def test_payload_too_large(client):
    """Request exceeding max_request_payload_size raises PayloadTooLargeError."""
    if client.max_request_payload_size is None:
        pytest.skip("Peripheral did not report capabilities")
    limit = client.max_request_payload_size
    # Build a message large enough to exceed the limit after protobuf+command encoding
    with pytest.raises(PayloadTooLargeError):
        await client.echo(message="A" * limit)


@pytest.mark.asyncio
async def test_echo_basic(client):
    result = await client.echo(message="hello")
    assert result.message == "hello"


@pytest.mark.asyncio
async def test_echo_empty(client):
    result = await client.echo(message="")
    assert result.message == ""


@pytest.mark.asyncio
async def test_echo_max_length(client):
    message = "A" * 256
    result = await client.echo(message=message)
    assert result.message == message


@pytest.mark.asyncio
async def test_flash_read_basic(client):
    result = await client.flash_read(address=0x00000000, length=16)
    assert len(result.data) == 16


@pytest.mark.asyncio
async def test_flash_read_8kb(client):
    """Test reading 8KB in a single call."""
    result = await client.flash_read(address=0x00000000, length=8192)
    assert len(result.data) == 8192


@pytest.mark.asyncio
async def test_flash_read_throughput(client):
    """Continuous reads to measure sustained throughput."""
    read_size = 8192
    num_reads = 10
    total_bytes = read_size * num_reads

    # Warm up
    await client.flash_read(address=0x00000000, length=read_size)

    start = time.monotonic()
    for i in range(num_reads):
        result = await client.flash_read(address=0x00000000, length=read_size)
        assert len(result.data) == read_size
    elapsed = time.monotonic() - start

    kb_per_sec = total_bytes / 1024.0 / elapsed
    per_call = elapsed / num_reads
    print(
        f"\n[BENCH] flash_read_throughput: {kb_per_sec:.1f} KB/s "
        f"({total_bytes}B in {elapsed * 1000:.0f}ms, "
        f"{per_call * 1000:.1f}ms/call)"
    )


@pytest.mark.asyncio
async def test_flash_read_overhead(client):
    """Measure per-call overhead with minimal payload (1 byte × 20 calls)."""
    count = 20

    # Warm up
    await client.flash_read(address=0x00000000, length=1)

    start = time.monotonic()
    for _ in range(count):
        await client.flash_read(address=0x00000000, length=1)
    elapsed = time.monotonic() - start

    ms_per_call = elapsed * 1000 / count
    print(
        f"\n[BENCH] flash_read_overhead: {ms_per_call:.1f} ms/call "
        f"(1 byte x {count} calls in {elapsed * 1000:.0f} ms)"
    )


@pytest.mark.asyncio
async def test_echo_roundtrip(client):
    """Measure echo round-trip latency (50 calls)."""
    count = 50

    # Warm up
    await client.echo(message="x")

    start = time.monotonic()
    for _ in range(count):
        await client.echo(message="hello")
    elapsed = time.monotonic() - start

    ms_per_call = elapsed * 1000 / count
    print(
        f"\n[BENCH] echo_roundtrip: {ms_per_call:.1f} ms/call "
        f"({count} calls in {elapsed * 1000:.0f} ms)"
    )


@pytest.mark.asyncio
async def test_data_write_basic(client):
    data = bytes(range(256)) * 4  # 1024 bytes
    result = await client.data_write(data=data)
    assert result.length == len(data)


@pytest.mark.asyncio
async def test_data_write_8kb(client):
    """Test writing 8KB in a single call."""
    data = bytes(range(256)) * 32  # 8192 bytes
    result = await client.data_write(data=data)
    assert result.length == 8192


@pytest.mark.asyncio
async def test_data_write_throughput(client):
    """Continuous writes to measure sustained upload throughput."""
    write_size = 200
    num_writes = 20
    total_bytes = write_size * num_writes
    data = bytes(i % 256 for i in range(write_size))

    # Warm up
    await client.data_write(data=data)

    start = time.monotonic()
    for i in range(num_writes):
        await client.data_write(data=data)
    elapsed = time.monotonic() - start

    kb_per_sec = total_bytes / 1024.0 / elapsed
    per_call = elapsed / num_writes
    print(
        f"\n[BENCH] data_write_throughput: {kb_per_sec:.1f} KB/s "
        f"({total_bytes}B in {elapsed * 1000:.0f}ms, "
        f"{per_call * 1000:.1f}ms/call)"
    )


@pytest.mark.asyncio
async def test_response_too_large(client):
    """flash_read(0, 128) should trigger ResponseTooLargeError."""
    if client.max_response_payload_size is None:
        pytest.skip("Peripheral did not report capabilities")
    if client.max_response_payload_size > 100:
        pytest.skip(
            f"Peripheral max_response={client.max_response_payload_size}, "
            "need <=100 for this test"
        )
    print(f"\nmax_response_payload_size={client.max_response_payload_size}")
    with pytest.raises(ResponseTooLargeError):
        await client.flash_read(address=0, length=128)


@pytest.mark.asyncio
async def test_echo_within_limit_with_small_max(client):
    """Short echo should succeed even when MAX_RESPONSE_PAYLOAD_SIZE=100."""
    if client.max_response_payload_size is None:
        pytest.skip("Peripheral did not report capabilities")
    if client.max_response_payload_size > 100:
        pytest.skip(
            f"Peripheral max_response={client.max_response_payload_size}, "
            "need <=100 for this test"
        )
    result = await client.echo(message="hello")
    assert result.message == "hello"


@pytest.mark.asyncio
async def test_multi_container_echo(client):
    """Test echo with a message that requires multi-container transport."""
    message = "B" * 250
    result = await client.echo(message=message)
    assert result.message == message


@pytest.mark.asyncio
async def test_counter_stream(client):
    """P→C stream: receive N counter values."""
    count = 5
    results = await client.counter_stream(count)
    print(f"\nCounterStream: received {len(results)} responses")
    assert len(results) == count
    for i, (seq, value) in enumerate(results):
        assert seq == i, f"seq mismatch at {i}: expected {i}, got {seq}"
        assert value == i * 10, f"value mismatch at {i}: expected {i * 10}, got {value}"


@pytest.mark.asyncio
async def test_counter_stream_large(client):
    """P→C stream: receive 20 counter values."""
    count = 20
    results = await client.counter_stream(count)
    assert len(results) == count
    for i, (seq, value) in enumerate(results):
        assert seq == i
        assert value == i * 10


@pytest.mark.asyncio
async def test_counter_upload(client):
    """C→P stream: upload N counter values."""
    count = 5
    received = await client.counter_upload(count)
    print(f"\nCounterUpload: received_count={received}")
    assert received == count


@pytest.mark.asyncio
async def test_counter_upload_large(client):
    """C→P stream: upload 20 counter values."""
    count = 20
    received = await client.counter_upload(count)
    assert received == count


@pytest.mark.asyncio
async def test_stream_throughput(client):
    """Measure stream throughput for counter_stream (P→C) and counter_upload (C→P)."""
    count = 20

    # counter_stream (P→C): peripheral sends 'count' responses
    start = time.monotonic()
    results = await client.counter_stream(count)
    elapsed = time.monotonic() - start
    assert len(results) == count
    ms_per_item = elapsed * 1000 / count
    print(
        f"\n[BENCH] counter_stream (P->C): {count} items in {elapsed * 1000:.0f} ms "
        f"({ms_per_item:.1f} ms/item)"
    )

    # counter_upload (C→P): central sends 'count' requests
    start = time.monotonic()
    received = await client.counter_upload(count)
    elapsed = time.monotonic() - start
    assert received == count
    ms_per_item = elapsed * 1000 / count
    print(
        f"[BENCH] counter_upload (C->P): {count} items in {elapsed * 1000:.0f} ms "
        f"({ms_per_item:.1f} ms/item)"
    )
