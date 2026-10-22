#!/usr/bin/env sh
# Build and run a throwaway module for every public term* library.
# Each module imports exactly one library and replaces terminull with this
# checkout, so the library must stand on its own.
# Usage: sh scripts/standalone.sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
case "$(uname -s)" in
	MINGW* | MSYS* | CYGWIN*) ROOT=$(cygpath -m "$ROOT") ;;
esac

TMP=$(mktemp -d "${TMPDIR:-/tmp}/terminull-standalone.XXXXXX")
trap 'rm -rf "$TMP"' EXIT INT TERM

EXE=
case "$(uname -s)" in
	MINGW* | MSYS* | CYGWIN*) EXE=.exe ;;
esac

LIBS="termcore termstyle termbox termtable termimage termoutput termprogress
termprompt termrender termpager termstatic termretry termconfig termapp"

for lib in $LIBS; do
	dir="$TMP/$lib"
	mkdir -p "$dir"

	cat >"$dir/go.mod" <<EOF
module example.com/standalone/$lib

go 1.26.0

require github.com/CHE3MZ/terminull v0.0.0

replace github.com/CHE3MZ/terminull => "$ROOT"
EOF

	cat >"$dir/main.go" <<EOF
package main

import (
	"fmt"

	_ "github.com/CHE3MZ/terminull/$lib"
)

func main() {
	fmt.Println("$lib OK")
}
EOF

	echo "==> $lib"
	(
		cd "$dir"
		go mod tidy
		imports=$(go list -f '{{join .Imports "\n"}}' .)
		count=$(printf '%s\n' "$imports" | grep -c "^github.com/CHE3MZ/terminull/" || true)
		if [ "$count" -ne 1 ]; then
			echo "$lib imports $count terminull packages, want exactly 1" >&2
			exit 1
		fi
		go build -o "standalone$EXE" .
		out=$(./standalone$EXE)
		if [ "$out" != "$lib OK" ]; then
			echo "$lib ran as $out, want $lib OK" >&2
			exit 1
		fi
	)
done

echo "standalone OK"
