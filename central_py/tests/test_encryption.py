"""E2E encryption integration tests.

Central client ↔ simulated encrypted Peripheral.

Tests the full protocol flow without BLE hardware:
- CAPABILITIES negotiation with encryption flag
- 4-step KEY_EXCHANGE handshake (X25519 + Ed25519 + AES-GCM)
- Encrypted echo/flash_read/data_write RPCs
- Encrypted P→C and C→P streams
- TOFU key management
"""

import asyncio
import os
import struct
import tempfile

import pytest
from blerpc.client import BlerpcClient
from blerpc.generated import blerpc_pb2
from blerpc_protocol.command import CommandPacket, CommandType
from blerpc_protocol.container import (
    CAPABILITY_FLAG_ENCRYPTION_SUPPORTED,
    Container,
    ContainerAssembler,
    ContainerSplitter,
    ContainerType,
    ControlCmd,
    make_stream_end_p2c,
)
from blerpc_protocol.crypto import (
    CONFIRM_CENTRAL,
    CONFIRM_PERIPHERAL,
    DIRECTION_C2P,
    DIRECTION_P2C,
    BlerpcCrypto,
)


class MockEncryptedPeripheral:
    """Simulates a peripheral that supports E2E encryption.

    Processes writes from the client and sends responses back via the notify queue.
    """

    def __init__(self, notify_queue: asyncio.Queue, mtu: int = 247):
        self._notify_queue = notify_queue
        self._mtu = mtu
        self._assembler = ContainerAssembler()
        self._splitter = ContainerSplitter(mtu=mtu)

        # Generate peripheral keys
        self._x25519_privkey, self._x25519_pubkey = (
            BlerpcCrypto.generate_x25519_keypair()
        )
        self._ed25519_privkey, self._ed25519_pubkey = (
            BlerpcCrypto.generate_ed25519_keypair()
        )

        # Encryption state
        self._session_key: bytes | None = None
        self._encryption_active = False
        self._tx_counter = 0
        self._rx_counter = 0

        # Stream state
        self._upload_count = 0

        # RPC handlers
        self._handlers = {
            "echo": self._handle_echo,
            "flash_read": self._handle_flash_read,
            "data_write": self._handle_data_write,
            "counter_upload": self._handle_counter_upload,
        }

    @property
    def ed25519_pubkey(self) -> bytes:
        return self._ed25519_pubkey

    async def process_write(self, data: bytes):
        """Process a write from the client (called by MockEncryptedTransport)."""
        container = Container.deserialize(data)

        if container.container_type == ContainerType.CONTROL:
            await self._handle_control(container)
            return

        result = self._assembler.feed(container)
        if result is not None:
            await self._process_request(result, container.transaction_id)

    async def _handle_control(self, container: Container):
        if container.control_cmd == ControlCmd.TIMEOUT:
            resp = Container(
                transaction_id=container.transaction_id,
                sequence_number=0,
                container_type=ContainerType.CONTROL,
                control_cmd=ControlCmd.TIMEOUT,
                payload=struct.pack("<H", 100),
            )
            self._notify_queue.put_nowait(resp.serialize())

        elif container.control_cmd == ControlCmd.CAPABILITIES:
            flags = CAPABILITY_FLAG_ENCRYPTION_SUPPORTED
            resp = Container(
                transaction_id=container.transaction_id,
                sequence_number=0,
                container_type=ContainerType.CONTROL,
                control_cmd=ControlCmd.CAPABILITIES,
                payload=struct.pack("<HHH", 65535, 65535, flags),
            )
            self._notify_queue.put_nowait(resp.serialize())

        elif container.control_cmd == ControlCmd.KEY_EXCHANGE:
            await self._handle_key_exchange(container)

        elif container.control_cmd == ControlCmd.STREAM_END_C2P:
            await self._handle_stream_end_c2p()

    async def _handle_key_exchange(self, container: Container):
        payload = container.payload
        if len(payload) < 1:
            return

        step = payload[0]

        if step == 0x01:
            # Step 1: Central sends X25519 pubkey
            central_x25519_pubkey = BlerpcCrypto.parse_step1_payload(payload)

            # Sign: central_pubkey || peripheral_pubkey
            sign_msg = central_x25519_pubkey + self._x25519_pubkey
            signature = BlerpcCrypto.ed25519_sign(self._ed25519_privkey, sign_msg)

            # Compute shared secret and session key
            shared_secret = BlerpcCrypto.x25519_shared_secret(
                self._x25519_privkey, central_x25519_pubkey
            )
            self._session_key = BlerpcCrypto.derive_session_key(
                shared_secret, central_x25519_pubkey, self._x25519_pubkey
            )

            # Build step 2 response
            step2 = BlerpcCrypto.build_step2_payload(
                self._x25519_pubkey, signature, self._ed25519_pubkey
            )
            resp = Container(
                transaction_id=container.transaction_id,
                sequence_number=0,
                container_type=ContainerType.CONTROL,
                control_cmd=ControlCmd.KEY_EXCHANGE,
                payload=step2,
            )
            self._notify_queue.put_nowait(resp.serialize())

        elif step == 0x03:
            # Step 3: Central sends encrypted confirmation
            encrypted = BlerpcCrypto.parse_step3_payload(payload)
            plaintext = BlerpcCrypto.decrypt_confirmation(self._session_key, encrypted)
            assert plaintext == CONFIRM_CENTRAL

            # Build step 4 response
            encrypted_confirm = BlerpcCrypto.encrypt_confirmation(
                self._session_key, CONFIRM_PERIPHERAL
            )
            step4 = BlerpcCrypto.build_step4_payload(encrypted_confirm)
            resp = Container(
                transaction_id=container.transaction_id,
                sequence_number=0,
                container_type=ContainerType.CONTROL,
                control_cmd=ControlCmd.KEY_EXCHANGE,
                payload=step4,
            )
            self._notify_queue.put_nowait(resp.serialize())
            self._encryption_active = True
            self._tx_counter = 0
            self._rx_counter = 0

    async def _process_request(self, payload: bytes, transaction_id: int):
        # Decrypt if encryption active
        if self._encryption_active:
            counter, payload = BlerpcCrypto.decrypt_command(
                self._session_key, DIRECTION_C2P, payload
            )
            self._rx_counter = counter

        cmd = CommandPacket.deserialize(payload)
        assert cmd.cmd_type == CommandType.REQUEST

        if cmd.cmd_name == "counter_stream":
            await self._handle_counter_stream(cmd.data)
            return

        handler = self._handlers.get(cmd.cmd_name)
        if not handler:
            return

        resp_data = handler(cmd.data)
        if resp_data is None:
            self._upload_count += 1
            return

        resp_cmd = CommandPacket(
            cmd_type=CommandType.RESPONSE,
            cmd_name=cmd.cmd_name,
            data=resp_data,
        )
        resp_payload = resp_cmd.serialize()

        # Encrypt if active
        if self._encryption_active:
            resp_payload = BlerpcCrypto.encrypt_command(
                self._session_key, self._tx_counter, DIRECTION_P2C, resp_payload
            )
            self._tx_counter += 1

        containers = self._splitter.split(resp_payload, transaction_id=transaction_id)
        for c in containers:
            self._notify_queue.put_nowait(c.serialize())

    async def _handle_counter_stream(self, req_data: bytes):
        req = blerpc_pb2.CounterStreamRequest()
        req.ParseFromString(req_data)

        for i in range(req.count):
            resp = blerpc_pb2.CounterStreamResponse(seq=i, value=i * 10)
            resp_cmd = CommandPacket(
                cmd_type=CommandType.RESPONSE,
                cmd_name="counter_stream",
                data=resp.SerializeToString(),
            )
            resp_payload = resp_cmd.serialize()
            if self._encryption_active:
                resp_payload = BlerpcCrypto.encrypt_command(
                    self._session_key, self._tx_counter, DIRECTION_P2C, resp_payload
                )
                self._tx_counter += 1

            tid = self._splitter.next_transaction_id()
            containers = self._splitter.split(resp_payload, transaction_id=tid)
            for c in containers:
                self._notify_queue.put_nowait(c.serialize())

        tid = self._splitter.next_transaction_id()
        stream_end = make_stream_end_p2c(transaction_id=tid)
        self._notify_queue.put_nowait(stream_end.serialize())

    async def _handle_stream_end_c2p(self):
        count = self._upload_count
        self._upload_count = 0

        resp = blerpc_pb2.CounterUploadResponse(received_count=count)
        resp_cmd = CommandPacket(
            cmd_type=CommandType.RESPONSE,
            cmd_name="counter_upload",
            data=resp.SerializeToString(),
        )
        resp_payload = resp_cmd.serialize()
        if self._encryption_active:
            resp_payload = BlerpcCrypto.encrypt_command(
                self._session_key, self._tx_counter, DIRECTION_P2C, resp_payload
            )
            self._tx_counter += 1

        tid = self._splitter.next_transaction_id()
        containers = self._splitter.split(resp_payload, transaction_id=tid)
        for c in containers:
            self._notify_queue.put_nowait(c.serialize())

    @staticmethod
    def _handle_echo(req_data: bytes) -> bytes:
        req = blerpc_pb2.EchoRequest()
        req.ParseFromString(req_data)
        resp = blerpc_pb2.EchoResponse(message=req.message)
        return resp.SerializeToString()

    @staticmethod
    def _handle_flash_read(req_data: bytes) -> bytes:
        req = blerpc_pb2.FlashReadRequest()
        req.ParseFromString(req_data)
        data = os.urandom(req.length)
        resp = blerpc_pb2.FlashReadResponse(address=req.address, data=data)
        return resp.SerializeToString()

    @staticmethod
    def _handle_data_write(req_data: bytes) -> bytes:
        req = blerpc_pb2.DataWriteRequest()
        req.ParseFromString(req_data)
        return blerpc_pb2.DataWriteResponse(length=len(req.data)).SerializeToString()

    @staticmethod
    def _handle_counter_upload(req_data: bytes) -> bytes | None:
        return None  # Accumulation handled via _upload_count


