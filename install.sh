#!/bin/sh
set -e

REPO="ErikHellman/hemkop-cli"
INSTALL_DIR="$HOME/.local/bin"
BINARY="hemkop"

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *)              echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest release tag
LATEST=$(curl -sI "https://github.com/$REPO/releases/latest" | grep -i '^location:' | sed 's|.*/||' | tr -d '\r')
if [ -z "$LATEST" ]; then
  echo "Failed to determine latest release"
  exit 1
fi

URL="https://github.com/$REPO/releases/download/$LATEST/hemkop-${OS}-${ARCH}"
echo "Downloading hemkop $LATEST for ${OS}/${ARCH}..."
curl -fsSL -o "$BINARY" "$URL"
chmod +x "$BINARY"

# Install
mkdir -p "$INSTALL_DIR"
mv "$BINARY" "$INSTALL_DIR/$BINARY"

# Check if INSTALL_DIR is in PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "NOTE: $INSTALL_DIR is not in your PATH. Add it with:"
     echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.$(basename "$SHELL")rc"
     ;;
esac

echo "hemkop installed to $INSTALL_DIR/$BINARY"
