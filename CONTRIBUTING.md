# Contributing to bleRPC

Thank you for your interest in contributing to bleRPC!

## Getting Started

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Submit a pull request

## Development Setup

### Python (Central / Peripheral)

```bash
pip install pytest pytest-asyncio protobuf bleak ruff
pip install git+https://github.com/tdaira/blerpc-protocol.git
```

Run tests:

```bash
cd central_py
python -m pytest tests/test_container.py tests/test_command.py tests/test_client.py -v
```

### Android

```bash
cd central_android
./gradlew assembleDebug
```

### iOS

```bash
cd central_ios
brew install xcodegen
xcodegen generate
xcodebuild -project BlerpcCentral.xcodeproj \
  -scheme BlerpcCentral \
  -destination 'generic/platform=iOS' \
  CODE_SIGNING_ALLOWED=NO build
```

### Code Generation

When modifying `proto/blerpc.proto`, regenerate handlers:

```bash
cd tools/generate-handlers
go run . ../../proto/blerpc.proto
```

## Code Style

- **Python**: Formatted with [ruff](https://docs.astral.sh/ruff/) (line length 88)
- **C**: Formatted with clang-format (LLVM style)
- **Kotlin/Swift**: Follow platform conventions

All style checks run automatically in CI.

## Pull Requests

- Keep PRs focused on a single change
- Include tests for new functionality
- Ensure CI passes before requesting review
- Write commit messages that explain *why*, not just *what*

## Reporting Issues

Use [GitHub Issues](https://github.com/tdaira/blerpc/issues) to report bugs or request features. Include:

- Platform and OS version
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs

## License

By contributing, you agree that your contributions will be licensed under the [LGPL-3.0](LICENSE) with [Static Linking Exception](LICENSING_EXCEPTION).
