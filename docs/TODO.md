# TODO

> Build order for terminull. One checkbox = one small commit.
> Rule: code in pieces per item, commit with conventional message, push,
> then check `gh run list` green before next piece.
> Local first: run `scripts/ops/*` before pushing, do not leave it all to CI.
> `scripts/ops/` is reserved for ops scripts only. New build scripts go in `scripts/` as `.sh`, build into `dist/`.
> Phases 1–13 record the v1 folder names; the tree was renamed to term* right after 14.1.

# phase 0 — foundation:

- [x] 0.1 pin go directive + add direct deps to go.mod (cobra, koanf, survey, color, isatty, colorable, runewidth, spinner, progressbar, asciigraph, tablewriter, yaml, sys, text, humanize, shlex)
- [x] 0.2 add `.golangci.yml` strict presets (vet, staticcheck, errcheck, govulncheck, gosec)
- [x] 0.3 expand `Taskfile.yml` (lint, test, build, smoke, ops tasks calling scripts/)
- [x] 0.4 add `.gitattributes` for sh LF + go files (windows safe)

# phase 1 — scripts (sh, dist/):

- [x] 1.1 `scripts/build.sh` builds all packages + terminull CLI into `dist/`
- [x] 1.2 `scripts/test.sh` runs `go test ./...` with logs/ output
- [x] 1.3 `scripts/smoke.sh` end-to-end generator smoke (scaffold, tidy, build, run)
- [x] 1.4 verify existing `scripts/ops/go-bugcheck.sh` + `go-seccheck.sh` pass

# phase 2 — workflows (.github/workflows, gh cli green):

- [x] 2.1 `ci.yml` on push/PR: setup-go, cache, `go build ./...`, `go test ./...`, upload logs/
- [x] 2.2 `lint.yml`: golangci-lint + gosec + govulncheck via ops scripts
- [x] 2.3 `build-matrix.yml`: cross build linux/mac/windows amd64+arm64 into `dist/`, upload artifacts
- [x] 2.4 `test-matrix.yml`: cross test ubuntu/macos/windows-latest, `go test -v ./...`
- [x] 2.5 `dogfood.yml`: scaffold, tidy, build, run generated project on 3 OSes with output artifacts
- [ ] 2.6 verify with `gh workflow list` + `gh run watch` all green

# phase 3 — code (reference: cirrus-cli/cmd/cirrus/main.go, internal/commands/root.go, logs/output.go, echelon/logger.go):

- [x] 3.1 `core/` signals context, ExitCodeError, version from build info, slog LevelVar
- [x] 3.2 `core/` tests: exit code mapping, signal cancel, version format
- [x] 3.3 `style/` tty detect (isatty), colorable writers, NO_COLOR + TERM=dumb + CI, runewidth helper
- [x] 3.4 `style/` tests: forced color/no-color, dumb, piped
- [x] 3.4b `style/` Faint/FaintRGB half-luminance dimmer, depth matrix tests, Muted stays sgr faint
- [x] 3.4c `style/` owned Attribute aliases (FgRed..FgHiWhite) + engine parity test, signatures show style types
- [x] 3.4d `style/` Italic/Underline/Strikethrough with exact-code tests, dedicated resets, no blink by design
- [x] 3.5 `render/` fork of echelon as `CHE3MZ/terminull/render`: logger, Scoped, Simple/Interactive/Stub, Emoji vs SymbolsOnly, keep MIT header
- [x] 3.6 `render/` tests: scope tree, finish types, stub silent, simple golden
- [x] 3.7 `config/` koanf merge flag > env > file > default, strict yaml KnownFields, XDG + home lookup
- [x] 3.8 `config/` tests: precedence, unknown key fails, missing file ok
- [x] 3.9 `app/` cobra wrapper: New, Add, global flags (--config --output --no-color --verbose --debug), PersistentPreRunE order, ctx build, SilenceUsage/Errors
- [x] 3.9b `app/` owned facade: FlagSet embeds pflag, PositionalArgs validators, ShellCompDirective + CompletionFunc, engine mapping with directive parity test
- [x] 3.10 `app/` tests: help/version/completion out-only, unknown flag exit 2
- [x] 3.11 `prompt/` survey wrapper: Select/MultiSelect/Confirm/Input/Password, WithStdio, non-tty flag-required error, pause render contract
- [x] 3.12 `prompt/` tests: piped stdio select, non-tty error message
- [x] 3.13 `progress/` Progresser wrapping spinner + progressbar on single render writer, ascii fallback
- [x] 3.14 `progress/` tests: spinner start/stop, bar incr, no garble smoke
- [x] 3.15 `output/` table (tablewriter ascii), json/yaml stable keys, asciigraph, humanize, shlex
- [x] 3.16 `output/` tests: table vs json vs yaml golden, wide letters, piped
- [x] 3.17 `static/` appname banner, help, version, completion scripts, out-only fast exit on ctrl+c
- [x] 3.18 `static/` tests: help/version golden, completion non-empty

