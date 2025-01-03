#!/usr/bin/env bash
set -e

# Default version (can be overridden by environment variable)
VERSION=${VERSION:-latest}
REPO="PagoPlus/numscript-wasm"

if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi

echo "Downloading numscript.wasm version $VERSION..."

# Create directory if it doesn't exist
mkdir -p priv/wasm

# Download the file
curl -L "https://github.com/$REPO/releases/download/$VERSION/numscript.wasm" -o priv/wasm/numscript.wasm

echo "Downloaded successfully to priv/wasm/numscript.wasm"