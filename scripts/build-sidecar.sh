#!/usr/bin/env bash
set -euo pipefail

# Build Go sidecar binary for current platform and place it in src-tauri/binaries/
# Tauri expects the binary to be named: autoshift-server-{target-triple}
# Target triples:
#   Linux x86_64:   x86_64-unknown-linux-gnu
#   macOS aarch64:  aarch64-apple-darwin
#   Windows x86_64: x86_64-pc-windows-msvc

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
BACKEND_DIR="$ROOT_DIR/backend"
OUT_DIR="$ROOT_DIR/src-tauri/binaries"

mkdir -p "$OUT_DIR"

detect_target() {
    local os arch triple

    case "$(uname -s)" in
        Linux)  os="unknown-linux-gnu" ;;
        Darwin) os="apple-darwin" ;;
        MINGW*|MSYS*|CYGWIN*) os="pc-windows-msvc" ;;
        *) echo "unsupported OS: $(uname -s)"; exit 1 ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64) arch="x86_64" ;;
        aarch64|arm64) arch="aarch64" ;;
        *) echo "unsupported arch: $(uname -m)"; exit 1 ;;
    esac

    echo "${arch}-${os}"
}

TARGET=$(detect_target)
BINARY_NAME="autoshift-server-${TARGET}"

echo "Building sidecar for target: ${TARGET}"
cd "$BACKEND_DIR"

# Build with Go
GOOS="${TARGET%%-*}"
GOARCH=""
case "$TARGET" in
    x86_64-*) GOARCH="amd64" ;;
    aarch64-*) GOARCH="arm64" ;;
esac

case "$TARGET" in
    *-windows-*)
        BINARY_NAME="${BINARY_NAME}.exe"
        GOOS="windows"
        ;;
    *-apple-darwin)
        GOOS="darwin"
        ;;
    *-linux-gnu)
        GOOS="linux"
        ;;
esac

env GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o "$OUT_DIR/$BINARY_NAME" .

echo "Sidecar built: $OUT_DIR/$BINARY_NAME"
ls -lh "$OUT_DIR/$BINARY_NAME"
