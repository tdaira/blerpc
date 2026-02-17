# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- CONTRIBUTING.md, SECURITY.md, and CHANGELOG.md for OSS readiness
- GitHub issue and pull request templates
- Python edge-case tests (malformed protobuf, empty data, special characters)
- Type hints for Python client public API
- Public API exports in `blerpc/__init__.py`
- Documentation comments in `proto/blerpc.proto`

### Changed
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
