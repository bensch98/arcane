#!/usr/bin/env bash
set -euo pipefail

REPO="bensch98/arcane"
INSTALL_DIR="/usr/local/bin"
BINARY="arcane"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "Downloading ${ASSET}..."
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

if command -v curl &>/dev/null; then
  curl -fsSL "$URL" -o "$TMP"
elif command -v wget &>/dev/null; then
  wget -qO "$TMP" "$URL"
else
  echo "Error: curl or wget required" >&2; exit 1
fi

chmod +x "$TMP"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "${INSTALL_DIR}/${BINARY}"
else
  sudo mv "$TMP" "${INSTALL_DIR}/${BINARY}"
fi

echo "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
${BINARY} version 2>/dev/null || true
