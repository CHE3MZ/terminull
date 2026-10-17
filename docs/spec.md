# goal:

terminull is a box of parts for building command line apps on linux, mac, windows.
apps look and behave the same everywhere, especially text rendering.
apps never touch cobra, color, prompts directly, only through terminull.
wrapping matters more than features, underlying mods can change, our names do not.

# non-goals:

no full screen apps, no ssh, no copy, no metrics in core.
no charmbracelet mods anywhere.
no new renderer from scratch, render is a fork of echelon.
no breaking public names without a major version.

# compat:

go 1.26 and up.
linux amd64 arm64, mac amd64 arm64, windows amd64 arm64.
windows conhost with cmd, windows terminal with pwsh, mac terminal, linux xterm.
dumb terminal, ci mode, piped out, no color must all still work.
cobra version is pinned once by terminull, apps do not pin it.

# structure:

app/ ( cobra wrapper , root and subcommands )
config/ ( koanf + yaml , flags env file merge )
render/ ( fork of echelon , single writer logs )
style/ ( color + isatty + colorable + runewidth , windows safe )
prompt/ ( line engine , selects and dialogs , ctx cancel , repl editor )
progress/ ( owned spinner + bar , tty gated )
output/ ( tables + json yaml + graph , humanize shlex )
static/ ( appname banner , out only , no input )
core/ ( signals + exit codes + version , slog )
retry/ ( backoff retries , attempts delay veto )
pager/ ( less paging , tty gated , epipe safe )
table/ ( aqua default + rodaine styles )
box/ ( framed boxes , ansi aware widths )
termimage/ ( image rendering , kitty + iterm2 + sixel + ansi fallback )
cmd/terminull ( dogfood generator cli )

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
retry/ ( public , importable as CHE3MZ/terminull/retry )
pager/ ( public , importable as CHE3MZ/terminull/pager )
table/ ( public , importable as CHE3MZ/terminull/table )
box/ ( public , importable as CHE3MZ/terminull/box )
termimage/ ( public , importable as CHE3MZ/terminull/termimage )
cmd/terminull ( dogfood binary , not importable )
internal/ ( generator + terminull cli wiring , never imported by apps )
tests/ ( integration , runs on windows mac linux )
scripts/ ( build test smoke + ops , all sh )
docs/ ( index + dependencies + architecture + spec + TODO )

# lifecycle:

how every app runs from start to finish:
1. main makes a stop context for ctrl+c and sigterm.
2. app.New creates root cobra command with SilenceUsage true and SilenceErrors true.
3. app adds global flags once: config, output, no-color, verbose, debug.
4. outputs default to process streams, tests redirect with SetOut SetErr SetIn.
5. PersistentPreRunE runs in fixed order: debug switch, console ansi setup, output validation, renderer plus logger, config load.
6. cobra parses flags and finds subcommand, help and version bypass live render.
7. RunE wrapper builds ctx with context, config, logger, io, output mode, flag getters.
8. app code runs with ctx only, returns error or nil.
9. core maps error to exit code, stops render, flushes logs, exits.

rules:
only one writer owns the terminal per phase, never two at once.
prompt pauses render before asking and resumes after via ctx PauseDrawing and ResumeDrawing.
spinners and bars do the same: pause, draw, resume, never beside live tree output.
logger drains before finalize via Sync, stops join the draw loop before tables print.
static help version completion always use simple plain output even on tty.
every prompt has a flag fallback, non terminal must still work with flags alone.
validate names and flags before banners and logs, so usage errors print clean and fast.

# app:

app is the boss, it owns cobra inside and never exposes it except one Raw escape.
app holds name, version, root command, subcommands.
app owns global flags: --config, --output table json yaml, --no-color, --verbose, --debug.
globals carry no shorthands, so subcommands own -f -o -e -v freely.
app uses stock cobra help, version, completion output.
apps declare Command with use, short, flags, run with ctx, plus Args and ValidArgs completion.
flag callbacks take app.FlagSet, arg validation takes app validators, completions take app directives.
Raw and Unwrap stay as narrow escape hatches, user code never imports cobra or pflag.
WithSymbolsOnly pins ascii spinners and statuses on any terminal.

