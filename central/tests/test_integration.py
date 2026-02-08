"""Integration tests requiring real BLE hardware.

Run with: pytest tests/test_integration.py -v -s
Requires an nRF54L15 DK running the blerpc peripheral firmware.
"""

import time

import pytest
import pytest_asyncio
from blerpc.client import BlerpcClient

# Skip all tests if no BLE hardware is available
pytestmark = pytest.mark.skipif(
    not pytest.importorskip("bleak"),
    reason="bleak not available",
)


@pytest_asyncio.fixture
async def client():
    c = BlerpcClient()
    try:
        await c.connect(timeout=15.0)
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
async def test_echo_basic(client):
    result = await client.echo("hello")
    assert result == "hello"


@pytest.mark.asyncio
async def test_echo_empty(client):
    result = await client.echo("")
    assert result == ""


@pytest.mark.asyncio
async def test_echo_max_length(client):
    message = "A" * 256
    result = await client.echo(message)
    assert result == message


@pytest.mark.asyncio
async def test_flash_read_basic(client):
    data = await client.flash_read(0x00000000, 16)
    assert len(data) == 16


@pytest.mark.asyncio
async def test_flash_read_8kb(client):
    """Test reading 8KB in a single call."""
    data = await client.flash_read(0x00000000, 8192)
    assert len(data) == 8192


@pytest.mark.asyncio
async def test_flash_read_throughput(client):
    """Continuous reads to measure sustained throughput."""
    read_size = 8192
    num_reads = 10
    total_bytes = read_size * num_reads

    # Warm up
    await client.flash_read(0x00000000, read_size)

    start = time.monotonic()
    for i in range(num_reads):
        data = await client.flash_read(0x00000000, read_size)
        assert len(data) == read_size
    elapsed = time.monotonic() - start

    throughput = total_bytes / elapsed
    per_call = elapsed / num_reads
    print(
        f"\nThroughput: {num_reads}x {read_size}B = {total_bytes}B "
        f"in {elapsed:.3f}s = {throughput:.0f} bytes/s ({per_call * 1000:.1f}ms/call)"
    )


@pytest.mark.asyncio
async def test_flash_read_overhead(client):
    """Compare 1x8KB vs 8x1KB to measure per-call overhead."""
    # Single 8KB read
    await client.flash_read(0x00000000, 8192)  # warm up
    start = time.monotonic()
    for _ in range(5):
        await client.flash_read(0x00000000, 8192)
    time_8kb = (time.monotonic() - start) / 5

    # 8x1KB reads
    await client.flash_read(0x00000000, 1024)  # warm up
    start = time.monotonic()
    for _ in range(5):
        for _ in range(8):
            await client.flash_read(0x00000000, 1024)
    time_8x1kb = (time.monotonic() - start) / 5

    overhead = time_8x1kb - time_8kb
    print(f"\n1x8KB:  {time_8kb * 1000:.1f}ms ({8192 / time_8kb:.0f} bytes/s)")
    print(f"8x1KB:  {time_8x1kb * 1000:.1f}ms ({8192 / time_8x1kb:.0f} bytes/s)")
    print(
        f"Overhead of 7 extra calls: {overhead * 1000:.1f}ms "
        f"({overhead / 7 * 1000:.1f}ms/call)"
    )


@pytest.mark.asyncio
async def test_data_write_basic(client):
    data = bytes(range(256)) * 4  # 1024 bytes
    confirmed = await client.data_write(data)
    assert confirmed == len(data)


@pytest.mark.asyncio
async def test_data_write_8kb(client):
    """Test writing 8KB in a single call."""
    data = bytes(range(256)) * 32  # 8192 bytes
    confirmed = await client.data_write(data)
    assert confirmed == 8192


@pytest.mark.asyncio
async def test_data_write_throughput(client):
    """Continuous writes to measure sustained upload throughput."""
    write_size = 8192
    num_writes = 10
    total_bytes = write_size * num_writes
    data = bytes(range(256)) * 32  # 8192 bytes

    # Warm up
    await client.data_write(data)

    start = time.monotonic()
    for i in range(num_writes):
        confirmed = await client.data_write(data)
        assert confirmed == write_size
    elapsed = time.monotonic() - start

    throughput = total_bytes / elapsed
    per_call = elapsed / num_writes
    print(
        f"\nWrite throughput: {num_writes}x {write_size}B = {total_bytes}B "
        f"in {elapsed:.3f}s = {throughput:.0f} bytes/s ({per_call * 1000:.1f}ms/call)"
    )


@pytest.mark.asyncio
async def test_multi_container_echo(client):
    """Test echo with a message that requires multi-container transport."""
    message = "B" * 250
    result = await client.echo(message)
    assert result == message
