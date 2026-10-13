#!/usr/bin/env sh
# End-to-end smoke: build the terminull CLI, scaffold a project with it,
# then build and run the scaffold. Exercises the framework as an
# external project would use it.
# Usage: sh scripts/smoke.sh
set -eu
cd "$(dirname "$0")/.."
sh scripts/build.sh
go build -o dist/terminull ./cmd/terminull
WORK="$(pwd)/dist/smoke-app"
rm -rf "$WORK"
./dist/terminull new smokeapp --dir "$WORK" --module example.com/smokeapp --replace "$(pwd)"
cd "$WORK"
go build ./...
go build -o smokeapp ./cmd/smokeapp
./smokeapp --help
./smokeapp hello --name smoke
./smokeapp status -o json
echo "smoke OK"