# phase 4 — dogfood generator CLI (replaces demo):

- [x] 4.1 drop demo leftovers (cmd/demo, internal/demo, tests/demo_test)
- [x] 4.2 `internal/generator/` templates + Generate + validation + gofmt + tests
- [x] 4.3 `internal/terminull/` + `cmd/terminull/` new command (flags, prompt, spinner tidy, table)
- [x] 4.4 `tests/` generator integration (help golden, scaffold files, bad name, non-interactive default)
- [x] 4.5 `scripts/smoke.sh` + dogfood workflow: scaffold, tidy, build, run on 3 OSes with output artifacts
- [ ] 4.3 docs: update `docs/index.md` mkdocs nav for TODO/spec/architecture when code lands

# phase 5 — release hygiene:

- [x] 5.1 `go mod tidy` + `govulncheck` clean, no charmbracelet in `go.sum`
- [x] 5.2 `scripts/ops/*` green locally, `gh run list` green on master
- [x] 5.3 tag via existing gitagger workflow (patch) after green
- [x] 5.4 goreleaser config + release workflow (tag-triggered, ldflags version)
- [x] 5.5 reviewdog workflow checked + pushed

# phase 6 — polish (gh-cli inspired, skills-assisted):

- [x] 6.1 ponytail-audit sweep: drop VersionLine, unify painting on Scheme, promptIO on StdIO, app on Format consts, cover Changed
- [x] 6.2 wire TerminalWidth into bars, EnsureConsole up front with render delegation
- [x] 6.3 security guidance pass: govulncheck clean, perms/exec/paths/scripts reviewed, no findings
- [x] 6.4 pager package with TTY gating and EPIPE tolerance
- [x] 6.5 retry helper and positional arg completion
- [x] 6.6 v1.1.0 tagged and released live (6 archives + checksums)

# phase 7 — self-contained forks (absorb deps):

- [x] 7.1 prompt line engine replaces survey (ctx cancel, numbered selects, no escape dependence)
- [x] 7.2 spinner + bar owned by progress (TTY gating, ASCII-safe, narrow fit)
- [x] 7.3 table + graph owned by output (ANSI-aware layout, verbatim pipes)
- [x] 7.4 go.mod slimmed to 19 direct + 8 indirect

# phase 8 — termtable library:

- [x] 8.1 render rows in five styles: bordered, rounded, ascii, markdown, borderless
- [x] 8.2 default bordered, unknown styles error, empty tables silent, ragged rows normalize
- [x] 8.3 output rerouted (Table/TableWith/TableStyle), style plumbed to termtable
- [x] 8.4 engine goldens: borders, align, wrap, cjk, merge, colspan, footers, fill width, padding, multiline, truncation, print errors
- [x] 8.5 engines vendored privately under termtable/internal, licenses and attribution kept, no upstream names in the public api
- [x] 8.6 examples/termtable renders one table in every style
- [x] 8.7 renderer rewritten in-house (termtable/internal/table), third-party engines and their license files removed, demo output byte-identical

# phase 9 — box rendering:

- [x] 9.1 `style/` ANSI-aware widths: VisibleWidth + EscapeLen (csi/osc) + RuneWidth, exact tests
- [x] 9.2 `box/` framed rendering: rounded hard double ascii none, title, padding, fixed width wrap, frame color, goldens

# phase 10 — repl line editor:

- [x] 10.1 `prompt/` Editor on liner v1.2.3: per-call console lend, ctx cancel, eof, piped fallback, history read/write/append/clear, completer, tab styles, beep, multiline, tests

