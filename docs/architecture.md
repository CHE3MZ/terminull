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

app/ ( public , importable )
config/ ( public , importable )
render/ ( public , importable )
style/ ( public , importable )
prompt/ ( public , importable )
progress/ ( public , importable )
output/ ( public , importable )
static/ ( public , importable )
core/ ( public , importable )
internal/ ( private , tooling only )
tests/ ( cross platform matrix , windows mac linux )
scripts/ ( task helpers )
docs/ ( index + dependencies + architecture )

# explanation:

terminull is a box of parts for building command line apps.
you pick the parts, you do not wire the messy bits yourself.

how an app runs from start to finish:
1. app starts and listens for stop keys like ctrl+c.
2. config loads settings in this order: flags win, then env, then file, then defaults.
3. style checks what terminal this is: windows, mac, linux, dumb, ci, no color.
4. render becomes the single writer. nothing else is allowed to print directly.
5. prompt asks questions only if a flag is missing and the terminal is interactive.
6. progress shows spinners and bars through render, so logs never garble.
7. output prints the final result as table, json, or yaml.
8. core shuts down cleanly and returns the right exit code.

what each folder does:
app/ is the boss. it holds the root command and subcommands. it owns flags, help, version, and shutdown. apps never touch cobra directly, only through app.
config/ loads yaml files and env vars and merges them with flags. it fails loudly on unknown keys so typos are caught early.
render/ is our fork of echelon. it keeps a tree of jobs like task and subtask. it draws live on a normal terminal and plain lines on ci. it has an emoji mode and a plain ascii mode for old windows.
style/ decides color or no color. it wraps all output so windows conhost shows the same as mac and linux. no app prints with fmt directly, all go through style and render.
prompt/ asks selects, confirms, inputs, passwords. it works when interactive and errors clearly when not, telling the user which flag to pass instead.
progress/ owns spinners and bars. apps ask progress for a spinner or bar, they never create their own. this keeps one draw loop and stops flicker.
output/ prints tables, graphs, sizes, durations. graphs are plain ascii so they work everywhere. tables handle wide letters correctly.
static/ prints out only things like app name, help, version, and shell completion. these never ask for input, never start spinners or prompts, and always go through style so no color and dumb modes still look right. they finish fast and exit cleanly on ctrl+c.
core/ holds signals, exit codes, version, and slog logging. debug and verbose live here.

rules to keep it stable:
only render writes to the screen. style, progress, output, static all go through render.
apps import only top level folders like app and config. never import internal. internal is private tools for us.
every prompt has a flag fallback. if there is no terminal, the app must still work with flags alone.
every new part needs a test on windows, mac, and linux, plus dumb terminal and ci mode.
wrapping matters more than features. underlying mods can change, our folder names and function names do not.