class MockEncryptedTransport:
    """Mock transport that wires client writes to MockEncryptedPeripheral."""

    def __init__(self, peripheral: MockEncryptedPeripheral, mtu: int = 247):
        self._peripheral = peripheral
        self._mtu = mtu
        self._notify_queue: asyncio.Queue[bytes] = peripheral._notify_queue
        self._address = "AA:BB:CC:DD:EE:FF"

    @property
    def mtu(self) -> int:
        return self._mtu

    @property
    def address(self) -> str:
        return self._address

    @property
    def is_connected(self) -> bool:
        return True

    async def scan(self, **kwargs):
        return []

    async def connect(self, device):
        pass

    async def write(self, data: bytes):
        await self._peripheral.process_write(data)

    async def read_notify(self, timeout: float = 5.0) -> bytes:
        return await asyncio.wait_for(self._notify_queue.get(), timeout=timeout)

    async def disconnect(self):
        pass


def make_encrypted_client(
    mtu: int = 247, known_keys_path: str | None = None
) -> tuple[BlerpcClient, MockEncryptedPeripheral]:
    """Create a BlerpcClient connected to a MockEncryptedPeripheral."""
    notify_queue = asyncio.Queue()
    peripheral = MockEncryptedPeripheral(notify_queue, mtu=mtu)
    transport = MockEncryptedTransport(peripheral, mtu=mtu)
    client = BlerpcClient(known_keys_path=known_keys_path)
    client._transport = transport
    client._splitter = ContainerSplitter(mtu=mtu)
    client._timeout_s = 5.0
    return client, peripheral