# config:

config loads in order: flags win, then env, then file, then defaults.
file lookup: explicit --config, then home dotfile, then user config dir, then local files.
env prefix is upper app name plus underscore, example MYAPP_ENV.
yaml decode is strict via mapstructure ErrorUnused, unknown keys fail loudly.
config never prompts, it only loads and merges, prompt decides what is missing.
koanf is the merge engine, yaml.v3 is the file codec.

# render:

render is CHE3MZ/terminull/render, internal fork of cirruslabs/echelon, MIT kept.
api stays compatible: NewLogger, Scoped, Infof Debugf Warnf Errorf, AsWriter, Finish.
single writer with one draw goroutine, all output goes through it.
modes: interactive live tree on tty, simple lines on ci pipe dumb.
emoji where the terminal renders it including windows terminal, ascii SymbolsOnly elsewhere.
ci decorators for travis github teamcity are kept from echelon pattern.
setup runs synchronously before handlers, drawing starts alongside, stop joins the loop.
deps are x/sys, x/text plus style, no heavy mods.

# style:

style decides color or no color and wraps every writer.
detect order: --no-color flag, NO_COLOR env, TERM dumb, not a tty, CI true.
ColorFor answers ansi-or-plain for any writer, pipes and buffers stay plain.
EnsureConsole enables ansi on windows consoles up front, best effort elsewhere.
Scheme paints named styles and icons, plain when disabled.
Scheme.Faint dims a named color by half luminance, FaintRGB dims any rgb triple.
Bold Italic Underline Strikethrough emit exact sgr with dedicated resets.
style foreground and decoration constants mirror the engine, user code never imports fatih color.
Muted dims the current color through sgr faint, Faint is the deterministic form.
16 color terms downshift bright colors or snap rgb to the nearest base.
TerminalWidth reports console columns or 80 fallback, used to fit bars.
all widths use runewidth plus x/text EastAsianWide, cjk never breaks tables.
piped out like app list piped to grep must be plain with no codes.

# prompt:

prompt prints a bullet question and reads one line, no cursor tracking, no re-renders.
selects are numbered lists, confirms are y/n, passwords use the os no-echo read.
every call takes ctx, ctrl+c cancels the read and maps to exit code 130.
presets return validated without prompting, so flags always win.
prompt only runs when isatty true, else it returns a clear flag required error naming the flag.
prompt/ Editor opts into liner editing for repl input: history, ctrl+r search, tab completion, multiline.
editor lends the console only for the duration of a Line call and hands it back, pipes and dumb terminals degrade to plain reads.
editor ctrl+c returns context canceled, ctrl+d and piped eof return io eof.
prompt tests drive real reads over pipes, never real tty.

# progress:

progress owns its spinner and bar, no widget mods remain.
spinners animate only on terminal files, quiet no-op on pipes and files.
spinner frames are ascii or braille dots by caller choice, ascii safe everywhere.
bars render every step as plain lines, so ci logs keep full progress.
bars shrink on narrow terminals instead of wrapping.
multiple bars and logs share one writer but never draw at once, callers sequence them.

# output:

output prints app results as table, json, yaml based on --output flag, default table.
tables come from the merged table lib, aqua default, rodaine for script-friendly.
styled headers apply only through the ansi-aware engine, pipes stay verbatim.
graphs are owned ascii point charts with labeled axis, no color codes inside.
sizes use humanize bytes, command splitting uses shlex.
output never reads input, it is out only, same static rules as static.
json and yaml must be stable key order for tests and pipes.

# table:

table merges aquasecurity and rodaine engines behind one facade.
StyleAqua is bordered unicode, the default look, with header alignment and footers.
StyleRodaine is borderless padded columns for script-friendly output.
new styles arrive as enum values, unknown styles error instead of guessing.
empty tables render nothing, ragged rows normalize to header width.

# box:

box renders framed boxes around arbitrary text, no third party code.
styles are rounded, hard, double, ascii, none, unknown styles render rounded.
titles sit in the top border and truncate with an ellipsis when the box is narrow.
padding is inner horizontal spaces, width sets the outer size and wraps content.
content may carry ansi, borders measure visible glyphs so lines stay aligned.
open styles reset at each line end and resume on the next wrapped line, borders stay clean.
cjk counts double via runewidth, combining marks stay with their base rune.
frame color goes through scheme, a disabled scheme prints plain bytes.

