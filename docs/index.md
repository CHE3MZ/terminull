# terminull

Terminull is a workspace of small Go libraries for building command line
apps on Linux, macOS, and Windows. Pick the parts you need; every library
works standalone, and `termapp` wires them together when you want the
whole stack.

- consistent terminal output: tables, boxes, images, progress, prompts
- layered config: flags, environment, YAML/TOML files, defaults
- owned engines: no cobra, color, prompt, or config framework leaks
- plain fallbacks for pipes, CI, dumb terminals, and legacy consoles

## Quick start

```sh
go install github.com/CHE3MZ/terminull/v2/cmd/terminull@latest
terminull new myapp
cd myapp
go run ./cmd/myapp --help
```

The generated app uses `termapp`, `termconfig`, `termrender`,
`termprompt`, `termprogress`, `termoutput`, and `termstyle` together.

## Libraries

| package | what it does |
| --- | --- |
| `termcore` | widths, ANSI, terminal capability, signals, exit codes |
| `termstyle` | colors, attributes, schemes, terminal depth |
| `termbox` | framed boxes with wrapping and titles |
| `termtable` | five table styles with ANSI-aware layout |
| `termimage` | kitty, iTerm2, sixel, half-block, and ASCII images |
| `termoutput` | tables, JSON, YAML, graphs, humanized values |
| `termprogress` | spinners and bars with legacy fallbacks |
| `termprompt` | inputs, selects, confirms, editors |
| `termrender` | one-writer logging and live trees |
| `termpager` | paging with TTY and EPIPE handling |
| `termstatic` | banners, help, version, completion output |
| `termretry` | retry with capped attempts and backoff |
| `termconfig` | layered YAML/TOML, env, flag configuration and polling watch |
| `termapp` | CLI framework combining the libraries |

## Documentation

- [Architecture](architecture.md)
- [Specification](spec.md)
- [Dependencies](dependencies.md)
- [Workflow](workflow.md)
- [TODO](TODO.md)
