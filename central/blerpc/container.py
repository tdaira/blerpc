"""Container split/merge/control layer for blerpc.

Container format (bits):
| transaction_id(8) | sequence_number(8) | type(2)|control_cmd(4)|reserved(2) |
| total_length(16 or 0) | payload_len(8) | payload(variable) |

type=0b00 (FIRST): has total_length, header = 6 bytes
type=0b01 (SUBSEQUENT): no total_length, header = 4 bytes
type=0b11 (CONTROL): no total_length, header = 4 bytes

All multi-byte fields are little-endian.
"""

from __future__ import annotations

import struct
from dataclasses import dataclass, field
from enum import IntEnum


class ContainerType(IntEnum):
    FIRST = 0b00
    SUBSEQUENT = 0b01
    CONTROL = 0b11


class ControlCmd(IntEnum):
    NONE = 0x0
    TIMEOUT = 0x1
    STREAM_END_C2P = 0x2
    STREAM_END_P2C = 0x3


# Header sizes
FIRST_HEADER_SIZE = 6  # txn_id(1) + seq(1) + flags(1) + total_len(2) + payload_len(1)
SUBSEQUENT_HEADER_SIZE = 4  # txn_id(1) + seq(1) + flags(1) + payload_len(1)
CONTROL_HEADER_SIZE = 4  # txn_id(1) + seq(1) + flags(1) + payload_len(1)

ATT_OVERHEAD = 3  # ATT header bytes subtracted from MTU


def _pack_flags(
    container_type: ContainerType, control_cmd: ControlCmd = ControlCmd.NONE
) -> int:
    """Pack type(2) | control_cmd(4) | reserved(2) into a single byte."""
    return ((container_type & 0x03) << 6) | ((control_cmd & 0x0F) << 2)


def _unpack_flags(flags_byte: int) -> tuple[ContainerType, ControlCmd]:
    """Unpack flags byte into (type, control_cmd)."""
    container_type = ContainerType((flags_byte >> 6) & 0x03)
    control_cmd = ControlCmd((flags_byte >> 2) & 0x0F)
    return container_type, control_cmd


@dataclass
class Container:
    """A single container packet."""

    transaction_id: int
    sequence_number: int
    container_type: ContainerType
    control_cmd: ControlCmd = ControlCmd.NONE
    total_length: int = 0  # Only meaningful for FIRST
    payload: bytes = b""

    def serialize(self) -> bytes:
        """Serialize container to bytes."""
        flags = _pack_flags(self.container_type, self.control_cmd)

        if self.container_type == ContainerType.FIRST:
            header = struct.pack(
                "<BBBHB",
                self.transaction_id,
                self.sequence_number,
                flags,
                self.total_length,
                len(self.payload),
            )
        else:
            header = struct.pack(
                "<BBBB",
                self.transaction_id,
                self.sequence_number,
                flags,
                len(self.payload),
            )
        return header + self.payload

    @staticmethod
    def deserialize(data: bytes) -> Container:
        """Deserialize bytes into a Container."""
        if len(data) < 4:
            raise ValueError(f"Container too short: {len(data)} bytes")

        transaction_id = data[0]
        sequence_number = data[1]
        container_type, control_cmd = _unpack_flags(data[2])

        if container_type == ContainerType.FIRST:
            if len(data) < FIRST_HEADER_SIZE:
                raise ValueError(f"FIRST container too short: {len(data)} bytes")
            total_length = struct.unpack_from("<H", data, 3)[0]
            payload_len = data[5]
            payload = data[FIRST_HEADER_SIZE : FIRST_HEADER_SIZE + payload_len]
        else:
            total_length = 0
            payload_len = data[3]
            header_size = SUBSEQUENT_HEADER_SIZE
            payload = data[header_size : header_size + payload_len]

        return Container(
            transaction_id=transaction_id,
            sequence_number=sequence_number,
            container_type=container_type,
            control_cmd=control_cmd,
            total_length=total_length,
            payload=payload,
        )


