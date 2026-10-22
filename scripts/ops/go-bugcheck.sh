#!/usr/bin/env bash

set -o pipefail

# Always run from the repo root regardless of where the script is invoked.
cd "$(dirname "$0")/../.."

mkdir -p logs

failed=0

run_check() {
    local name="$1"
    shift

    echo "==> Running $name"

    if "$@" 2>&1 | tee "logs/go-bugcheck-${name}.txt"; then
        echo "    $name: OK"
    else
        echo "    $name: FAILED"
        failed=1
    fi

    echo
}

run_check golangci-lint golangci-lint run ./...

# Pass Go files safely, including paths containing spaces. Vendored engine
# forks live under each package's internal tree and stay out of our lint gates.
if [ -n "$(find . \( \
    -path './termcore/internal/*' -o \
    -path './termstyle/internal/*' -o \
    -path './termretry/internal/*' -o \
    -path './termoutput/internal/*' -o \
    -path './termprompt/internal/*' -o \
    -path './termconfig/internal/*' -o \
    -path './termapp/internal/*' -o \
    -path './dist/*' \
\) -prune -o -name '*.go' -type f -print -quit)" ]; then
    run_check gopls sh -c 'find . \( -path "./termcore/internal/*" -o -path "./termstyle/internal/*" -o -path "./termretry/internal/*" -o -path "./termoutput/internal/*" -o -path "./termprompt/internal/*" -o -path "./termconfig/internal/*" -o -path "./termapp/internal/*" -o -path "./dist/*" \) -prune -o -name "*.go" -type f -print0 | xargs -0 -n 50 gopls check'
else
    echo "No Go files found; skipping gopls check."
fi

run_check govulncheck govulncheck ./...
run_check staticcheck sh -c 'go list ./... | grep -Ev "/(termcore|termstyle|termretry|termoutput|termprompt|termconfig|termapp)/internal/" | xargs staticcheck'
run_check errcheck sh -c 'go list ./... | grep -Ev "/(termcore|termstyle|termretry|termoutput|termprompt|termconfig|termapp)/internal/" | xargs errcheck'
run_check vet go vet ./...
run_check test go test ./...

exit "$failed"
