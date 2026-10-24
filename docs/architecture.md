# structure (v2 target):

termcore/ ( shared base: text widths , ansi parse , tty console depth , shlex , signals exit codes version slog )
termstyle/ ( scheme paint faint rgb , owned color engine fork )
termbox/ ( framed boxes , ansi aware widths )
termtable/ ( one table library , five styles , owned renderer )
termimage/ ( image rendering , kitty + iterm2 + sixel + ansi fallback )
termoutput/ ( tables + json yaml + graph , owned humanize subset )
termprogress/ ( owned spinner + bar , tty gated )
termprompt/ ( line engine , selects and dialogs , owned lineedit fork )
termrender/ ( fork of echelon , single writer logs )
termpager/ ( less paging , tty gated , epipe safe )
termstatic/ ( appname banner , out only , no input )
termretry/ ( owned retry engine fork )
termconfig/ ( owned merge engine fork + yaml )
termapp/ ( owned cli engine fork , root and subcommands )
cmd/terminull ( dogfood generator cli )

note: the tree uses the term* names; the v1 folders were renamed right after phase 14.1, ahead of 14.7.

# layout:

termcore/ ( public , importable as CHE3MZ/terminull/termcore )
termstyle/ ( public , importable as CHE3MZ/terminull/termstyle )
termbox/ ( public , importable as CHE3MZ/terminull/termbox )
termtable/ ( public , importable as CHE3MZ/terminull/termtable )
termimage/ ( public , importable as CHE3MZ/terminull/termimage )
termoutput/ ( public , importable as CHE3MZ/terminull/termoutput )
termprogress/ ( public , importable as CHE3MZ/terminull/termprogress )
termprompt/ ( public , importable as CHE3MZ/terminull/termprompt )
termrender/ ( public , importable as CHE3MZ/terminull/termrender )
termpager/ ( public , importable as CHE3MZ/terminull/termpager )
termstatic/ ( public , importable as CHE3MZ/terminull/termstatic )
termretry/ ( public , importable as CHE3MZ/terminull/termretry )
termconfig/ ( public , importable as CHE3MZ/terminull/termconfig )
termapp/ ( public , importable as CHE3MZ/terminull/termapp )
*/internal/ ( forked engines , never importable by apps )
cmd/ ( binaries , not importable )
internal/ ( generator + terminull wiring , private )
tests/ ( integration , windows mac linux )
examples/ ( runnable demos , not importable )
scripts/ ( build test smoke + ops , all sh )
docs/ ( index + dependencies + architecture + spec + TODO + v2 plan )

# ownership model:

terminull is a workspace of libraries, not a monolith with helpers.
every term* folder is a library you can import alone and get its full feature set.
termcore is the shared base: it is the only place that owns terminal facts (width, height, tty, depth, console mode), terminal I/O (raw-mode queries, no-echo password reads), and text primitives (ansi parse, shell split).
component engines (color, cli, config merge, line editing, retry, humanize subset) are forked literally into that component's internal tree, licensed and attributed, and never surface in its public api.
termapp is the combination layer: it wires the libraries into an app framework, convenience rather than requirement.
cmd/terminull proves the combination end to end.

# layers:

L0 termcore → x/sys , x/term
L1 termstyle → termcore
L2 termbox termtable termprogress termstatic termpager termprompt termrender termimage → termcore ( + termstyle where color is used )
L3 termoutput → termtable termstyle termcore yaml.v3 ; termconfig → yaml.v3 ; termretry → stdlib
L4 termapp → every library below + its internal cli fork ; cmd and internal → termapp and friends

imports point down only, no cycles. a library's public api may only reference its own and lower-layer types. `tests/dependency_law_test.go` checks the edges and `scripts/standalone.sh` builds each library in a throwaway module.

# explanation:

