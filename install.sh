#!/usr/bin/env bash
set -euo pipefail

# User curls the url defined in the server routes and this script is executed on their machine

# Try to infer SERVER_URL from where the script was fetched
if [[ -n "${BASH_SOURCE[0]}" && "${BASH_SOURCE[0]}" != */* ]]; then
  echo "Unable to infer SERVER_URL — this script must be fetched remotely (e.g. via curl)."
  echo "Example: curl -fsSL https://your.app/install.sh | bash"
  exit 1
else
  SERVER_URL="$(dirname "$(curl -fsSL -I -o /dev/null -w %{url_effective} "$0")")"
fi

OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH=amd64 ;;
    aarch64) ARCH=arm64 ;;
esac

BINARY_URL="${SERVER_URL}/releases/chat-cli-${OS}-${ARCH}"
INSTALL_DIR="/usr/local/bin"
INSTALL_PATH="${INSTALL_DIR}/chat-cli"

echo "📦 Downloading chat-cli for ${OS}-${ARCH} from ${BINARY_URL}..."
sudo mkdir -p "$INSTALL_DIR"

# Try to download, exit if not found
if ! curl -fsSL "$BINARY_URL" -o "$INSTALL_PATH"; then
  echo "Failed to download client binary for ${OS}-${ARCH} from ${BINARY_URL}"
  echo "Make sure the server provides this build."
  exit 1
fi

sudo chmod +x "$INSTALL_PATH"

echo "Installed chat-cli to $INSTALL_PATH"
echo "Run 'chat-cli' to start."

