package app

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Command describes one CLI command. Children build groups like
// `worker {run|pause}`; a parent without Run prints its help.
type Command struct {
	// Use is the one-line usage, e.g. "deploy --env prod".
	Use string
	// Short is the short help shown in command lists.
	Short string
	// Long is the full help text.
	Long string
	// Args validates positional args (e.g. cobra.MaximumNArgs(1)).
	Args cobra.PositionalArgs
	// Flags registers command-local flags. Read them back with
	// Context.String/Bool/Int/StringSlice.
	Flags func(*pflag.FlagSet)
	// Children builds a command group; they share the parent's flags.
	Children []Command
	// Run executes the command. Nil on a group parent prints help.
	Run func(*Context) error
}
