package termapp

// Command describes one CLI command. Children build groups like
// `worker {run|pause}`; a parent without Run prints its help.
//
// Every type here is terminull-owned: define commands importing only
// app, never the internal engine packages. The framework maps to it inside
// build; App.Raw escapes for advanced use.
type Command struct {
	// Use is the one-line usage, e.g. "deploy --env prod".
	Use string
	// Short is the short help shown in command lists.
	Short string
	// Long is the full help text.
	Long string
	// Example shows usage examples under an Examples heading.
	Example string
	// Args validates positional args (e.g. MaximumNArgs(1)).
	Args PositionalArgs
	// ValidArgs lists static shell completions for positional args.
	ValidArgs []string
	// ValidArgsFunction computes dynamic completions from args and the
	// prefix; the framework owns the command, so no engine type leaks.
	ValidArgsFunction CompletionFunc
	// Hidden removes the command from help lists without disabling it,
	// for internal and deprecated commands.
	Hidden bool
	// Flags registers command-local flags. Read them back with
	// Context.String/Bool/Int/StringSlice.
	Flags func(*FlagSet)
	// Children builds a command group; they share the parent's flags.
	Children []Command
	// Run executes the command. Nil on a group parent prints help.
	Run func(*Context) error
}