# ── Key exchange and encryption establishment ─────────────────────────


@pytest.mark.asyncio
async def test_key_exchange_establishes_encryption():
    """Full CAPABILITIES + KEY_EXCHANGE flow establishes encryption."""
    client, peripheral = make_encrypted_client()

    # Trigger capabilities request which will auto-trigger key exchange
    await client._request_capabilities()

    assert client._session is not None
    assert peripheral._encryption_active is True
    assert peripheral._session_key is not None


@pytest.mark.asyncio
async def test_key_exchange_with_tofu():
    """Key exchange stores and verifies Ed25519 key via TOFU."""
    with tempfile.TemporaryDirectory() as tmpdir:
        keys_path = os.path.join(tmpdir, "known_keys.json")

        # First connection: store key
        client, peripheral = make_encrypted_client(known_keys_path=keys_path)
        await client._request_capabilities()
        assert client._session is not None
        assert os.path.exists(keys_path)

        # Second connection with same peripheral key: should succeed
        client2, peripheral2 = make_encrypted_client(known_keys_path=keys_path)
        # Copy the same Ed25519 key pair to peripheral2
        peripheral2._ed25519_privkey = peripheral._ed25519_privkey
        peripheral2._ed25519_pubkey = peripheral._ed25519_pubkey
        await client2._request_capabilities()
        assert client2._session is not None


@pytest.mark.asyncio
async def test_tofu_rejects_changed_key():
    """TOFU rejects a peripheral whose Ed25519 key has changed."""
    with tempfile.TemporaryDirectory() as tmpdir:
        keys_path = os.path.join(tmpdir, "known_keys.json")

        # First connection: store key
        client, peripheral = make_encrypted_client(known_keys_path=keys_path)
        await client._request_capabilities()
        assert client._session is not None

        # Second connection with DIFFERENT peripheral key: should fail
        client2, peripheral2 = make_encrypted_client(known_keys_path=keys_path)
        # peripheral2 has a different key pair (auto-generated)
        # With require_encryption=True (default), this raises ValueError
        with pytest.raises(ValueError, match="Peripheral key rejected"):
            await client2._request_capabilities()


# ── Encrypted RPC calls ──────────────────────────────────────────────


@pytest.mark.asyncio
async def test_encrypted_echo():
    """Echo RPC works over encrypted channel."""
    client, _ = make_encrypted_client()
    await client._request_capabilities()
    assert client._session is not None

    result = await client.echo(message="hello encrypted")
    assert result.message == "hello encrypted"


