# terminull v2 — owned library workspace plan

> Phase 14 planning spec. Nothing here changes the v1 tree until its phase lands.
> Todo items live in TODO.md as 14.x; this file is the contract.

# decisions:

- module path stays github.com/CHE3MZ/terminull; public packages use term* names (renamed right after 14.1, ahead of the planned 14.7).
- every non-exempt dependency is forked literally into the component that uses it, owned by this repo.
- shared text and terminal primitives are forked once into termcore, never copied per component.
- every library must work standalone: importing it without terminull must give the full functionality of that component.
- external deps after v2: gopkg.in/yaml.v3, golang.org/x/sys, golang.org/x/term. yaml is a file codec, x/* is platform tier. nothing else stays.
- behavior is preserved byte for byte; this is a move, not a feature release.

# target roster:

| now | v2 | owns |
| --- | --- | --- |
| core | termcore | signals exit codes version slog + text widths ansi parse isatty console vt color depth terminal width shlex |
| style | termstyle | owned color engine fork, scheme paint faint rgb |
| box | termbox | framed boxes over termcore widths |
| termtable | termtable | five styles, owned renderer, termcore widths |
| termimage | termimage | kitty iterm2 sixel half-block ascii |
| output | termoutput | tables json yaml graph, owned humanize subset |
| progress | termprogress | spinner + bar |
| prompt | termprompt | line engine, owned lineedit fork |
| render | termrender | echelon fork, single writer logs |
| pager | termpager | less paging |
| static | termstatic | banner help version completion |
| retry | termretry | owned retry engine fork |
| app | termapp | owned cli engine fork, root and subcommands |
| config | termconfig | owned merge engine fork, yaml codec |

not libraries: cmd/terminull (binary), internal/terminull (wiring), internal/generator (scaffolder), examples, tests, scripts.

# dependency law:

```
L0  termcore    → x/sys, x/term
L1  termstyle   → termcore
L2  termbox termtable termprogress termstatic termpager
    termprompt termrender termimage → termcore (+ termstyle where color is used)
L3  termoutput  → termtable termstyle termcore yaml.v3
    termconfig  → yaml.v3 (+ its internal fork)
    termretry   → stdlib (+ its internal fork)
L4  termapp     → all libraries below + its internal fork
    cmd/terminull, internal/* → termapp and friends
```

rules:

- imports point down only, no cycles.
- a library's public api may reference its own and lower-layer types only.
- only termcore asks the os about terminal facts: width, tty, depth, console mode.
- only termstyle builds text sgr; termcore parses ansi; termimage builds image protocol escapes only.
- only the owning library touches its internal fork.
- a standalone build per library is a CI gate, not a convention.

# fork policy:

- literal forks: upstream source is copied into the owning library's internal tree at a pinned version.
- provenance: THIRD_PARTY.md records repo, version, commit, license, and fork date for every fork.
- minimal modification: only integration seams change (termcore width/ansi, console setup, error wrapping, removal of unused surface). engine logic and upstream tests stay recognizable so security fixes cherry-pick cleanly.
- forked engines never appear in public apis; users see term* types only.
- no cosmetic churn. upstream fixes arrive only for a bug we hit or a capability we document.
- upstream watch: task upstream checks the recorded origins for newer releases and advisories. renamed modules are invisible to osv/govulncheck, so this sweep replaces that signal.
- exempt by design: gopkg.in/yaml.v3 and golang.org/x/* stay external. test-only modules stay external or get dropped.
- history: render keeps echelon api names and attribution so the cirrus pattern still ports. termtable reclaimed its renderer in v1 with byte-identical output; v2 generalizes that play to the rest of the stack.

# fork map:

| upstream | version | destination | license |
| --- | --- | --- | --- |
| mattn/go-isatty | v0.0.24 | termcore/internal/isatty | MIT |
| mattn/go-colorable | v0.1.16 | termcore/internal/colorable | MIT |
| mattn/go-runewidth | v0.0.31 | termcore/internal/width (+ tables and regen script) | MIT |
| clipperhouse/uax29/v2 | v2.7.0 | termcore/internal/uax29/graphemes (width's grapheme engine) | MIT |
| google/shlex | e7afc7fbc510 | termcore/internal/shlex | Apache-2.0 |
| fatih/color | v1.19.0 | termstyle/internal/color | MIT |
| avast/retry-go/v4 | v4.7.0 | termretry/internal/retry | MIT |
| dustin/go-humanize | v1.0.1 | termoutput/internal/humanize | MIT |
| peterh/liner | v1.2.3 | termprompt/internal/lineedit | MIT |
| knadh/koanf/v2 | v2.3.8 | termconfig/internal/koanf | MIT |
| knadh/koanf/parsers/yaml | v1.1.1 | termconfig/internal/koanf/yaml | MIT |
| knadh/koanf/providers/{confmap,env/v2,file} | v1.0.1 v2.0.2 v1.2.1 | termconfig/internal/koanf | MIT |
| go-viper/mapstructure/v2 | v2.4.0 | termconfig/internal/mapstructure | MIT |
| spf13/cobra | v1.10.2 | termapp/internal/cli | Apache-2.0 |
| spf13/pflag | v1.0.9 | termapp/internal/cli/pflag | BSD-3-Clause |
| gopkg.in/yaml.v3 | v3.0.1 | stays external | MIT/Apache-2.0 |
| golang.org/x/{sys,term} | v0.49.0 v0.44.0 | stays external | BSD-3-Clause |

# what leaves go.mod:

cobra, pflag, koanf (all), mapstructure, go.yaml.in/yaml/v3, liner, retry-go, humanize, shlex, fatih/color, isatty, colorable, runewidth, uax29, x/text, plus stale indirects fsnotify, mousetrap, testify.

# phases:

## 14.1 termcore extraction

- move ansi parse and visible width from termstyle to termcore; keep termstyle aliases for v1 callers.
- fork isatty, colorable, runewidth (+unicode tables and the uax29 grapheme engine it uses), shlex into termcore/internal.
- move capability detection to termcore: Color, ColorFor, ColorDepth, IsTerminal, EnsureConsole, TerminalWidth; termstyle aliases remain.
- switch termrender/renderers/internal/node from x/text/width to termcore width; x/text leaves go.mod.
- switch termtable/internal/table and termimage to termcore width and depth.
- gate: full suite green on windows, demo byte-identical, go.mod drops shlex and x/text; isatty/colorable leave in 14.2 and runewidth/uax29 in 14.4 with their last consumers.

## 14.2 termstyle fork

- fork fatih/color as termstyle/internal/color; Scheme, Faint, FaintRGB, attributes stay on top.
- gate: exact sgr goldens unchanged, no fatih import remains, go.mod drops color, isatty, colorable.

## 14.3 small forks

- fork retry-go as termretry/internal/retry; fork the humanize subset as termoutput/internal/humanize.
- gate: retry timing tests and output format goldens unchanged, go.mod drops two modules.

## 14.4 termprompt lineedit fork

- fork liner as termprompt/internal/lineedit; Editor keeps history, ctrl+r, completion, multiline, raw-mode behavior.
- gate: editor tests on all three oss, piped and non-tty paths, go.mod drops liner, runewidth, uax29.

## 14.5 termconfig fork

- fork koanf core, yaml parser, confmap/env/file providers, mapstructure into termconfig/internal; use yaml.v3 directly.
- gate: precedence and strict-key tests unchanged, go.yaml.in leaves go.mod.

## 14.6 termapp cli fork

- fork cobra and pflag as termapp/internal/cli; trim to the surface termapp exposes; keep Raw/Unwrap escape hatches.
- gate: help, version, completion directive parity, unknown-flag exit 2 unchanged, go.mod drops two modules.

## 14.7 renames (landed right after 14.1)

- rename every public package to its term* name; update imports, examples, docs.
- gate: build and full suite green, goldens unchanged, no v1 import path remains.

## 14.8 final integration review

- parity harness per forked engine (snapshot or golden zero-diffs, termtable precedent).
- standalone-build CI job: a throwaway module importing exactly one term* library builds and runs.
- dependency-law test: go list -deps asserts no upward edges and no unexpected externals.
- windows conhost + windows terminal, mac, linux matrix; dogfood workflow; ops and scan tasks.
- go.mod end state check: exactly yaml.v3, x/sys, x/term. go mod tidy clears the rest.

## 14.9 provenance and upstream watch

- THIRD_PARTY.md complete: repo, version, commit, license, fork date for every fork.
- task upstream implemented: checks recorded origins for newer releases and advisories, green locally and in ci.

## 14.10 release

- tag v2.0.0 after all gates; approval required before the tag push.

# risks and mitigations:

- windows console behavior in colorable/liner forks: dogfood + test matrix gates, behavior parity tests.
- unicode width tables are now ours: regen script and a documented upstream watch duty.
- renamed forks lose osv/govulncheck visibility: task upstream sweep plus THIRD_PARTY.md provenance.
- big-bang renames: kept last and isolated; behavior work never mixes with rename commits.

# non-goals:

- no new features, no api redesign, no module path change, no multi-repo split.
- no forking yaml or golang.org/x/*; no forking test-only modules.
- no public api surface for any forked engine.
