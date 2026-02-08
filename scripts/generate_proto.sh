#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PROTO_DIR="$PROJECT_ROOT/proto"

# Generate Python protobuf code
protoc \
    --python_out="$PROJECT_ROOT/central/blerpc/generated/" \
    --proto_path="$PROTO_DIR" \
    "$PROTO_DIR/blerpc.proto"

echo "Python protobuf code generated."

# Generate nanopb C code (for peripheral)
NANOPB_DIR=$(python3 -c "import nanopb; import os; print(os.path.dirname(nanopb.__file__))" 2>/dev/null || true)
if [ -n "$NANOPB_DIR" ] && [ -f "$NANOPB_DIR/generator/nanopb_generator.py" ]; then
    python3 "$NANOPB_DIR/generator/nanopb_generator.py" \
        -I "$PROTO_DIR" \
        -D "$PROJECT_ROOT/peripheral/src" \
        "$PROTO_DIR/blerpc.proto"
    echo "nanopb C code generated."
else
    echo "nanopb not found, skipping C code generation."
fi

# Generate handler stubs and client code
python3 "$SCRIPT_DIR/generate_handlers.py"
echo "Handler and client code generated."

# Format generated Python files if ruff is available
if python3 -m ruff --version >/dev/null 2>&1; then
    python3 -m ruff format \
        "$PROJECT_ROOT/peripheral_py/generated_handlers.py" \
        "$PROJECT_ROOT/central/blerpc/generated/generated_client.py" \
        2>/dev/null || true
    echo "Generated Python files formatted."
fi