@pytest.mark.asyncio
async def test_encrypted_echo_empty():
    """Empty echo over encrypted channel."""
    client, _ = make_encrypted_client()
    await client._request_capabilities()

    result = await client.echo(message="")
    assert result.message == ""


@pytest.mark.asyncio
async def test_encrypted_echo_long():
    """Long echo message that spans multiple containers, all encrypted."""
    client, _ = make_encrypted_client()
    await client._request_capabilities()

    msg = "X" * 500
    result = await client.echo(message=msg)
    assert result.message == msg


@pytest.mark.asyncio
async def test_encrypted_echo_small_mtu():
    """Encrypted echo with small MTU forcing multi-container transport."""
    client, _ = make_encrypted_client(mtu=50)
    await client._request_capabilities()

    msg = "Hello from small MTU encrypted channel"
    result = await client.echo(message=msg)
    assert result.message == msg


@pytest.mark.asyncio
async def test_encrypted_flash_read():
    """FlashRead over encrypted channel returns correct length."""
    client, _ = make_encrypted_client()
    await client._request_capabilities()

    result = await client.flash_read(address=0x1000, length=256)
    assert len(result.data) == 256
    assert result.address == 0x1000


@pytest.mark.asyncio
async def test_encrypted_flash_read_large():
    """Large flash read (8KB) over encrypted channel."""
    client, _ = make_encrypted_client()
    await client._request_capabilities()

    result = await client.flash_read(address=0, length=8192)
    assert len(result.data) == 8192


@pytest.mark.asyncio
async def test_encrypted_data_write():
    """DataWrite over encrypted channel."""
    client, _ = make_encrypted_client()
    await client._request_capabilities()

    data = bytes(range(256)) * 4  # 1024 bytes
    result = await client.data_write(data=data)
    assert result.length == 1024


@pytest.mark.asyncio
async def test_encrypted_data_write_large():
    """Large DataWrite (4KB) over encrypted channel."""
    client, _ = make_encrypted_client()
    await client._request_capabilities()

    data = bytes(range(256)) * 16  # 4096 bytes
    result = await client.data_write(data=data)
    assert result.length == 4096


# ── Encrypted streams ────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_encrypted_counter_stream():
    """P→C stream over encrypted channel."""
    client, _ = make_encrypted_client()
    await client._request_capabilities()

    results = await client.counter_stream(count=5)
    assert len(results) == 5
    for i, (seq, value) in enumerate(results):
        assert seq == i
        assert value == i * 10


@pytest.mark.asyncio
async def test_encrypted_counter_stream_large():
    """P→C stream with 20 items over encrypted channel."""
    client, _ = make_encrypted_client()
    await client._request_capabilities()

    results = await client.counter_stream(count=20)
    assert len(results) == 20


@pytest.mark.asyncio
async def test_encrypted_counter_upload():
    """C→P stream over encrypted channel."""
    client, _ = make_encrypted_client()
    await client._request_capabilities()

    received = await client.counter_upload(count=5)
    assert received == 5


@pytest.mark.asyncio
async def test_encrypted_counter_upload_large():
    """C→P stream with 20 items over encrypted channel."""
    client, _ = make_encrypted_client()
    await client._request_capabilities()

    received = await client.counter_upload(count=20)
    assert received == 20


# ── Counter validation ───────────────────────────────────────────────


@pytest.mark.asyncio
async def test_tx_rx_counters_increment():
    """Verify that tx/rx counters increment correctly after each RPC."""
    client, peripheral = make_encrypted_client()
    await client._request_capabilities()

    for i in range(5):
        await client.echo(message=f"msg{i}")

    # Client sent 5 requests (tx_counter incremented 5 times)
    assert client._session._tx_counter == 5
    # Client received 5 responses (rx_counter is the last counter value = 4)
    assert client._session._rx_counter == 4
    # Peripheral sent 5 responses (tx_counter incremented 5 times)
    assert peripheral._tx_counter == 5
    # Peripheral received 5 requests (rx_counter is the last counter value = 4)
    assert peripheral._rx_counter == 4


# ── Multiple RPCs in sequence ────────────────────────────────────────


@pytest.mark.asyncio
async def test_encrypted_multiple_rpcs_sequence():
    """Multiple different RPCs in sequence over encrypted channel."""
    client, _ = make_encrypted_client()
    await client._request_capabilities()

    # Echo
    result = await client.echo(message="first")
    assert result.message == "first"

    # Flash read
    result = await client.flash_read(address=0, length=64)
    assert len(result.data) == 64

    # Data write
    result = await client.data_write(data=b"\x00" * 128)
    assert result.length == 128

    # Echo again
    result = await client.echo(message="second")
    assert result.message == "second"

    # Stream
    results = await client.counter_stream(count=3)
    assert len(results) == 3

    # Upload
    received = await client.counter_upload(count=3)
    assert received == 3
