# goal:

terminull is a box of parts for building command line apps on linux, mac, windows.
apps look and behave the same everywhere, especially text rendering.
apps never touch cobra, color, prompts directly, only through terminull.
wrapping matters more than features, engines are frozen owned forks, our names and behavior do not change.

# non-goals:

no full screen apps, no ssh, no copy, no metrics in core.
no charmbracelet mods anywhere.
no third-party terminal mods; only yaml, toml, and golang.org/x/* stay external.
no new renderer from scratch, render is a fork of echelon.
no breaking public names without a major version.

# compat:

go 1.26 and up.
linux amd64 arm64, mac amd64 arm64, windows amd64 arm64.
windows conhost with cmd, windows terminal with pwsh, mac terminal, linux xterm.
dumb terminal, ci mode, piped out, no color must all still work.
the cli engine is forked and owned by terminull, apps never see it or pin it.

# structure (v2 target):

termapp/ ( owned cli engine fork , root and subcommands )
termconfig/ ( owned merge engine fork + yaml + toml , flags env file merge )
termrender/ ( fork of echelon , single writer logs )
termcore/ ( signals exit codes version slog + widths ansi tty depth console shlex )
termstyle/ ( owned color engine fork , windows safe )
termprompt/ ( line engine , selects and dialogs , ctx cancel , owned lineedit fork )
termprogress/ ( owned spinner + bar , tty gated )
termoutput/ ( tables + json yaml + graph , owned humanize subset , termcore shlex )
termstatic/ ( appname banner , out only , no input )
termretry/ ( owned retry engine fork , attempts delay veto )
termpager/ ( less paging , tty gated , epipe safe )
termtable/ ( one table library , five styles , owned renderer )
termbox/ ( framed boxes , termcore widths )
termimage/ ( image rendering , kitty + iterm2 + sixel + ansi fallback )
cmd/terminull ( dogfood generator cli )
examples/ ( runnable termtable style demo )

note: the tree uses the term* names; the v1 folders were renamed right after phase 14.1, ahead of 14.7.

# layout:

termapp/ ( public , importable as CHE3MZ/terminull/termapp )
termconfig/ ( public , importable as CHE3MZ/terminull/termconfig )
termrender/ ( public , importable as CHE3MZ/terminull/termrender )
termcore/ ( public , importable as CHE3MZ/terminull/termcore )
termstyle/ ( public , importable as CHE3MZ/terminull/termstyle )
termprompt/ ( public , importable as CHE3MZ/terminull/termprompt )
termprogress/ ( public , importable as CHE3MZ/terminull/termprogress )
termoutput/ ( public , importable as CHE3MZ/terminull/termoutput )
termstatic/ ( public , importable as CHE3MZ/terminull/termstatic )
termretry/ ( public , importable as CHE3MZ/terminull/termretry )
termpager/ ( public , importable as CHE3MZ/terminull/termpager )
termtable/ ( public , importable as CHE3MZ/terminull/termtable )
termbox/ ( public , importable as CHE3MZ/terminull/termbox )
termimage/ ( public , importable as CHE3MZ/terminull/termimage )
*/internal/ ( forked engines , owned by their library , never importable by apps )
cmd/terminull ( dogfood binary , not importable )
internal/ ( generator + terminull cli wiring , never imported by apps )
tests/ ( integration , runs on windows mac linux )
examples/ ( runnable demos , not importable )
scripts/ ( build test smoke + ops , all sh )
docs/ ( index + dependencies + architecture + spec + TODO )

# lifecycle:

how every app runs from start to finish:
1. main makes a stop context for ctrl+c and sigterm.
2. termapp.New creates the root command from the owned cli engine with SilenceUsage true and SilenceErrors true.
3. termapp adds global flags once: config, output, no-color, verbose, debug, log-style.
4. outputs default to process streams, tests redirect with SetOut SetErr SetIn.
5. PersistentPreRunE runs in fixed order: debug switch, console ansi setup, output and log-style validation, renderer plus logger, config load.
6. the cli engine parses flags and finds subcommand, help and version bypass live render.
7. RunE wrapper builds ctx with context, config, logger, io, output mode, flag getters.
8. app code runs with ctx only, returns error or nil.
9. termcore maps error to exit code, stops render, flushes logs, exits.

