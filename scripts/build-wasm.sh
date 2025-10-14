#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC_DIR="$REPO_ROOT/src"
OUTPUT="$REPO_ROOT/plugin.wasm"

echo "Building plugin.wasm using TinyGo (local) or Docker fallback"

# Default module path and destination dir for local plugins tree. Honor
# MODULE_PATH env var if set.
MODULE_PATH_VAL="${MODULE_PATH:-github.com/checkin247/traefik-wasm-query-param-check}"
DEST_DIR="$REPO_ROOT/plugins-local/src/$MODULE_PATH_VAL"

# Ensure destination exists so docker-compose mounts pick up changes; remove
# any stale artifact at repo root so build produces a fresh file.
mkdir -p "$DEST_DIR"
rm -f "$OUTPUT" || true

if command -v tinygo >/dev/null 2>&1; then
  echo "Using local tinygo"
  # Disable VCS stamping inside the build so tinygo doesn't fail in minimal images
  (cd "$SRC_DIR" && GOFLAGS='-buildvcs=false' tinygo build -o "$OUTPUT" -scheduler=none --no-debug -target=wasi .)
  # try to optimize with wasm-opt (Binaryen) if available
  if [ -n "${WASMOPT:-}" ]; then
    WASMOPT_PATH="$WASMOPT"
  elif command -v wasm-opt >/dev/null 2>&1; then
    WASMOPT_PATH="$(command -v wasm-opt)"
  else
    WASMOPT_PATH=""
  fi

  if [ -n "$WASMOPT_PATH" ]; then
    echo "Optimizing wasm with: $WASMOPT_PATH"
    tmpfile="$(mktemp -u)".wasm
    if "$WASMOPT_PATH" -O2 -o "$tmpfile" "$OUTPUT"; then
      mv -f "$tmpfile" "$OUTPUT"
      echo "Optimized $OUTPUT"
    else
      echo "wasm-opt run failed, leaving unoptimized artifact" >&2
      [ -f "$tmpfile" ] && rm -f "$tmpfile"
    fi
  else
    echo "WARNING: wasm-opt not found. Install Binaryen or set WASMOPT to its path to produce optimized WASM artifacts." >&2
  fi

  echo "Built $OUTPUT"
  echo "Copying artifacts to: $DEST_DIR"
  cp -f "$OUTPUT" "$DEST_DIR/"
  if [ -f "$REPO_ROOT/.traefik.yml" ]; then
    cp -f "$REPO_ROOT/.traefik.yml" "$DEST_DIR/"
  fi
  echo "Copied plugin.wasm and any detected traefik config to plugins-local tree"
  exit 0
fi

if command -v docker >/dev/null 2>&1; then
  echo "Using Docker tinygo image"
  docker run --rm -e GOFLAGS=-buildvcs=false -v "$REPO_ROOT":/work -w /work/src tinygo/tinygo:0.34.0 tinygo build -o /work/plugin.wasm -scheduler=none --no-debug -target=wasi .

  # Try to optimize with host wasm-opt if available (WASMOPT or wasm-opt in PATH)
  if [ -n "${WASMOPT:-}" ]; then
    WASMOPT_PATH="$WASMOPT"
  elif command -v wasm-opt >/dev/null 2>&1; then
    WASMOPT_PATH="$(command -v wasm-opt)"
  else
    WASMOPT_PATH=""
  fi

  if [ -n "$WASMOPT_PATH" ]; then
    echo "Optimizing wasm with: $WASMOPT_PATH"
    tmpfile="$(mktemp -u)".wasm
    if "$WASMOPT_PATH" -O2 -o "$tmpfile" "$OUTPUT"; then
      mv -f "$tmpfile" "$OUTPUT"
      echo "Optimized $OUTPUT"
    else
      echo "wasm-opt run failed, leaving unoptimized artifact" >&2
      [ -f "$tmpfile" ] && rm -f "$tmpfile"
    fi
  else
    echo "WARNING: wasm-opt not found. Install Binaryen or set WASMOPT to its path to produce optimized WASM artifacts." >&2
  fi

  echo "Built $OUTPUT"
  echo "Copying artifacts to: $DEST_DIR"
  cp -f "$OUTPUT" "$DEST_DIR/"
  if [ -f "$REPO_ROOT/.traefik.yml" ]; then
    cp -f "$REPO_ROOT/.traefik.yml" "$DEST_DIR/"
  fi
  echo "Copied plugin.wasm and any detected traefik config to plugins-local tree"
  exit 0
fi

echo "Neither tinygo nor docker found. Install tinygo or docker and retry." >&2
exit 1
