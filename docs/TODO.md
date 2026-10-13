# TODO

> Build order for terminull. One checkbox = one small commit.
> Rule: code in pieces per item, commit with conventional message, push,
> then check `gh run list` green before next piece.
> Local first: run `scripts/ops/*` before pushing, do not leave it all to CI.
> `scripts/ops/` is reserved for ops scripts only. New build scripts go in `scripts/` as `.sh`, build into `dist/`.

# phase 0 — foundation:

- [ ] 0.1 pin go directive + add direct deps to go.mod (cobra, koanf, survey, color, isatty, colorable, runewidth, spinner, progressbar, asciigraph, tablewriter, yaml, sync, sys, text, humanize, shlex, retry-go)
- [ ] 0.2 add `.golangci.yml` + `golangci.json` strict presets (vet, staticcheck, errcheck, govulncheck, gosec)
- [ ] 0.3 expand `Taskfile.yml` (lint, test, build, demo, ops tasks calling scripts/)
- [ ] 0.4 add `.gitattributes` for sh LF + go files (windows safe)

# phase 1 — scripts (sh, dist/):

- [ ] 1.1 `scripts/build.sh` builds all packages + demo-app into `dist/` (GOOS/GOARCH matrix locally)
- [ ] 1.2 `scripts/test.sh` runs `go test ./...` with logs/ output
- [ ] 1.3 `scripts/demo.sh` runs demo-app non-interactive (`--yes --output table|json`) for smoke
- [ ] 1.4 verify existing `scripts/ops/go-bugcheck.sh` + `go-seccheck.sh` pass on empty repo

# phase 2 — workflows (.github/workflows, gh cli green):

- [ ] 2.1 `ci.yml` on push/PR: setup-go, cache, `go build ./...`, `go test ./...`, upload logs/
- [ ] 2.2 `lint.yml`: golangci-lint + reviewdog (uses existing `.reviewdog.yml`), gosec, govulncheck, staticcheck, errcheck, vet
- [ ] 2.3 `build-matrix.yml`: cross build linux/mac/windows amd64+arm64 into `dist/`, upload artifacts
- [ ] 2.4 `test-matrix.yml`: cross test ubuntu/macos/windows-latest, `go test -v ./...`
- [ ] 2.5 `demo.yml`: run demo-app on all 3 OS, capture text output as artifact for visual check
- [ ] 2.6 verify with `gh workflow list` + `gh run watch` all green

# phase 3 — code (reference: cirrus-cli/cmd/cirrus/main.go, internal/commands/root.go, logs/output.go, echelon/logger.go):

- [x] 3.1 `core/` signals context, ExitCodeError, version from build info, slog LevelVar
- [x] 3.2 `core/` tests: exit code mapping, signal cancel, version format
- [x] 3.3 `style/` tty detect (isatty), colorable writers, NO_COLOR + TERM=dumb + CI, runewidth helper
- [x] 3.4 `style/` tests: forced color/no-color, dumb, piped
- [x] 3.5 `render/` fork of echelon as `CHE3MZ/terminull/render`: logger, Scoped, Simple/Interactive/Stub, Emoji vs SymbolsOnly, keep MIT header
- [x] 3.6 `render/` tests: scope tree, finish types, stub silent, simple golden
- [x] 3.7 `config/` koanf merge flag > env > file > default, strict yaml KnownFields, XDG + home lookup
- [x] 3.8 `config/` tests: precedence, unknown key fails, missing file ok
- [x] 3.9 `app/` cobra wrapper: New, Add, global flags (--config --output --no-color --verbose --debug), PersistentPreRunE order, ctx build, SilenceUsage/Errors
- [x] 3.10 `app/` tests: help/version/completion out-only, unknown flag exit 2
- [x] 3.11 `prompt/` survey wrapper: Select/MultiSelect/Confirm/Input/Password, WithStdio, non-tty flag-required error, pause render contract
- [x] 3.12 `prompt/` tests: piped stdio select, non-tty error message
- [ ] 3.13 `progress/` Progresser wrapping spinner + progressbar on single render writer, ascii fallback
- [ ] 3.14 `progress/` tests: spinner start/stop, bar incr, no garble smoke
- [ ] 3.15 `output/` table (tablewriter ascii), json/yaml stable keys, asciigraph, humanize, shlex
- [ ] 3.16 `output/` tests: table vs json vs yaml golden, wide letters, piped
- [ ] 3.17 `static/` appname banner, help, version, completion scripts, out-only fast exit on ctrl+c
- [ ] 3.18 `static/` tests: help/version golden, completion non-empty

# phase 4 — demo + integration:

- [ ] 4.1 `cmd/demo/` demo-app using everything: deploy cmd with --env flag + select fallback, spinner + bar + tree logs, table/json out
- [ ] 4.2 `tests/` integration: demo --help/--version golden, demo --output json golden, non-tty select error
- [ ] 4.3 docs: update `docs/index.md` mkdocs nav for TODO/spec/architecture when code lands

# phase 5 — release hygiene:

- [ ] 5.1 `go mod tidy` + `govulncheck` clean, no charmbracelet in `go.sum`
- [ ] 5.2 `scripts/ops/*` green locally, `gh run list` green on master
- [ ] 5.3 tag via existing gitagger workflow (patch) after green
