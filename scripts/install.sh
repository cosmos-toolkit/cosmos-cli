#!/usr/bin/env sh
# Install cosmos from GitHub Releases (no Go required).
# Usage: curl -sSL https://raw.githubusercontent.com/cosmos-toolkit/cosmos-cli/main/scripts/install.sh | sh

set -e

REPO="cosmos-toolkit/cosmos-cli"
API="https://api.github.com/repos/${REPO}/releases/latest"
BINARY="cosmos"

# Detect OS and arch (Linux/macOS only; Windows: download from Releases page)
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
  Darwin)  OS="darwin" ;;
  Linux)   OS="linux" ;;
  *)
    echo "Unsupported OS: $OS. For Windows, download from https://github.com/${REPO}/releases"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported arch: $ARCH"
    exit 1
    ;;
esac

SUFFIX="${OS}-${ARCH}.tar.gz"
echo "Installing cosmos (${OS}/${ARCH})..."

# Resolve download URL (prefer jq; fallback grep/sed)
if command -v jq >/dev/null 2>&1; then
  DOWNLOAD_URL=$(curl -sSL "$API" | jq -r ".assets[] | select(.name | endswith(\"${SUFFIX}\")) | .browser_download_url")
else
  # Fallback: grep for the asset URL containing our suffix
  DOWNLOAD_URL=$(curl -sSL "$API" | grep "browser_download_url" | grep "$SUFFIX" | head -1 | sed 's/.*"browser_download_url": "\([^"]*\)".*/\1/')
fi

if [ -z "$DOWNLOAD_URL" ] || [ "$DOWNLOAD_URL" = "null" ]; then
  echo "No pre-built binary for ${OS}/${ARCH}. See https://github.com/${REPO}/releases"
  exit 1
fi

# Prefer ~/.local/bin; fallback to /usr/local/bin if writable
if [ -w /usr/local/bin ] 2>/dev/null; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -sSL -o "$TMP_DIR/archive.tar.gz" "$DOWNLOAD_URL"
tar -xzf "$TMP_DIR/archive.tar.gz" -C "$TMP_DIR"

# Binary may be at root of tarball or in a subdir (e.g. cosmos-1.0.0-darwin-arm64/cosmos)
if [ -f "$TMP_DIR/$BINARY" ]; then
  mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
  BIN=$(find "$TMP_DIR" -maxdepth 2 -type f -name "$BINARY" 2>/dev/null | head -1)
  if [ -n "$BIN" ]; then
    mv "$BIN" "$INSTALL_DIR/$BINARY"
  else
    echo "Binary not found in archive"
    exit 1
  fi
fi

chmod +x "$INSTALL_DIR/$BINARY"
echo "Installed to $INSTALL_DIR/$BINARY"

if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo ""
  echo "Add to your PATH (e.g. in ~/.zshrc or ~/.bashrc):"
  echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
fi
