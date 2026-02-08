"""Tests for the command encode/decode layer."""

import struct

import pytest
from blerpc_protocol.command import CommandPacket, CommandType


class TestCommandSerialize:
    def test_serialize_request(self):
        cmd = CommandPacket(
            cmd_type=CommandType.REQUEST, cmd_name="echo", data=b"\x01\x02"
        )
        raw = cmd.serialize()
        # Byte 0: type=0 in MSB => 0x00
        assert raw[0] == 0x00
        # Byte 1: cmd_name_len = 4
        assert raw[1] == 4
        # Bytes 2-5: "echo"
        assert raw[2:6] == b"echo"
        # Bytes 6-7: data_len = 2 (little-endian)
        assert struct.unpack_from("<H", raw, 6)[0] == 2
        # Bytes 8-9: data
        assert raw[8:10] == b"\x01\x02"

    def test_serialize_response(self):
        cmd = CommandPacket(
            cmd_type=CommandType.RESPONSE, cmd_name="echo", data=b"\x03"
        )
        raw = cmd.serialize()
        # Byte 0: type=1 in MSB => 0x80
        assert raw[0] == 0x80

    def test_roundtrip(self):
        original = CommandPacket(
            cmd_type=CommandType.REQUEST, cmd_name="flash_read", data=b"\xaa\xbb\xcc"
        )
        raw = original.serialize()
        decoded = CommandPacket.deserialize(raw)
        assert decoded.cmd_type == CommandType.REQUEST
        assert decoded.cmd_name == "flash_read"
        assert decoded.data == b"\xaa\xbb\xcc"

    def test_roundtrip_response(self):
        original = CommandPacket(
            cmd_type=CommandType.RESPONSE, cmd_name="echo", data=b"hello"
        )
        raw = original.serialize()
        decoded = CommandPacket.deserialize(raw)
        assert decoded.cmd_type == CommandType.RESPONSE
        assert decoded.cmd_name == "echo"
        assert decoded.data == b"hello"

    def test_ascii_cmd_name(self):
        cmd = CommandPacket(
            cmd_type=CommandType.REQUEST, cmd_name="test_cmd_123", data=b""
        )
        raw = cmd.serialize()
        decoded = CommandPacket.deserialize(raw)
        assert decoded.cmd_name == "test_cmd_123"

    def test_empty_data(self):
        cmd = CommandPacket(cmd_type=CommandType.REQUEST, cmd_name="ping", data=b"")
        raw = cmd.serialize()
        decoded = CommandPacket.deserialize(raw)
        assert decoded.data == b""
        # data_len should be 0
        name_len = raw[1]
        data_len = struct.unpack_from("<H", raw, 2 + name_len)[0]
        assert data_len == 0

    def test_data_len_little_endian(self):
        """Verify data_len is encoded as little-endian uint16."""
        data = b"\x00" * 300
        cmd = CommandPacket(cmd_type=CommandType.REQUEST, cmd_name="x", data=data)
        raw = cmd.serialize()
        # cmd_name_len=1, cmd_name="x"(1 byte), data_len at offset 3
        data_len_bytes = raw[3:5]
        assert data_len_bytes == struct.pack("<H", 300)  # 0x2C, 0x01

    def test_deserialize_too_short(self):
        with pytest.raises(ValueError):
            CommandPacket.deserialize(b"\x00")
