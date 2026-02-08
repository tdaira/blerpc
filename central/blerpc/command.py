"""Command encode/decode layer for blerpc.

Command format (bits):
| type(1) | reserved(7) | cmd_name_len(8) | cmd_name(N*8) |
| data_len(16) | data(data_len*8) |

- type: 0=request, 1=response
- cmd_name: ASCII command name
- data_len: little-endian uint16
- data: protobuf-encoded bytes
"""

from __future__ import annotations

import struct
from dataclasses import dataclass
from enum import IntEnum


class CommandType(IntEnum):
    REQUEST = 0
    RESPONSE = 1


@dataclass
class CommandPacket:
    """A single command packet."""

    cmd_type: CommandType
    cmd_name: str
    data: bytes = b""

    def serialize(self) -> bytes:
        """Serialize command to bytes."""
        name_bytes = self.cmd_name.encode("ascii")
        if len(name_bytes) > 255:
            raise ValueError(f"cmd_name too long: {len(name_bytes)} > 255")
        if len(self.data) > 65535:
            raise ValueError(f"data too long: {len(self.data)} > 65535")

        # Byte 0: type in MSB (bit 7), reserved bits 6-0 = 0
        byte0 = (self.cmd_type & 0x01) << 7
        return (
            bytes([byte0])
            + struct.pack("<B", len(name_bytes))
            + name_bytes
            + struct.pack("<H", len(self.data))
            + self.data
        )

    @staticmethod
    def deserialize(data: bytes) -> CommandPacket:
        """Deserialize bytes into a CommandPacket."""
        if len(data) < 2:
            raise ValueError(f"Command packet too short: {len(data)} bytes")

        # Byte 0: type in MSB
        cmd_type = CommandType((data[0] >> 7) & 0x01)
        cmd_name_len = data[1]

        offset = 2
        if len(data) < offset + cmd_name_len + 2:
            raise ValueError("Command packet truncated")

        cmd_name = data[offset : offset + cmd_name_len].decode("ascii")
        offset += cmd_name_len

        data_len = struct.unpack_from("<H", data, offset)[0]
        offset += 2

        payload = data[offset : offset + data_len]
        return CommandPacket(
            cmd_type=cmd_type,
            cmd_name=cmd_name,
            data=payload,
        )
