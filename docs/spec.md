# goal:

terminull is a box of parts for building command line apps on linux, mac, windows.
apps look and behave the same everywhere, especially text rendering.
apps never touch cobra, color, spinner, prompts directly, only through terminull.
wrapping matters more than features, underlying mods can change, our names do not.

# non-goals:

no full screen apps, no ssh, no copy, no metrics in core.
no charmbracelet mods anywhere.
no new renderer from scratch, render is a fork of echelon.
no breaking public names without a major version.

# compat:

go 1.23 and up.
linux amd64 arm64, mac amd64 arm64, windows amd64 arm64.
windows conhost with cmd, windows terminal with pwsh, mac terminal, linux xterm.
dumb terminal, ci mode, piped out, no color must all still work.
cobra version is pinned once by terminull, apps do not pin it.

# structure:

app/ ( cobra wrapper , root and subcommands )
config/ ( koanf + yaml , flags env file merge )
render/ ( fork of echelon , single writer logs )
style/ ( color + isatty + colorable + runewidth , windows safe )
prompt/ ( survey , selects and dialogs )
progress/ ( spinner + progressbar , wrapped )
output/ ( table + graph + humanize , table json yaml )
static/ ( appname + help + version + completion , out only , no input )
core/ ( signals + exit codes + version , slog )

# layout:

app/ ( public , importable as CHE3MZ/terminull/app )
config/ ( public , importable as CHE3MZ/terminull/config )
render/ ( public , importable as CHE3MZ/terminull/render )
style/ ( public , importable as CHE3MZ/terminull/style )
prompt/ ( public , importable as CHE3MZ/terminull/prompt )
progress/ ( public , importable as CHE3MZ/terminull/progress )
output/ ( public , importable as CHE3MZ/terminull/output )
static/ ( public , importable as CHE3MZ/terminull/static )
core/ ( public , importable as CHE3MZ/terminull/core )
internal/ ( private , tooling only , never imported by apps )
tests/ ( cross platform matrix , windows mac linux )
scripts/ ( task helpers )
docs/ ( index + dependencies + architecture + spec )

# lifecycle:

how every app runs from start to finish:
1. main makes a stop context for ctrl+c and sigterm.
2. app.New creates root cobra command with SilenceUsage true and SilenceErrors true.
3. app adds global flags once: config, output, no-color, verbose, debug.
4. app redirects cobra out err in to colorable writers from style, never os.Stdout directly.
5. PersistentPreRunE runs in fixed order: core signals, style detect, config load, render start.
6. cobra parses flags and finds subcommand, help and version bypass live render.
7. RunE wrapper builds ctx with context, config, logger, prompt, progress, output, io.
8. app code runs with ctx only, returns error or nil.
9. core maps error to exit code, stops render, flushes logs, exits.

rules:
only render writes to the screen.
prompt pauses render before asking and resumes after.
static help version completion always use simple plain output even on tty.
every prompt has a flag fallback, non terminal must still work with flags alone.

# app:

app is the boss, it owns cobra inside and never exposes it except one Raw escape.
app holds name, version, root command, subcommands.
app owns global flags: --config -f, --output table json yaml, --no-color, --verbose, --debug.
app owns help template, version template, completion command, man generation.
apps declare Command with use, short, flags, run with ctx, no cobra types in app code.
ValidArgsFunction for completion is provided through app, same pattern as cirrus run task completion.

# config:

config loads in order: flags win, then env, then file, then defaults.
file lookup: --config flag, then XDG_CONFIG_HOME app yaml, then home dotfile, then local dotfile.
env prefix is upper app name plus underscore, example MYAPP_ENV.
yaml decode is strict with KnownFields true, unknown keys fail loudly.
config never prompts, it only loads and merges, prompt decides what is missing.
koanf is the merge engine, yaml.v3 is the file codec.

# render:

