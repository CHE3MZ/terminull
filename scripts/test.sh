#!/usr/bin/env sh
# Run go tests with log capture.
# Usage: sh scripts/test.sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p logs
echo "==> go test ./..."
go test ./... 2>&1 | tee logs/go-test.txt
echo "test OK"
