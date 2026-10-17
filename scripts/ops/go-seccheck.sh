#!/usr/bin/env bash

set -o pipefail

# Always run from the repo root regardless of where the script is invoked.
cd "$(dirname "$0")/../.."

# gosec's x/tools loader cannot read the export data of a newer Go
# compiler, so analyze with the toolchain the module declares whenever
# the local Go is a newer series.
mod_go=$(sed -n 's/^go \([0-9][0-9.]*\).*/\1/p' go.mod | head -n 1)
local_go=$(go env GOVERSION | sed 's/^go//')
if [ -n "$mod_go" ] && [ -z "$GOTOOLCHAIN" ] && [ "${local_go%.*}" != "${mod_go%.*}" ]; then
    GOTOOLCHAIN="go$mod_go"
    export GOTOOLCHAIN
fi

mkdir -p logs

if [ -z "$(go list ./... 2>/dev/null)" ]; then
    echo "no Go packages yet, skipping gosec"
    exit 0
fi

failed=0

run_check() {
    local name="$1"
    shift

    echo "==> Running $name"

    if "$@" 2>&1 | tee "logs/go-seccheck-${name}.txt"; then
        echo "    $name: OK"
    else
        echo "    $name: FAILED"
        failed=1
    fi

    echo
}

run_check gosec gosec ./...

exit "$failed"