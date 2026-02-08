"""BLE transport layer using bleak."""

from __future__ import annotations

import asyncio
import logging

from bleak import BleakClient, BleakScanner

logger = logging.getLogger(__name__)

SERVICE_UUID = "12340001-0000-1000-8000-00805f9b34fb"
CHAR_UUID = "12340002-0000-1000-8000-00805f9b34fb"
DEVICE_NAME = "blerpc"
DEFAULT_TIMEOUT_S = 0.1  # 100ms


class BleTransport:
    """BLE transport wrapping bleak for blerpc communication."""

    def __init__(self):
        self._client: BleakClient | None = None
        self._notify_queue: asyncio.Queue[bytes] = asyncio.Queue()
        self._mtu: int = 23  # Default minimum BLE MTU

    @property
    def mtu(self) -> int:
        return self._mtu

    async def connect(self, device_name: str = DEVICE_NAME, timeout: float = 10.0):
        """Scan for device by name and connect."""
        logger.info("Scanning for device '%s'...", device_name)
        device = await BleakScanner.find_device_by_name(device_name, timeout=timeout)
        if device is None:
            raise ConnectionError(f"Device '{device_name}' not found")

        logger.info("Connecting to %s...", device.address)
        self._client = BleakClient(device)
        await self._client.connect()

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