render is CHE3MZ/terminull/render, internal fork of cirruslabs/echelon, MIT kept.
api stays compatible: NewLogger, Scoped, Infof Debugf Warnf Errorf, AsWriter, Finish.
single writer with one draw goroutine, all output goes through it.
modes: interactive live tree on tty, simple lines on ci pipe dumb.
emoji mode on modern terminals, ascii SymbolsOnly mode on old windows conhost and dumb.
ci decorators for travis github teamcity are kept from echelon pattern.
render start returns Stop func, app defers Stop, ctrl+c also stops.
deps are only x/sys, x/text, colorable, runewidth, no heavy mods.

# style:

style decides color or no color and wraps every writer.
detect order: --no-color flag, NO_COLOR env, TERM dumb, not a tty, CI true.
windows conhost gets ENABLE_VIRTUAL_TERMINAL_PROCESSING via x/sys, fallback to plain if it fails.
all widths use runewidth plus x/text EastAsianWide, cjk never breaks tables.
no app uses fmt print or fatih color directly, all go through style and render.
piped out like app list piped to grep must be plain with no codes.

# prompt:

prompt uses AlecAivazis/survey v2 for Input, Password, Confirm, Select, MultiSelect.
prompt takes stdio from app io, testable with WithStdio, same as survey pattern.
prompt only runs when isatty true, else it returns a clear flag required error naming the flag.
prompt pauses render before drawing and resumes after, so menus never garble logs.
ctrl+c during prompt returns InterruptErr, app maps it to exit code 130 and stops cleanly.
survey is unmaintained but kept because no charm is allowed, wrapped so it can be swapped later.

# progress:

progress owns spinner from briandowns/spinner and bars from schollz/progressbar, wrapped behind one Progresser.
apps never create spinner or bar directly, they ask progress for one with title and total.
spinner frames have ascii fallback for server core and raster fonts.
bars use colorable writer from style, width follows terminal width.
multiple bars and logs share the same render writer, one draw loop, no flicker.
progress is paused during prompt and during static out.

# output:

output prints app results as table, json, yaml based on --output flag, default table.
tables use olekukonko/tablewriter with runewidth, ascii borders only for windows safe.
graphs use guptarohit/asciigraph plain ascii, no color codes inside.
sizes and durations use dustin/go-humanize.
command splitting uses google/shlex.
retry helper uses avast/retry-go v4.
output never reads input, it is out only, same static rules as static.
json and yaml must be stable key order for tests and pipes.

# static:

static prints framework things that are the same for every app: appname banner, help, version, shell completion.
static never asks input, never starts spinner or bar, always plain through style.
static uses simple renderer even on tty, no live tree.
static works on dumb, ci, piped, no color without change.
static finishes fast and exits cleanly on ctrl+c.
app specific lists like list users belong in output, not static, but follow the same out only rules.

# core:

core holds signals, exit codes, version, slog logging.
signals: SIGINT SIGTERM cancel the app context, render stops, prompt aborts, exit code 130 on ctrl+c.
exit codes: 0 ok, 1 error, 2 usage, 130 interrupted, custom via ExitCodeError like cirrus helpers.
version: semver from build info plus git tag, --version uses static plain output.
logging: slog with LevelVar, --verbose sets debug, --debug sets debug plus stack, lumberjack style rotation only for daemon apps, not core.
x/sync errgroup is the concurrency helper, no other runtime deps.

# errors:

all app errors go to stderr once through style, never twice.
usage errors exit 2 and show help for that command only.
interrupt exits 130 with no stack.
unknown config keys exit 2 with key name.
missing flag in non terminal exits 2 with flag name and example.
Sentry and otel are not in core, apps add them as contrib if needed.

# testing:

every part needs tests on windows, mac, linux, plus dumb, ci true, piped, no color.
tests live in tests/ with matrix names like windows-conhost, windows-terminal, mac-terminal, linux-xterm, dumb, ci.
golden files for help, version, table, json, yaml, ascii vs emoji tree.
prompt tests use WithStdio pipes, never real tty.
render tests assert no garble when spinner plus bar plus log run together.
survey and go-prompt are not added together in one test binary to avoid two input loops.

# versioning:

semver, public folder names and function names do not break without major bump.
underlying mods like cobra, survey, spinner can be upgraded without breaking apps because they are wrapped.
render fork keeps echelon api names so cirrus pattern still ports.
deprecations warn for one minor before removal.
