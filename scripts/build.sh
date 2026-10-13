#!/usr/bin/env sh
# Build all packages. If cmd/demo exists, also produce dist/demo binary.
# Usage: sh scripts/build.sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p dist logs
echo "==> go build ./..."
go build ./... 2>&1 | tee logs/go-build.txt
if [ -d "cmd/demo" ]; then
  echo "==> building cmd/demo"
  go build -o dist/demo ./cmd/demo 2>&1 | tee -a logs/go-build.txt
  ls -la dist/
else
  echo "no cmd/demo yet, skipping binary build"
fi
echo "build OK"