terminull is a box of parts for building command line apps.
you pick the parts, you do not wire the messy bits yourself.
each part stands alone; the framework is the parts plugged together.

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
termapp/ is the boss. it holds the root command and subcommands, and owns the internal cli engine ( literal fork of cobra + pflag ). it owns flags, help, version, and shutdown. apps never touch the engine directly, only through termapp.
termconfig/ loads yaml files and env vars and merges them with flags. the merge engine is the internal koanf fork, yaml.v3 is the file codec. it fails loudly on unknown keys so typos are caught early.
termrender/ is our fork of echelon. it keeps a tree of jobs like task and subtask. it draws live on a normal terminal and plain lines on ci. it has an emoji mode and a plain ascii mode, picked by capability including windows terminal. its deps are termcore and termstyle.
termstyle/ decides color or no color. it owns the internal color engine fork and is the only place that builds text sgr. no app prints with fmt directly, all go through termstyle and termrender. widths come from termcore.
termprompt/ prints a bullet question and reads one line, no cursor games. it works when interactive and errors clearly when not, telling the user which flag to pass instead. its Editor is the opt-in exception for repls: internal lineedit fork-backed editing with history, ctrl+r, and completion, lent to a single Line call and handed back after, plain reads on pipes.
termprogress/ owns its spinner and bar. spinners animate only on terminals and stay quiet on pipes. spinner frames are caller-provided (CharSets carries the briandowns tables); ansi terminals clear wrapped frames and hide the cursor, legacy consoles overwrite with spaces, and stop clears leftovers before the final message. bars print every step so ci logs stay complete.
termoutput/ prints tables, graphs, sizes, durations. graphs are plain ascii so they work everywhere. tables come in bordered style by default and borderless for scripts. sizes use the owned humanize subset, command splitting uses termcore shlex.
termstatic/ prints out only things like app name banner. it never asks for input, never starts spinners or prompts, and stays plain unless the terminal earned color. it finishes fast and exits cleanly on ctrl+c.
termcore/ holds signals, exit codes, version, and slog logging, plus the shared terminal and text primitives: widths, ansi parse, isatty, console setup, ANSI capability, color depth, terminal width, shell split.
termretry/ reruns flaky ops like tidy with backoff until attempts run out, on the internal retry engine fork.
termpager/ pages long output through less on terminals and passes pipes through untouched.
termtable/ renders the same rows in five styles: bordered default, rounded, ascii, markdown, borderless. the renderer is private and never surfaces in the api.
termbox/ draws frames around text: rounded, hard, double, ascii corners, titles, padding, and wrapping. it measures visible widths so ansi colors and cjk stay aligned inside the frame.
termimage/ renders images with no third-party image or protocol code: kitty graphics, iterm2 inline png, sixel, and half-block or ascii fallbacks, picked by env hints plus a DA1 probe, sized in cells, tmux-safe.
cmd/terminull/ is our own cli built from all of the above, it scaffolds new projects.

rules to keep it stable:
one writer owns the terminal per phase, never two at once.
pause drawing around prompts and widgets, stop drawing before result tables.
logger drains before finalize, stops join the draw loop.
apps import only public term* libraries. never import internal, and never import a fork.
no package outside termcore asks the os about terminal facts: width, height, tty, depth, console mode, raw queries, or password reads.
only termstyle builds text sgr. termcore parses ansi. termimage builds image protocol escapes only.
width facts come from termcore only; no component keeps its own width table.
dependencies point down only; ci builds every package with go build ./... and dogfoods the generator.
every prompt has a flag fallback. if there is no terminal, the app must still work with flags alone.
every new part needs tests on windows, mac, and linux, plus piped and non tty paths.
wrapping matters more than features. underlying engines are frozen forks now; our folder names and function names do not change without a major.

fork policy:
literal forks: upstream source is copied into the owning library's internal tree at a pinned version, license kept.
provenance: .github/THIRD_PARTY.md records repo, version, commit, license, and fork date for every fork.
minimal modification: only integration seams change (termcore widths and ansi, console setup, error wrapping, removal of unused surface). engine logic and upstream tests stay recognizable so security fixes cherry-pick cleanly.
forked engines never surface in public apis; users see term* types only.
no cosmetic churn. upstream fixes arrive only for a bug we hit or a capability we document.
upstream watch: task upstream checks recorded origins for releases and advisories, replacing the signal osv and govulncheck lose once modules are renamed.
exempt: gopkg.in/yaml.v3 and golang.org/x/* stay external by design.
history: render keeps echelon api names and attribution, so the cirrus pattern still ports. termtable reclaimed its renderer end to end in v1, third-party engines removed, demo output byte-identical. v2 generalizes that play to the rest of the stack.
divergences are deliberate and documented next to the code or in spec: rows pad or truncate to headers, write errors propagate on every style, termtable stays monochrome, sixel stays raw inside tmux.
