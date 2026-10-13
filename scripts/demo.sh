#!/usr/bin/env sh
# Smoke-run demo-app non-interactively. Skips cleanly if not built yet.
# Usage: sh scripts/demo.sh
set -eu
cd "$(dirname "$0")/.."
if [ ! -x "dist/demo" ] && [ ! -f "dist/demo" ] && [ ! -f "dist/demo.exe" ]; then
  echo "no dist/demo binary yet, skipping demo smoke"
  exit 0
fi
BIN="dist/demo"
if [ -f "dist/demo.exe" ]; then BIN="dist/demo.exe"; fi
echo "==> $BIN --help"
"$BIN" --help
echo "==> $BIN --version"
"$BIN" --version || true
echo "==> $BIN deploy --help"
"$BIN" deploy --help || true
echo "demo smoke OK"
