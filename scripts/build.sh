#!/usr/bin/env sh
# Build all packages plus the terminull CLI natively and for every
# supported platform (mirrors CI build-matrix and goreleaser targets).
# Usage: sh scripts/build.sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p dist logs
echo "==> go build ./..."
go build ./... 2>&1 | tee logs/go-build.txt
echo "==> building cmd/terminull (native)"
go build -o dist/terminull ./cmd/terminull 2>&1 | tee -a logs/go-build.txt
echo "==> building cmd/terminull (all platforms)"
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64; do
  goos=${target%/*}
  goarch=${target#*/}
  out="dist/terminull-${goos}-${goarch}"
  if [ "$goos" = "windows" ]; then out="$out.exe"; fi
  echo "==> $out"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -o "$out" ./cmd/terminull 2>&1 | tee -a logs/go-build.txt
done
ls -la dist/
echo "build OK"
