#!/usr/bin/env sh
# Build all packages plus the terminull CLI binary into dist/.
# Usage: sh scripts/build.sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p dist logs
echo "==> go build ./..."
go build ./... 2>&1 | tee logs/go-build.txt
echo "==> building cmd/terminull"
go build -o dist/terminull ./cmd/terminull 2>&1 | tee -a logs/go-build.txt
ls -la dist/
echo "build OK"
