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
	// Example shows usage examples under an Examples heading.
	Example string
	// Args validates positional args (e.g. cobra.MaximumNArgs(1)).
	Args cobra.PositionalArgs
	// ValidArgs lists static shell completions for positional args.
	ValidArgs []string
	// ValidArgsFunction computes dynamic completions. It mirrors
	// cobra's hook, including the ShellCompDirective return.
	ValidArgsFunction func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
	// Hidden removes the command from help lists without disabling it,
	// for internal and deprecated commands.
	Hidden bool
	// Flags registers command-local flags. Read them back with
	// Context.String/Bool/Int/StringSlice.
	Flags func(*pflag.FlagSet)
	// Children builds a command group; they share the parent's flags.
	Children []Command
	// Run executes the command. Nil on a group parent prints help.
	Run func(*Context) error
}
