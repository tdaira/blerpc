# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in bleRPC, please report it responsibly.

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please email the maintainer directly. You can find contact information on the [GitHub profile](https://github.com/tdaira).

## Scope

bleRPC communicates over Bluetooth Low Energy, which has inherent security considerations:

- **No encryption at the application layer**: bleRPC relies on the BLE link-layer encryption provided by the OS/stack. Ensure BLE pairing and bonding are configured appropriately for your use case.
- **No authentication**: bleRPC does not implement application-level authentication. Any device that can connect to the GATT service can send RPC requests.

## Supported Versions

Security fixes are applied to the latest release on the `main` branch.