class ContainerSplitter:
    """Splits a payload into containers respecting MTU."""

    def __init__(self, mtu: int = 247):
        self._mtu = mtu
        self._transaction_counter = 0

    @property
    def effective_mtu(self) -> int:
        """Usable bytes per BLE packet (MTU - ATT overhead)."""
        return self._mtu - ATT_OVERHEAD

    def next_transaction_id(self) -> int:
        tid = self._transaction_counter
        self._transaction_counter = (self._transaction_counter + 1) & 0xFF
        return tid

    def split(
        self, payload: bytes, transaction_id: int | None = None
    ) -> list[Container]:
        """Split payload into a list of containers.

        Raises ValueError if payload is too large for 8-bit
        sequence_number (>255 containers).
        """
        if transaction_id is None:
            transaction_id = self.next_transaction_id()

        total_length = len(payload)
        if total_length > 65535:
            raise ValueError(f"Payload too large: {total_length} > 65535")
        containers: list[Container] = []

        # First container
        first_max_payload = self.effective_mtu - FIRST_HEADER_SIZE
        first_payload = payload[:first_max_payload]
        containers.append(
            Container(
                transaction_id=transaction_id,
                sequence_number=0,
                container_type=ContainerType.FIRST,
                total_length=total_length,
                payload=first_payload,
            )
        )

        offset = len(first_payload)
        seq = 1

        # Subsequent containers
        subsequent_max_payload = self.effective_mtu - SUBSEQUENT_HEADER_SIZE
        while offset < total_length:
            if seq > 255:
                raise ValueError(
                    f"Payload requires more than 256 containers (seq={seq}), "
                    "exceeding 8-bit sequence_number limit"
                )
            chunk = payload[offset : offset + subsequent_max_payload]
            containers.append(
                Container(
                    transaction_id=transaction_id,
                    sequence_number=seq,
                    container_type=ContainerType.SUBSEQUENT,
                    payload=chunk,
                )
            )
            offset += len(chunk)
            seq += 1

        return containers


class ContainerAssembler:
    """Reassembles containers into a complete payload."""

    def __init__(self):
        self._transactions: dict[int, _AssemblyState] = {}

    def feed(self, container: Container) -> bytes | None:
        """Feed a container. Returns complete payload when done, else None."""
        if container.container_type == ContainerType.CONTROL:
            return None  # Control containers are handled separately

        tid = container.transaction_id

        if container.container_type == ContainerType.FIRST:
            self._transactions[tid] = _AssemblyState(
                total_length=container.total_length,
                expected_seq=1,
                fragments=[container.payload],
                received_length=len(container.payload),
            )
        elif tid in self._transactions:
            state = self._transactions[tid]
            if container.sequence_number != state.expected_seq:
                # Sequence gap — discard entire transaction
                del self._transactions[tid]
                return None
            state.fragments.append(container.payload)
            state.received_length += len(container.payload)
            state.expected_seq += 1
        else:
            # Subsequent without a FIRST — ignore
            return None

        state = self._transactions[tid]
        if state.received_length >= state.total_length:
            payload = b"".join(state.fragments)[: state.total_length]
            del self._transactions[tid]
            return payload

        return None

    def reset(self):
        """Clear all pending assembly state."""
        self._transactions.clear()


@dataclass
class _AssemblyState:
    total_length: int
    expected_seq: int
    fragments: list[bytes] = field(default_factory=list)
    received_length: int = 0


def make_timeout_request(transaction_id: int, sequence_number: int = 0) -> Container:
    """Create a timeout request control container (Central -> Peripheral)."""
    return Container(
        transaction_id=transaction_id,
        sequence_number=sequence_number,
        container_type=ContainerType.CONTROL,
        control_cmd=ControlCmd.TIMEOUT,
        payload=b"",
    )


def make_timeout_response(
    transaction_id: int, timeout_ms: int, sequence_number: int = 0
) -> Container:
    """Create a timeout response control container (Peripheral -> Central)."""
    return Container(
        transaction_id=transaction_id,
        sequence_number=sequence_number,
        container_type=ContainerType.CONTROL,
        control_cmd=ControlCmd.TIMEOUT,
        payload=struct.pack("<H", timeout_ms),
    )


def make_stream_end_c2p(transaction_id: int, sequence_number: int = 0) -> Container:
    """Create stream end container (Central -> Peripheral)."""
    return Container(
        transaction_id=transaction_id,
        sequence_number=sequence_number,
        container_type=ContainerType.CONTROL,
        control_cmd=ControlCmd.STREAM_END_C2P,
        payload=b"",
    )


def make_stream_end_p2c(transaction_id: int, sequence_number: int = 0) -> Container:
    """Create stream end container (Peripheral -> Central)."""
    return Container(
        transaction_id=transaction_id,
        sequence_number=sequence_number,
        container_type=ContainerType.CONTROL,
        control_cmd=ControlCmd.STREAM_END_P2C,
        payload=b"",
    )