rules:
only one writer owns the terminal per phase, never two at once.
prompt pauses render before asking and resumes after via ctx PauseDrawing and ResumeDrawing.
spinners and bars do the same: pause, draw, resume, never beside live tree output.
logger drains before finalize via Sync, stops join the draw loop before tables print.
static help version completion always use simple plain output even on tty.
every prompt has a flag fallback, non terminal must still work with flags alone.
validate names and flags before banners and logs, so usage errors print clean and fast.

# termapp:

termapp is the boss, it owns the cli engine inside ( literal fork of cobra + pflag ) and never exposes it except one Raw escape.
termapp holds name, version, root command, subcommands.
termapp owns global flags: --config, --output table json yaml, --no-color, --verbose, --debug, --log-style tree plain time time-level logrus k8s.
globals carry no shorthands, so subcommands own -f -o -e -v freely.
WithLogStyle sets the default log style; --log-style overrides it.
WithLogStyleColors customizes flat prefix colors; --no-color wins.
help, version, and completion output come from the owned engine and stay stock-compatible.
apps declare Command with use, short, flags, run with ctx, plus Args and ValidArgs completion.
flag callbacks take termapp.FlagSet, arg validation takes termapp validators, completions take termapp directives.
Raw and Unwrap stay as narrow escape hatches, user code never imports cobra, pflag, or the cli fork.
WithSymbolsOnly pins ascii spinners and statuses on any terminal.
App.ManPages renders one roff page per visible command; the hidden man command writes them to disk.
CheckUpdate and NotifyUpdate implement a state-file-throttled GitHub release check.

# termconfig:

termconfig loads in order: flags win, then env, then file, then defaults.
file lookup: explicit --config, then home dotfile, then user config dir, then local files; both .yaml and .toml are searched.
env prefix is upper app name plus underscore, example MYAPP_ENV.
decode is strict via the internal mapstructure fork's ErrorUnused, unknown keys fail loudly.
termconfig never prompts, it only loads and merges, prompt decides what is missing.
the internal koanf fork is the merge engine; yaml.v3 and BurntSushi/toml are the codecs, with a thin toml adapter in termconfig/internal/koanf/toml.
Config.Watch polls the file search path and returns a fresh Config on size or mtime changes; the receiver is not mutated.

# termrender:

termrender is the fork of cirruslabs/echelon, MIT kept and owned in this repo.
api stays compatible: NewLogger, Scoped, Infof Debugf Warnf Errorf, AsWriter, Finish.
single writer with one draw goroutine, all output goes through it.
modes: interactive live tree on tty, simple lines on ci pipe dumb.
flat styles: plain, time [15:04:05], time-level [15:04:05] INFO, logrus INFO[0000], k8s I1026 15:04:05.123456.
use termapp.WithLogStyle or --log-style; root messages stay ctx.Logger.Infof, scopes still work.
non-tree styles force flat lines on tty; AsWriter stays raw so subprocess output is not decorated.
tty prefix colors: error red, warn yellow, info blue, debug and trace gray, timestamps and elapsed use the terminal default; termapp.WithLogStyleColors overrides, --no-color wins.
logrus truncates level names to four characters: ERRO, WARN, INFO, DEBU, TRAC.
levels are error warn info debug trace; success and skipped are scope outcomes, not log levels; there is no fatal level.
emoji where the terminal renders it including windows terminal, ascii SymbolsOnly elsewhere.
ci decorators for travis github teamcity are kept from echelon pattern.
setup runs synchronously before handlers, drawing starts alongside, stop joins the loop.
deps are termcore and termstyle, no heavy mods and no x/text.

# termstyle:

