"""Tests for the container split/merge/control layer."""

import struct

import pytest
from blerpc_protocol.container import (
    ATT_OVERHEAD,
    FIRST_HEADER_SIZE,
    SUBSEQUENT_HEADER_SIZE,
    Container,
    ContainerAssembler,
    ContainerSplitter,
    ContainerType,
    ControlCmd,
    make_capabilities_request,
    make_capabilities_response,
    make_stream_end_c2p,
    make_stream_end_p2c,
    make_timeout_request,
    make_timeout_response,
)

# --- serialize / deserialize roundtrip ---


class TestContainerSerializeDeserialize:
    def test_first_container_roundtrip(self):
        c = Container(
            transaction_id=42,
            sequence_number=0,
            container_type=ContainerType.FIRST,
            total_length=100,
            payload=b"\x01\x02\x03",
        )
        data = c.serialize()
        c2 = Container.deserialize(data)
        assert c2.transaction_id == 42
        assert c2.sequence_number == 0
        assert c2.container_type == ContainerType.FIRST
        assert c2.total_length == 100
        assert c2.payload == b"\x01\x02\x03"

    def test_subsequent_container_roundtrip(self):
        c = Container(
            transaction_id=7,
            sequence_number=3,
            container_type=ContainerType.SUBSEQUENT,
            payload=b"\xaa\xbb",
        )
        data = c.serialize()
        c2 = Container.deserialize(data)
        assert c2.transaction_id == 7
        assert c2.sequence_number == 3
        assert c2.container_type == ContainerType.SUBSEQUENT
        assert c2.payload == b"\xaa\xbb"

    def test_control_container_roundtrip(self):
        c = Container(
            transaction_id=1,
            sequence_number=0,
            container_type=ContainerType.CONTROL,
            control_cmd=ControlCmd.TIMEOUT,
            payload=struct.pack("<H", 500),
        )
        data = c.serialize()
        c2 = Container.deserialize(data)
        assert c2.container_type == ContainerType.CONTROL
        assert c2.control_cmd == ControlCmd.TIMEOUT
        assert struct.unpack("<H", c2.payload)[0] == 500

    def test_flags_byte_encoding(self):
        """Verify byte 2 bitfield: type(7-6) | control_cmd(5-2) | reserved(1-0)."""
        c = Container(
            transaction_id=0,
            sequence_number=0,
            container_type=ContainerType.CONTROL,  # 0b11
            control_cmd=ControlCmd.STREAM_END_C2P,  # 0x2 = 0b0010
            payload=b"",
        )
        data = c.serialize()
        flags = data[2]
        # type=0b11 in bits 7-6 => 0b11000000 = 0xC0
        # control_cmd=0x2 in bits 5-2 => 0b00001000 = 0x08
        assert flags == 0xC0 | 0x08  # 0xC8

    def test_deserialize_too_short(self):
        with pytest.raises(ValueError):
            Container.deserialize(b"\x00\x01")

    def test_first_container_header_size(self):
        c = Container(
            transaction_id=0,
            sequence_number=0,
            container_type=ContainerType.FIRST,
            total_length=0,
            payload=b"",
        )
        data = c.serialize()
        assert len(data) == FIRST_HEADER_SIZE

    def test_subsequent_container_header_size(self):
        c = Container(
            transaction_id=0,
            sequence_number=0,
            container_type=ContainerType.SUBSEQUENT,
            payload=b"",
        )
        data = c.serialize()
        assert len(data) == SUBSEQUENT_HEADER_SIZE


# --- ContainerSplitter ---


