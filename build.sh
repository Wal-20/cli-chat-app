#!/usr/bin/env bash
set -euo pipefail

NO_SOURCE=false
for arg in "$@"; do
    case "$arg" in
        --no-source) NO_SOURCE=true ;;
    esac
done

if [ "$NO_SOURCE" = false ]; then
    set -a
    [ -f .env ] && . .env
    set +a
fi

RELEASE_DIR="./releases"
mkdir -p "$RELEASE_DIR"

if [[ -z "${SERVER_URL:-}" ]]; then
  echo "❌ SERVER_URL not found in environment or .env"
  exit 1
fi

# Base64-encode the server URL to embed safely into the binary
SERVER_URL_B64=$(echo -n "$SERVER_URL" | base64)

echo "Building server..."
go build -o "$RELEASE_DIR/server" main.go
chmod +x "$RELEASE_DIR/server"

build_client() {
  local GOOS=$1
  local GOARCH=$2
  local OUTFILE="$RELEASE_DIR/chat-cli-${GOOS}-${GOARCH}"

  [[ "$GOOS" == "windows" ]] && OUTFILE="${OUTFILE}.exe"

  echo "Building client for ${GOOS}/${GOARCH}..."
  GOOS=$GOOS GOARCH=$GOARCH go build \
    -ldflags "-X github.com/Wal-20/cli-chat-app/internal/tui/client.DefaultServerURLB64=${SERVER_URL_B64}" \
    -o "$OUTFILE" ./internal/tui

  chmod +x "$OUTFILE"
}

build_client linux amd64

build_client darwin arm64

build_client darwin amd64

build_client windows amd64

echo "All builds completed successfully."
echo "Binaries are in: $RELEASE_DIR"