# phase 11 — image rendering:

- [x] 11.1 `termimage/` facade: Auto/Kitty/ITerm2/Sixel/HalfBlock/ASCII backends, Options width/height in cells, String/Render with scheme, tmux passthrough for kitty/iterm2 with sixel left raw for tmux 3.4+
- [x] 11.2 encoders: kitty chunked graphics, iterm2 inline png (straight-alpha), sixel with adaptive median-cut palette, DEC 10x20 windows sizing, transparency as native alpha/unpainted cells on every tier, floyd-steinberg dither on sixel + 256-color half-blocks + ascii, half-block truecolor/256, ascii ramp, box+bilinear scaler
- [x] 11.3 detection: env hints, DA1 sixel probe, per-os cell pixel query, 8x16 default
- [x] 11.4 tests: golden protocol bytes, chunk reassembly, palette, detection env, sizing; demo verified piped

# phase 12 — style consistency:

- [x] 12.1 `prompt/` bullets paint via style.Scheme under the style.Color gate; direct fatih/color and go-isatty use removed
- [x] 12.2 `progress/` WithScheme on spinner and bar: cyan frame and fill, zero scheme stays plain
- [x] 12.3 `termtable/` borderless measures visible width, so styled cells align on all five styles

# phase 13 — depth and table style polish:

- [x] 13.1 `style.ColorDepth` single detector: case-insensitive COLORTERM, real windows consoles 24-bit under VT, 256color terms, else 16; Faint/FaintRGB consume it
- [x] 13.2 `termimage` consumes style.ColorDepth; local depth detector, sentinel consts, and console probe removed; half-block tiers unchanged
- [x] 13.3 `output.TableStyleWith` styles headers on any termtable style, TableWith delegates to it

# phase 14 — v2 owned library workspace (plan: docs/v2-plan.md):

- [x] 14.1 `termcore/` becomes the base: width/ansi/capability primitives move in; isatty, colorable, runewidth (+tables), uax29 graphemes, shlex forked into internal; termrender/node drops x/text; termtable + termimage consume termcore width/depth
- [x] 14.2 `termstyle/` forks fatih/color into internal, exact sgr parity, no color mod left
- [x] 14.3 small forks: retry-go → `termretry/internal`, humanize subset → `termoutput/internal`
- [ ] 14.4 `termprompt/` forks liner into internal lineedit, editor parity incl. windows raw mode
- [ ] 14.5 `termconfig/` forks koanf + yaml parser + providers + mapstructure, yaml.v3 direct, go.yaml.in dropped
- [ ] 14.6 `termapp/` forks cobra + pflag into internal cli, facade + completion directive parity
- [x] 14.7 rename packages to term* (termcore, termstyle, termbox, termtable, termimage, termoutput, termprogress, termprompt, termrender, termpager, termstatic, termretry, termconfig, termapp); pulled forward right after 14.1
- [ ] 14.8 final integration review: parity harness per fork, standalone-build CI per library, dependency-law test, windows/mac/linux matrix, tidy to yaml.v3 + x/sys + x/term
- [ ] 14.9 THIRD_PARTY.md provenance + `task upstream` advisory sweep
- [ ] 14.10 tag v2.0.0 after all gates (approval required)

# next (backlog, not scheduled):

- [ ] prompt live verification on a real console (conhost vs terminal matrix)
- [x] termimage live verification: wt sixel default, ghostty kitty default, terminal.app + alacritty half-block default, wezterm now routes kitty to iterm2 (conpty mangles kitty escapes, imgcat needs explicit compensation), iterm2 and kitty full matrix on mac, kitty id=0 multi-image fix confirmed
- [ ] termimage live verification inside tmux: kitty/iterm2 passthrough with allow-passthrough on, sixel raw on tmux 3.4+ (behavior set from tmux source and issues, not yet run live)
- [ ] docs nav for mkdocs once index content lands (see 4.3)
- [ ] man pages via cobra doc generation
- [ ] update notifier with state file throttle
- [ ] brew scoop choco distribution extras
- [ ] config file watching and toml codec
- [ ] evaluate plain-line prompt fallback if a console family misbehaves
