#!/usr/bin/env bash
set -e

echo "👉 Detecting platform..."

# Detect OS
OS="$(uname -s)"
case "$OS" in
    Linux*)     PLATFORM="Linux";;
    Darwin*)    PLATFORM="Darwin";;
    MINGW*|MSYS*|CYGWIN*|Windows_NT) PLATFORM="Windows";;
    *)          echo "❌ Unsupported OS: $OS"; exit 1;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64) ARCH="x86_64";;
    amd64)  ARCH="x86_64";; # Just in case
    arm64|aarch64) ARCH="arm64";;
    *) echo "❌ Unsupported architecture: $ARCH"; exit 1;;
esac

echo "✅ Platform: $PLATFORM"
echo "✅ Architecture: $ARCH"


# Get latest release tag
LATEST_TAG=$(curl -sI https://github.com/formancehq/numscript/releases/latest | grep -i location | awk -F '/' '{print $NF}' | tr -d '\r')

echo "📦 Latest tag: $LATEST_TAG"

# Determine file extension
if [ "$PLATFORM" = "Windows" ]; then
    EXT="zip"
else
    EXT="tar.gz"
fi

# Build file name and URL
FILENAME="numscript_${LATEST_TAG#v}_${PLATFORM}_${ARCH}.${EXT}"
URL="https://github.com/formancehq/numscript/releases/download/$LATEST_TAG/$FILENAME"

echo "⬇️ Downloading: $URL"
curl -L -o "$FILENAME" "$URL"

# Extract and install
if [ "$PLATFORM" = "Windows" ]; then
    unzip "$FILENAME"
    BIN="numscript.exe"
else
    tar -xf "$FILENAME" numscript
    BIN="numscript"
fi


INSTALL_PATH="$HOME/.local/bin"
mkdir -p "$INSTALL_PATH"

mv "$BIN" "$INSTALL_PATH"
chmod +x "$INSTALL_PATH/$BIN"

rm "$FILENAME"

echo "✅ Installed $BIN to $INSTALL_PATH"
echo "🎉 Done!"