termstyle decides color or no color and wraps every writer.
detect order: --no-color flag, NO_COLOR env, TERM dumb, not a tty, CI true.
termcore ColorFor answers ansi-or-plain for any writer, pipes and buffers stay plain.
termcore EnsureConsole enables ansi on windows consoles up front, best effort elsewhere.
Scheme paints named styles and icons, plain when disabled.
Scheme.Faint dims a named color by half luminance, FaintRGB dims any rgb triple.
Bold Italic Underline Strikethrough emit exact sgr with dedicated resets.
termstyle foreground and decoration constants are owned values from the internal color fork, user code never imports a color mod.
Muted dims the current color through sgr faint, Faint is the deterministic form.
16 color terms downshift bright colors or snap rgb to the nearest base.
termcore ColorDepth feeds dimming and image tiers: case-insensitive COLORTERM truecolor wins, a real windows console is 24-bit under VT, then 256color terms, else 16.
termcore TerminalWidth and TerminalHeight report console metrics with safe fallbacks, used to fit bars and bound frames.
termcore also owns timed terminal capability queries, cell pixel size, and no-echo password reads, so no other package touches raw or console modes.
all widths come from termcore, whose owned width engine carries its own unicode tables; cjk never breaks tables.
piped out like app list piped to grep must be plain with no codes.

# termprompt:

termprompt prints a bullet question and reads one line, no cursor tracking, no re-renders.
the bullet paints green through termstyle.Scheme only on capable terminal destinations, plain everywhere else.
selects are numbered lists, confirms are y/n, passwords use the os no-echo read.
every call takes ctx, ctrl+c cancels the read and maps to exit code 130.
presets return validated without prompting, so flags always win.
termprompt only runs when isatty true, else it returns a clear flag required error naming the flag.
termprompt Editor opts into the internal lineedit fork for repl input: history, ctrl+r search, tab completion, multiline.
editor lends the console only for the duration of a Line call and hands it back, pipes and dumb terminals degrade to plain reads.
editor ctrl+c returns context canceled, ctrl+d and piped eof return io eof.
termprompt tests drive real reads over pipes, never real tty.

# termprogress:

termprogress owns its spinner and bar, no widget mods remain.
spinners animate only on terminal files, quiet no-op on pipes and files.
spinner frames are caller-provided; CharSets carries the briandowns tables, and NewSpinner gates dots with termcore.Emoji so ascii stays safe everywhere.
spinner redraws clear wrapped lines and hide the cursor on ansi terminals; legacy windows consoles overwrite with spaces instead.
spinner stop clears leftover glyphs and writes the final message, so no stale frame remains.
bars render every step as plain lines, so ci logs keep full progress.
spinner frames and bar fills paint through an optional style scheme; the zero scheme keeps them plain.
bars shrink on narrow terminals instead of wrapping.
multiple bars and logs share one writer but never draw at once, callers sequence them.

# termoutput:

termoutput prints app results as table, json, yaml based on --output flag, default table.
tables come from termtable, bordered by default, borderless for script-friendly.
TableWith and TableStyleWith style headers on any termtable style, pipes stay verbatim.
graphs are owned ascii point charts with labeled axis, no color codes inside.
sizes use the owned humanize subset, command splitting uses termcore shlex.
termoutput never reads input, it is out only, same static rules as termstatic.
json and yaml must be stable key order for tests and pipes.

# termtable:

termtable renders the same rows in five styles behind one small api.
StyleBordered is a unicode box, the default look; StyleRounded, StyleASCII, and StyleMarkdown cover docs, logs, and pipes; StyleBorderless is padded columns for scripts.
styles arrive as enum values, unknown styles error instead of guessing.
empty tables render nothing, ragged rows normalize to header width.
termtable carries no style scheme and stays monochrome: termoutput.TableWith applies header emphasis at the call site and every renderer measures visible width, so styled cells align on all five styles.

# termbox:

termbox renders framed boxes around arbitrary text, no third party code.
styles are rounded, hard, double, ascii, none, unknown styles render rounded.
titles sit in the top border and truncate with an ellipsis when the box is narrow.
padding is inner horizontal spaces, width sets the outer size and wraps content.
content may carry ansi, borders measure visible glyphs so lines stay aligned.
open styles reset at each line end and resume on the next wrapped line, borders stay clean.
cjk counts double via termcore width, combining marks stay with their base rune.
frame color goes through scheme, a disabled scheme prints plain bytes.

