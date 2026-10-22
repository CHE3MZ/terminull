#!/usr/bin/env sh
# Check every pinned fork in .github/THIRD_PARTY.md for newer releases and advisories.
# Renamed forks are invisible to govulncheck and osv-scanner, so this table is
# the source of truth for the upstream watch.
# Usage: sh scripts/upstream.sh
set -eu

cd "$(dirname "$0")/.."

for tool in gh jq curl sort; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		echo "missing required tool: $tool" >&2
		exit 1
	fi
done

if ! printf 'v1.2.3\nv1.2.10\n' | sort -V >/dev/null 2>&1; then
	echo "GNU sort -V is required" >&2
	exit 1
fi

TMP=$(mktemp -d "${TMPDIR:-/tmp}/terminull-upstream.XXXXXX")
trap 'rm -rf "$TMP"' EXIT INT TERM

advisories=0
updates=0
forks=0

tags_for() {
	repo=$1
	cache="$TMP/tags-$(printf '%s' "$repo" | tr '/' '_')"
	if [ ! -f "$cache" ]; then
		if ! gh api "repos/$repo/tags" --paginate --jq '.[].name' >"$cache" 2>"$TMP/gh-error"; then
			echo "  warning: cannot list tags for $repo: $(cat "$TMP/gh-error")" >&2
			: >"$cache"
		fi
	fi
	printf '%s\n' "$cache"
}

repo_for() {
	case "$1" in
		github.com/knadh/koanf | github.com/knadh/koanf/*) echo "knadh/koanf" ;;
		github.com/avast/retry-go/v4) echo "avast/retry-go" ;;
		github.com/clipperhouse/uax29/v2) echo "clipperhouse/uax29" ;;
		github.com/go-viper/mapstructure/v2) echo "go-viper/mapstructure" ;;
		github.com/*) echo "${1#github.com/}" ;;
		*) echo "$1" ;;
	esac
}

tag_prefix_for() {
	case "$1" in
		github.com/knadh/koanf/maps) echo "maps/" ;;
		github.com/knadh/koanf/parsers/yaml) echo "parsers/yaml/" ;;
		github.com/knadh/koanf/providers/confmap) echo "providers/confmap/" ;;
		github.com/knadh/koanf/providers/env/v2) echo "providers/env/" ;;
		github.com/knadh/koanf/providers/file) echo "providers/file/" ;;
		*) echo "" ;;
	esac
}

is_version() {
	case "$1" in
		v[0-9]*) return 0 ;;
		*) return 1 ;;
	esac
}

while IFS='|' read -r _ fork upstream version _ _ _ _ _; do
	fork=$(printf '%s' "$fork" | sed 's/^ *//; s/ *$//')
	upstream=$(printf '%s' "$upstream" | sed 's/^ *//; s/ *$//')
	version=$(printf '%s' "$version" | sed 's/^ *//; s/ *$//' | awk '{print $1}')

	# Skip the header and the separator row.
	case "$fork" in
		fork | ---* | "") continue ;;
	esac

	forks=$((forks + 1))
	echo "==> $fork ($upstream $version)"

	repo=$(repo_for "$upstream")
	prefix=$(tag_prefix_for "$upstream")
	tags=$(tags_for "$repo")
	if [ -n "$prefix" ]; then
		latest=$(grep "^${prefix}v[0-9]" "$tags" | sed "s|^${prefix}||" | sort -V | tail -n1 || true)
	else
		latest=$(grep '^v[0-9]' "$tags" | sort -V | tail -n1 || true)
	fi

	if [ -z "$latest" ]; then
		echo "    latest: unknown (no release tags)"
	elif is_version "$version" && [ "$latest" != "$version" ]; then
		newer=$(printf '%s\n%s\n' "$version" "$latest" | sort -V | tail -n1)
		if [ "$newer" = "$latest" ]; then
			echo "    latest: $latest (update available)"
			updates=$((updates + 1))
		else
			echo "    latest: $latest"
		fi
	else
		echo "    latest: $latest"
	fi

	if is_version "$version"; then
		query=$(jq -n --arg module "$upstream" --arg version "$version" \
			'{package: {name: $module, ecosystem: "Go"}, version: $version}')
	else
		query=$(jq -n --arg module "$upstream" \
			'{package: {name: $module, ecosystem: "Go"}}')
		echo "    advisories: checking package (commit pin, no version filter)"
	fi

	if ! response=$(curl -fsS --retry 3 -X POST \
		-H 'Content-Type: application/json' \
		-d "$query" https://api.osv.dev/v1/query); then
		echo "    error: OSV query failed" >&2
		exit 1
	fi

	ids=$(printf '%s' "$response" | jq -r '.vulns[]?.id')
	if [ -z "$ids" ]; then
		if is_version "$version"; then
			echo "    advisories: none"
		fi
		continue
	fi

	echo "    advisories:"
	printf '%s\n' "$ids" | sed 's/^/      /'
	advisories=$((advisories + $(printf '%s\n' "$ids" | wc -l)))
done <.github/THIRD_PARTY.md

echo
if [ "$advisories" -gt 0 ]; then
	echo "upstream: $forks forks, $advisories advisory match(es), $updates update(s); review required"
	exit 1
fi
echo "upstream OK: $forks forks, 0 advisories, $updates update(s) available"