class TestContainerSplitter:
    def test_small_payload_single_container(self):
        splitter = ContainerSplitter(mtu=247)
        payload = b"hello"
        containers = splitter.split(payload, transaction_id=0)
        assert len(containers) == 1
        assert containers[0].container_type == ContainerType.FIRST
        assert containers[0].total_length == 5
        assert containers[0].payload == b"hello"

    def test_large_payload_multiple_containers(self):
        mtu = 27  # Small MTU for testing
        splitter = ContainerSplitter(mtu=mtu)
        effective = mtu - ATT_OVERHEAD  # 24
        first_payload_max = effective - FIRST_HEADER_SIZE  # 24 - 6 = 18
        subsequent_payload_max = effective - SUBSEQUENT_HEADER_SIZE  # 24 - 4 = 20

        payload = bytes(range(256)) * 2  # 512 bytes
        containers = splitter.split(payload, transaction_id=5)

        assert containers[0].container_type == ContainerType.FIRST
        assert containers[0].total_length == 512
        assert len(containers[0].payload) == first_payload_max

        for c in containers[1:]:
            assert c.container_type == ContainerType.SUBSEQUENT
            assert len(c.payload) <= subsequent_payload_max

        # Verify all data is accounted for
        reassembled = b"".join(c.payload for c in containers)
        assert reassembled == payload

    def test_boundary_payload_exactly_first_max(self):
        mtu = 30
        splitter = ContainerSplitter(mtu=mtu)
        effective = mtu - ATT_OVERHEAD  # 27
        first_max = effective - FIRST_HEADER_SIZE  # 21

        payload = b"A" * first_max
        containers = splitter.split(payload, transaction_id=0)
        assert len(containers) == 1
        assert containers[0].payload == payload

    def test_boundary_payload_one_byte_over_first_max(self):
        mtu = 30
        splitter = ContainerSplitter(mtu=mtu)
        effective = mtu - ATT_OVERHEAD
        first_max = effective - FIRST_HEADER_SIZE

        payload = b"A" * (first_max + 1)
        containers = splitter.split(payload, transaction_id=0)
        assert len(containers) == 2
        assert len(containers[0].payload) == first_max
        assert len(containers[1].payload) == 1

    def test_empty_payload(self):
        splitter = ContainerSplitter(mtu=247)
        containers = splitter.split(b"", transaction_id=0)
        assert len(containers) == 1
        assert containers[0].total_length == 0
        assert containers[0].payload == b""

    def test_transaction_id_auto_increment(self):
        splitter = ContainerSplitter(mtu=247)
        c1 = splitter.split(b"a")
        c2 = splitter.split(b"b")
        assert c1[0].transaction_id == 0
        assert c2[0].transaction_id == 1

    def test_transaction_id_wraps_at_256(self):
        splitter = ContainerSplitter(mtu=247)
        splitter._transaction_counter = 255
        c1 = splitter.split(b"a")
        c2 = splitter.split(b"b")
        assert c1[0].transaction_id == 255
        assert c2[0].transaction_id == 0


# --- ContainerAssembler ---


class TestContainerAssembler:
    def test_single_container_assembly(self):
        assembler = ContainerAssembler()
        c = Container(
            transaction_id=0,
            sequence_number=0,
            container_type=ContainerType.FIRST,
            total_length=5,
            payload=b"hello",
        )
        result = assembler.feed(c)
        assert result == b"hello"

    def test_multi_container_assembly(self):
        assembler = ContainerAssembler()
        c1 = Container(
            transaction_id=1,
            sequence_number=0,
            container_type=ContainerType.FIRST,
            total_length=8,
            payload=b"hell",
        )
        c2 = Container(
            transaction_id=1,
            sequence_number=1,
            container_type=ContainerType.SUBSEQUENT,
            payload=b"o wo",
        )
        assert assembler.feed(c1) is None
        result = assembler.feed(c2)
        assert result == b"hello wo"

    def test_sequence_gap_discards_transaction(self):
        assembler = ContainerAssembler()
        c1 = Container(
            transaction_id=2,
            sequence_number=0,
            container_type=ContainerType.FIRST,
            total_length=10,
            payload=b"abc",
        )
        c_bad = Container(
            transaction_id=2,
            sequence_number=2,  # Gap: expected 1
            container_type=ContainerType.SUBSEQUENT,
            payload=b"def",
        )
        assert assembler.feed(c1) is None
        result = assembler.feed(c_bad)
        assert result is None
        # Transaction should be discarded
        assert 2 not in assembler._transactions

    def test_control_container_ignored(self):
        assembler = ContainerAssembler()
        c = make_timeout_request(transaction_id=0)
        result = assembler.feed(c)
        assert result is None

    def test_subsequent_without_first_ignored(self):
        assembler = ContainerAssembler()
        c = Container(
            transaction_id=99,
            sequence_number=1,
            container_type=ContainerType.SUBSEQUENT,
            payload=b"orphan",
        )
        result = assembler.feed(c)
        assert result is None