# termimage:

termimage renders images for terminals with no third-party image or protocol mods, stdlib decode.
backends are kitty graphics, iterm2 inline png, sixel, and ansi half-blocks or ascii.
auto detection reads env hints first, then probes the terminal for sixel with a DA1 query.
kitty graphics cover kitty, ghostty, and konsole 22.04+, iterm2 covers iterm and wezterm (kitty placements there bind to cells and windows conpty mangles the escapes), sixel covers windows terminal, mlterm, foot, and contour.
kitty images are anonymous by default so sequential renders coexist, a nonzero Options.ImageID redraws one image in place.
width and height are cell counts, zero width uses the terminal width and keeps aspect.
sixel builds an adaptive palette from the image, and dithering spreads quantization error on request across every tier that quantizes: sixel, 256-color half-blocks, and the ascii ramp; kitty and iterm2 send exact pixels.
windows consoles size sixel rasters in DEC 10x20 cell units so aspect holds on any font, posix terminals use real cell pixels.
transparency survives by default: kitty and iterm2 send native alpha, sixel asks the terminal to keep its background and leaves those pixels unpainted, half-blocks fall back to lower-half or blank cells with default colors, and ascii leaves blanks; a background color with alpha flattens every backend instead.
half-blocks need 256 colors, color off or 16 colors degrade to ascii, pipes never carry escapes.
image tiers read termcore ColorDepth, the same detector that drives text dimming, so backends and dimming always agree.
tmux is detected: kitty and iterm2 streams route through a passthrough wrapper (allow-passthrough must be on), sixel stays raw so tmux 3.4+ renders it natively, and tmux plus iterm2 drop single sequences much over a megabyte, so oversized renders can vanish there.
file decode is stdlib png jpeg gif, scale is box average for shrink and bilinear for grow.

# termretry:

termretry reruns fallible ops with capped attempts and backoff, on the internal retry engine fork.
defaults are 3 attempts and 1 second base, veto with RetryIf for permanent errors.
used for go mod tidy over flaky networks.

# termpager:

termpager pages long output through less on capable terminals.
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

# termstatic:

termstatic prints framework things that are the same for every app: appname banner and version line.
banner paints green when capable through termcore ColorFor, plain otherwise.
termstatic never asks input, never starts spinner or bar, always plain through termstyle.
termstatic works on dumb, ci, piped, no color without change.
termstatic finishes fast and exits cleanly on ctrl+c.
app specific lists like list users belong in output, not static, but follow the same out only rules.

# termcore:

termcore holds signals, exit codes, version, slog logging, plus the shared terminal and text primitives: widths, ansi parse, tty, color depth, console setup, terminal width and height, capability queries, password reads, shell split.
signals: SIGINT SIGTERM cancel the app context, render stops, prompt aborts, exit code 130 on ctrl+c.
exit codes: 0 ok, 1 error, 2 usage, 130 interrupted, custom via ExitCodeError like cirrus helpers.
version: semver from build info plus git tag, --version uses static plain output.
logging: slog with LevelVar, --verbose sets debug, --debug sets debug.
Level wires into app slog handlers with termcore.Level.

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
termprompt tests drive real reads over pipes, never real tty.
termrender tests assert join silence after stop and no garble under race detector.
no network in tests except dogfood CI which tidies for real.

# versioning:

semver, public folder names and function names do not break without a major bump. the term* renames ship with the /v2 module at v2.0.1.
engines are frozen literal forks pinned in .github/THIRD_PARTY.md; upstream fixes are cherry-picked only for a bug we hit or a capability we document.
owned code (termrender, termprompt, termprogress, termoutput, termtable, termcore) evolves in place with tests.
termrender keeps echelon api names so the cirrus pattern still ports.
deprecations warn for one minor before removal.
