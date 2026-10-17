# structure:

app/ ( cobra wrapper , root and subcommands )
config/ ( koanf + yaml , flags env file merge )
render/ ( fork of echelon , single writer logs )
style/ ( color + isatty + colorable + runewidth , windows safe )
prompt/ ( line engine , selects and dialogs , ctx cancel )
progress/ ( owned spinner + bar , tty gated )
output/ ( tables + json yaml + graph , humanize shlex )
static/ ( appname banner , out only , no input )
core/ ( signals + exit codes + version , slog )
retry/ ( backoff retries , attempts delay veto )
pager/ ( less paging , tty gated , epipe safe )
table/ ( aqua default + rodaine styles )
box/ ( framed boxes , ansi aware widths )
cmd/terminull ( dogfood generator cli )

# layout:

app/ ( public , importable )
config/ ( public , importable )
render/ ( public , importable )
style/ ( public , importable )
prompt/ ( public , importable )
progress/ ( public , importable )
output/ ( public , importable )
static/ ( public , importable )
core/ ( public , importable )
retry/ ( public , importable )
pager/ ( public , importable )
table/ ( public , importable )
box/ ( public , importable )
cmd/ ( binaries , not importable )
internal/ ( generator + terminull wiring , private )
tests/ ( integration , windows mac linux )
scripts/ ( build test smoke + ops , all sh )
docs/ ( index + dependencies + architecture + spec + TODO )

# explanation:

terminull is a box of parts for building command line apps.
you pick the parts, you do not wire the messy bits yourself.

how an app runs from start to finish:
1. app starts and listens for stop keys like ctrl+c.
2. config loads settings in this order: flags win, then env, then file, then defaults.
3. style checks what terminal this is: windows, mac, linux, dumb, ci, no color.
4. render becomes the single writer, but only one party draws at a time.
5. prompt asks questions only if a flag is missing and the terminal is interactive.
6. progress animates while the tree is paused, then the tree resumes alone.
7. output prints the final result as table, json, or yaml, after drawing stops.
8. core shuts down cleanly and returns the right exit code.

what each folder does:
app/ is the boss. it holds the root command and subcommands. it owns flags, help, version, and shutdown. apps never touch cobra directly, only through app.
config/ loads yaml files and env vars and merges them with flags. it fails loudly on unknown keys so typos are caught early.
render/ is our fork of echelon. it keeps a tree of jobs like task and subtask. it draws live on a normal terminal and plain lines on ci. it has an emoji mode and a plain ascii mode, picked by capability including windows terminal.
style/ decides color or no color. it wraps all output so windows conhost shows the same as mac and linux. no app prints with fmt directly, all go through style and render.
prompt/ prints a bullet question and reads one line, no cursor games. it works when interactive and errors clearly when not, telling the user which flag to pass instead.
progress/ owns its spinner and bar. spinners animate only on terminals and stay quiet on pipes. bars print every step so ci logs stay complete.
output/ prints tables, graphs, sizes, durations. graphs are plain ascii so they work everywhere. tables come in aqua bordered style by default and rodaine borderless for scripts.
static/ prints out only things like app name banner. it never asks for input, never starts spinners or prompts, and stays plain unless the terminal earned color. it finishes fast and exits cleanly on ctrl+c.
core/ holds signals, exit codes, version, and slog logging. debug and verbose live here.
retry/ reruns flaky ops like tidy with backoff until attempts run out.
pager/ pages long output through less on terminals and passes pipes through untouched.
table/ merges two table engines behind one style switch, aqua default, rodaine for scripts.
box/ draws frames around text: rounded, hard, double, ascii corners, titles, padding, and wrapping. it measures visible widths so ansi colors and cjk stay aligned inside the frame.
cmd/terminull/ is our own cli built from all of the above, it scaffolds new projects.

rules to keep it stable:
one writer owns the terminal per phase, never two at once.
pause drawing around prompts and widgets, stop drawing before result tables.
logger drains before finalize, stops join the draw loop.
apps import only top level folders like app and config. never import internal. internal is private tools for us.
every prompt has a flag fallback. if there is no terminal, the app must still work with flags alone.
every new part needs tests on windows, mac, and linux, plus piped and non tty paths.
wrapping matters more than features. underlying mods can change, our folder names and function names do not.
