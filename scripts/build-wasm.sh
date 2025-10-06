#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC_DIR="$REPO_ROOT/src"
OUTPUT="$REPO_ROOT/plugin.wasm"

echo "Building plugin.wasm using TinyGo (local) or Docker fallback"

if command -v tinygo >/dev/null 2>&1; then
  echo "Using local tinygo"
  (cd "$SRC_DIR" && tinygo build -o "$OUTPUT" -scheduler=none --no-debug -target=wasi .)
  echo "Built $OUTPUT"
  exit 0
fi

if command -v docker >/dev/null 2>&1; then
  echo "Using Docker tinygo image"
  docker run --rm -v "$REPO_ROOT":/work -w /work/src tinygo/tinygo:0.34.0 tinygo build -o /work/plugin.wasm -scheduler=none --no-debug -target=wasi .
  echo "Built $OUTPUT"
  exit 0
fi

echo "Neither tinygo nor docker found. Install tinygo or docker and retry." >&2
exit 1
