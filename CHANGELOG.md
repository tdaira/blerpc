# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Security (Audit Round 2)
- **C-1**: `require_encryption` parameter (default: True) in Python/Kotlin/Swift clients prevents silent encryption downgrade
- **C-2**: Race condition fix in firmware `req_work` — busy check before data write, optional double-buffer (`CONFIG_BLERPC_DOUBLE_BUFFER`)
- **H-1**: Kotlin crypto counters changed from `Int` to `Long` (full unsigned 32-bit range)
- **H-2**: `CONFIG_BLERPC_MAX_FLASH_READ_ADDRESS` Kconfig option restricts flash read range
- **H-3**: `counter_stream` count capped at 10,000 in firmware and Python peripheral
- **H-4**: `memset` replaced with `mbedtls_platform_zeroize` for key material in C crypto and firmware
- **H-5**: `known_keys.json` created with 0600 permissions
- **H-6**: Thread safety locks in Python peripheral for session, upload_count, and splitter
- **H-7**: `@Volatile` on Android BLE continuations; serial `DispatchQueue` for iOS CoreBluetooth
- **H-8**: `permissions: { contents: read }` added to all CI workflows
- New `BLERPC_ERROR_BUSY` (0x02) error code in all protocol libraries

### Changed
- Protocol libraries updated to 0.6.0

## [0.5.0] - 2026-02-22

### Added
- E2E encryption: X25519 ECDH + Ed25519 signatures + AES-128-GCM + HKDF-SHA256
- Ephemeral X25519 keypairs per connection for forward secrecy
- TOFU (Trust On First Use) key management for Ed25519 identity keys
- CAPABILITIES format extended to 6 bytes with encryption flags field
- `ControlCmd.KEY_EXCHANGE` (0x6) for 4-step key exchange handshake
- Flash read address bounds validation (integer overflow and out-of-bounds prevention)
- Mandatory encryption mode: reject unencrypted payloads when encryption is enabled
- Key exchange state machine validation (prevents out-of-order steps)
- TX counter overflow check (prevents AES-GCM nonce reuse at 2^32-1)
- Thread-safe encrypt/decrypt in Python, Kotlin, and Swift crypto sessions
- Disconnect state reset for Python peripheral and firmware
- CONTRIBUTING.md, SECURITY.md, and CHANGELOG.md for OSS readiness
- GitHub issue and pull request templates
- Python edge-case tests (malformed protobuf, empty data, special characters)
- Type hints for Python client public API
- Public API exports in `blerpc/__init__.py`
- Documentation comments in `proto/blerpc.proto`

### Changed
- Protocol libraries updated to 0.5.0
- Narrowed exception handling in Android/iOS connect (catch specific timeout errors)
- Replaced magic numbers with named constants in C and Python
- Fixed Android deprecated API usage with API 33+ version branching

### Fixed
- Android scan defaulting to unfiltered (no service UUID filter)
- Python client crashing with `AttributeError` when called before `connect()`
- Buffer overflow in peripheral firmware command header parsing

## [0.1.0] - 2025-02-09

Initial public release.

### Added
- RPC-over-BLE framework with Protocol Buffers
- Code generation from `.proto` files (Go tool)
- Central clients: Python (macOS), Swift (iOS), Kotlin (Android), C (Zephyr)
- Peripheral servers: C (Zephyr on nRF54L15 DK, EFR32xG22E EK), Python (macOS)
- MTU-aware container fragmentation and reassembly
- Streaming support (P->C and C->P)
- Capabilities negotiation and timeout control
- Multi-device scan and select
- Unit and integration test suites
