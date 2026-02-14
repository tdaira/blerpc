# bleRPC

Type-safe, high-performance RPC over Bluetooth Low Energy using Protocol Buffers.

**Documentation: [blerpc.net](https://blerpc.net)**

## Overview

bleRPC generates client and handler code from `.proto` files for multiple platforms, enabling type-safe RPC calls over BLE GATT with automatic MTU-aware fragmentation and reassembly.

- Define services once in Protocol Buffers, generate code for each platform
- ~30 KB/s throughput over BLE
- Runs on devices with as little as 32 KB RAM

## Supported Platforms

### Central (Client)

| Platform | Language |
|----------|----------|
| macOS | Python |
| iOS | Swift |
| Android | Kotlin |
| nRF54L15 DK | C (Zephyr) |

### Peripheral (Server)

| Platform | Language |
|----------|----------|
| nRF54L15 DK | C (Zephyr) |
| EFR32xG22E EK | C (Zephyr) |
| macOS | Python |

## Repository Structure

| Directory | Description |
|-----------|-------------|
| `proto/` | Protocol Buffer definitions |
| `central/` | Python Central client (macOS) |
| `central_fw/` | C Central firmware (nRF54L15 DK / Zephyr) |
| `central_ios/` | Swift Central app (iOS) |
| `central_android/` | Kotlin Central app (Android) |
| `peripheral/` | C Peripheral firmware (nRF54L15 DK, EFR32xG22E / Zephyr) |
| `peripheral_py/` | Python Peripheral server (macOS) |
| `boards/` | Custom Zephyr board definitions |
| `tools/` | Code generation and debugging tools |
| `docs/` | Firmware build and flash guide |

## Getting Started

See the [Getting Started](https://blerpc.net/getting-started.html) guide.

## Protocol Libraries

| Language | Repository |
|----------|------------|
| Python / C | [blerpc-protocol](https://github.com/tdaira/blerpc-protocol) |
| Swift | [blerpc-protocol-swift](https://github.com/tdaira/blerpc-protocol-swift) |
| Kotlin | [blerpc-protocol-kt](https://github.com/tdaira/blerpc-protocol-kt) |

## License

[LGPL-3.0](LICENSE)
