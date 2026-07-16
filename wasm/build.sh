#!/usr/bin/env bash
set -euo pipefail

# Builds the numscript WASM bindings and copies the matching wasm_exec.js glue.
# Output lands in wasm/dist/. Copy both files into the consumer app.

cd "$(dirname "$0")/.."

OUT_DIR="wasm/dist"
mkdir -p "$OUT_DIR"

GOOS=js GOARCH=wasm GOTOOLCHAIN=auto go build -o "$OUT_DIR/numscript.wasm" ./wasm

GOROOT="$(GOTOOLCHAIN=auto go env GOROOT)"
install -m 644 "$GOROOT/lib/wasm/wasm_exec.js" "$OUT_DIR/wasm_exec.js"

# Provenance: record the numscript commit this artifact was built from, so a
# copied .wasm can be traced back to source (and drift spotted).
git rev-parse HEAD > "$OUT_DIR/numscript.version"

echo "built (numscript $(cat "$OUT_DIR/numscript.version")):"
ls -la "$OUT_DIR"