# --- Split-then-assemble roundtrip ---


class TestSplitAssembleRoundtrip:
    def test_roundtrip_small(self):
        splitter = ContainerSplitter(mtu=247)
        assembler = ContainerAssembler()
        payload = b"hello world"

        containers = splitter.split(payload, transaction_id=0)
        result = None
        for c in containers:
            serialized = c.serialize()
            deserialized = Container.deserialize(serialized)
            result = assembler.feed(deserialized)

        assert result == payload

    def test_roundtrip_large(self):
        splitter = ContainerSplitter(mtu=27)
        assembler = ContainerAssembler()
        payload = bytes(range(256)) * 4  # 1024 bytes

        containers = splitter.split(payload, transaction_id=10)
        result = None
        for c in containers:
            serialized = c.serialize()
            deserialized = Container.deserialize(serialized)
            result = assembler.feed(deserialized)

        assert result == payload

    def test_roundtrip_large_payload(self):
        """Test large payload staying within 256 seq limit."""
        # MTU=247, effective=244, subsequent_payload=240.
        # 255 subsequent + 1 first ≈ 61438 bytes max.
        splitter = ContainerSplitter(mtu=247)
        assembler = ContainerAssembler()
        payload = b"\xab" * 60000

        containers = splitter.split(payload, transaction_id=0)
        assert len(containers) > 200  # Verify it's actually multi-container
        result = None
        for c in containers:
            serialized = c.serialize()
            deserialized = Container.deserialize(serialized)
            result = assembler.feed(deserialized)

        assert result == payload

    def test_payload_too_large_raises(self):
        """Payload requiring > 256 containers should raise ValueError."""
        splitter = ContainerSplitter(mtu=27)  # Small MTU => few bytes per container
        # effective=24, subsequent=20 bytes each. 256*20 ≈ 5120 max
        payload = b"\x00" * 10000
        with pytest.raises(ValueError, match="sequence_number"):
            splitter.split(payload, transaction_id=0)

    def test_roundtrip_empty(self):
        splitter = ContainerSplitter(mtu=247)
        assembler = ContainerAssembler()
        payload = b""

        containers = splitter.split(payload, transaction_id=0)
        result = None
        for c in containers:
            serialized = c.serialize()
            deserialized = Container.deserialize(serialized)
            result = assembler.feed(deserialized)

        assert result == payload


# --- Control container helpers ---


class TestControlContainers:
    def test_timeout_request(self):
        c = make_timeout_request(transaction_id=5)
        assert c.container_type == ContainerType.CONTROL
        assert c.control_cmd == ControlCmd.TIMEOUT
        assert c.payload == b""

    def test_timeout_response(self):
        c = make_timeout_response(transaction_id=5, timeout_ms=200)
        assert c.container_type == ContainerType.CONTROL
        assert c.control_cmd == ControlCmd.TIMEOUT
        assert struct.unpack("<H", c.payload)[0] == 200

    def test_stream_end_c2p(self):
        c = make_stream_end_c2p(transaction_id=3)
        assert c.control_cmd == ControlCmd.STREAM_END_C2P
        data = c.serialize()
        c2 = Container.deserialize(data)
        assert c2.control_cmd == ControlCmd.STREAM_END_C2P

    def test_stream_end_p2c(self):
        c = make_stream_end_p2c(transaction_id=3)
        assert c.control_cmd == ControlCmd.STREAM_END_P2C

    def test_capabilities_request(self):
        c = make_capabilities_request(transaction_id=7)
        assert c.container_type == ContainerType.CONTROL
        assert c.control_cmd == ControlCmd.CAPABILITIES
        assert c.payload == b""

    def test_capabilities_response(self):
        c = make_capabilities_response(
            transaction_id=7,
            max_request_payload_size=256,
            max_response_payload_size=65535,
        )
        assert c.container_type == ContainerType.CONTROL
        assert c.control_cmd == ControlCmd.CAPABILITIES
        assert len(c.payload) == 4
        assert struct.unpack("<HH", c.payload) == (256, 65535)
