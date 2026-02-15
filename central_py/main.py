"""blerpc Central entry point."""

import asyncio
import logging
import time

from blerpc.client import BlerpcClient

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s"
)
logger = logging.getLogger(__name__)


async def main():
    client = BlerpcClient()

    try:
        devices = await client.scan()
        if not devices:
            raise ConnectionError("No blerpc devices found")
        logger.info(
            "Found %d device(s), connecting to %s (RSSI %d)...",
            len(devices),
            devices[0].name,
            devices[0].rssi,
        )
        await client.connect(devices[0])
        logger.info("MTU: %d", client.mtu)

        # Echo test
        message = "Hello, blerpc!"
        result = await client.echo(message)
        logger.info(
            "Echo: sent=%r received=%r match=%s", message, result, message == result
        )

        # Flash read test
        address = 0x00000000
        length = 16
        data = await client.flash_read(address, length)
        logger.info(
            "Flash read: address=0x%08X length=%d data=%s", address, length, data.hex()
        )

        # Flash read speed test
        address = 0x00000000
        length = 4096
        start = time.monotonic()
        data = await client.flash_read(address, length)
        elapsed = time.monotonic() - start
        throughput = len(data) / elapsed if elapsed > 0 else 0
        logger.info(
            "Flash speed test: %d bytes in %.3fs = %.1f bytes/s (%.1f kbps)",
            len(data),
            elapsed,
            throughput,
            throughput * 8 / 1000,
        )

    finally:
        await client.disconnect()


if __name__ == "__main__":
    asyncio.run(main())
