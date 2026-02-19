"""BLE transport layer using bleak."""

from __future__ import annotations

import asyncio
import logging
from dataclasses import dataclass, field

from bleak import BleakClient, BleakScanner
from bleak.backends.device import BLEDevice

logger = logging.getLogger(__name__)

SERVICE_UUID = "12340001-0000-1000-8000-00805f9b34fb"
CHAR_UUID = "12340002-0000-1000-8000-00805f9b34fb"
DEFAULT_TIMEOUT_S = 0.1  # 100ms


@dataclass
class ScannedDevice:
    """A BLE device discovered during scanning."""

    name: str | None
    address: str
    rssi: int
    manufacturer_data: dict[int, bytes] = field(default_factory=dict)
    service_data: dict[str, bytes] = field(default_factory=dict)
    service_uuids: list[str] = field(default_factory=list)
    tx_power: int | None = None
    _bleak_device: BLEDevice = field(repr=False, default=None)


class BleTransport:
    """BLE transport wrapping bleak for blerpc communication."""

    def __init__(self):
        self._client: BleakClient | None = None
        self._notify_queue: asyncio.Queue[bytes] = asyncio.Queue()
        self._mtu: int = 23  # Default minimum BLE MTU
        self._address: str | None = None

    @property
    def mtu(self) -> int:
        return self._mtu

    @property
    def address(self) -> str | None:
        return self._address

    async def scan(
        self,
        timeout: float = 5.0,
        service_uuid: str | None = SERVICE_UUID,
    ) -> list[ScannedDevice]:
        """Scan for BLE devices. If service_uuid is set, filter by it."""
        service_uuids = [service_uuid] if service_uuid else None
        devices = await BleakScanner.discover(
            timeout=timeout,
            return_adv=True,
            service_uuids=service_uuids,
        )
        result = []
        for device, adv_data in devices.values():
            result.append(
                ScannedDevice(
                    name=device.name,
                    address=device.address,
                    rssi=adv_data.rssi,
                    manufacturer_data=dict(adv_data.manufacturer_data),
                    service_data={str(k): v for k, v in adv_data.service_data.items()},
                    service_uuids=list(adv_data.service_uuids),
                    tx_power=adv_data.tx_power,
                    _bleak_device=device,
                )
            )
        return sorted(result, key=lambda d: d.rssi, reverse=True)

    async def connect(self, device: ScannedDevice):
        """Connect to a previously scanned device."""
        logger.info("Connecting to %s...", device.address)
        self._client = BleakClient(device._bleak_device)
        await self._client.connect()
        self._address = device.address

        # Get negotiated MTU
        self._mtu = self._client.mtu_size
        logger.info("Connected. MTU=%d", self._mtu)

        # Start notifications
        await self._client.start_notify(CHAR_UUID, self._notify_handler)

    def _notify_handler(self, _sender, data: bytearray):
        """Callback for BLE notifications."""
        self._notify_queue.put_nowait(bytes(data))

    async def write(self, data: bytes):
        """Write data to the characteristic (write without response)."""
        if not self._client or not self._client.is_connected:
            raise ConnectionError("Not connected")
        await self._client.write_gatt_char(CHAR_UUID, data, response=False)

    async def read_notify(self, timeout: float = DEFAULT_TIMEOUT_S) -> bytes:
        """Wait for a notification with timeout."""
        return await asyncio.wait_for(self._notify_queue.get(), timeout=timeout)

    async def disconnect(self):
        """Disconnect from the device."""
        if self._client and self._client.is_connected:
            try:
                await self._client.stop_notify(CHAR_UUID)
            except Exception:
                pass  # May already be disconnected
            try:
                await self._client.disconnect()
            except Exception:
                pass
            logger.info("Disconnected")
        self._client = None

    @property
    def is_connected(self) -> bool:
        return self._client is not None and self._client.is_connected