# termimage:

termimage renders images for terminals with no new mods, stdlib only.
backends are kitty graphics, iterm2 inline png, sixel, and ansi half-blocks or ascii.
auto detection reads env hints first, then probes the terminal for sixel with a DA1 query.
kitty graphics cover kitty, ghostty, and wezterm, iterm2 covers iterm, sixel covers windows terminal, mlterm, foot, and contour.
width and height are cell counts, zero width uses the terminal width and keeps aspect.
sixel builds an adaptive palette from the image and dithers only on request, kitty and iterm2 send exact pixels.
windows consoles size sixel rasters in DEC 10x20 cell units so aspect holds on any font, posix terminals use real cell pixels.
transparency survives by default: kitty and iterm2 send native alpha, sixel asks the terminal to keep its background and leaves those pixels unpainted, half-blocks and ascii flatten over the background color, and a background color with alpha flattens sixel too.
half-blocks need 256 colors, color off or 16 colors degrade to ascii, pipes never carry escapes.
tmux is detected and graphics sequences route through a passthrough wrapper.
file decode is stdlib png jpeg gif, scale is box average for shrink and bilinear for grow.

# retry:

retry reruns fallible ops with capped attempts and backoff.
defaults are 3 attempts and 1 second base, veto with RetryIf for permanent errors.
used for go mod tidy over flaky networks.

# pager:

pager pages long output through less on capable terminals.
precedence is explicit flag, then PAGER env, then less, cat declines.
dumb terminals, pipes, and missing binaries pass writers through untouched.
LESS defaults to FRX unless the caller sets it.
quitting early breaks writes with EPIPE, check with Closed and treat as success.

# generator:

internal/generator scaffolds go.mod, main package, sample commands, readme.
names validate up front, reserved windows names rejected, existing dirs must be empty.
generated go files pass through go/format, broken templates fail generation.
internal/terminull wires the new command: flags, prompts, spinner tidy with retries, tables.
smoke.sh and dogfood CI scaffold, tidy, build, and run a project on all platforms.

# static:

static prints framework things that are the same for every app: appname banner and version line.
banner paints green when capable through ColorFor, plain otherwise.
static never asks input, never starts spinner or bar, always plain through style.
static works on dumb, ci, piped, no color without change.
static finishes fast and exits cleanly on ctrl+c.
app specific lists like list users belong in output, not static, but follow the same out only rules.

# core:

core holds signals, exit codes, version, slog logging.
signals: SIGINT SIGTERM cancel the app context, render stops, prompt aborts, exit code 130 on ctrl+c.
exit codes: 0 ok, 1 error, 2 usage, 130 interrupted, custom via ExitCodeError like cirrus helpers.
version: semver from build info plus git tag, --version uses static plain output.
logging: slog with LevelVar, --verbose sets debug, --debug sets debug.
Level wires into app slog handlers with core.Level.

# errors:

all app errors go to stderr once, never twice.
usage errors exit 2 with the error on stderr, no usage dump.
interrupt exits 130 with no stack.
plain errors exit 1, missing flags in non terminals exit 1 with flag named.
unknown config keys exit 2 with key name.
Sentry and otel are not in core, apps add them as contrib if needed.

# testing:

unit tests live beside code on every package, integration in tests/.
tests run on windows, mac, linux, plus piped and non tty paths everywhere.
golden files for help, version, table, json, yaml, ascii vs emoji tree.
prompt tests drive real reads over pipes, never real tty.
render tests assert join silence after stop and no garble under race detector.
no network in tests except dogfood CI which tidies for real.

# versioning:

semver, public folder names and function names do not break without major bump.
underlying mods like cobra, koanf, fatih can be upgraded without breaking apps because they are wrapped.
owned code (render, prompt, progress, output, table) evolves in place with tests.
render fork keeps echelon api names so cirrus pattern still ports.
deprecations warn for one minor before removal.
