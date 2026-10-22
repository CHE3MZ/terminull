package termapp

// ShellCompDirective controls shell completion behavior. Values are
// bit-identical to cobra's directives (asserted in tests), so behavior
// matches exactly while user code imports only app.
type ShellCompDirective int

const (
	// ShellCompDirectiveError discards completions on error.
	ShellCompDirectiveError ShellCompDirective = 1 << iota
	// ShellCompDirectiveNoSpace skips the trailing space after a
	// single completion.
	ShellCompDirectiveNoSpace
	// ShellCompDirectiveNoFileComp disables file completion when no
	// completion is provided.
	ShellCompDirectiveNoFileComp
	// ShellCompDirectiveFilterFileExt filters files by the provided
	// extensions.
	ShellCompDirectiveFilterFileExt
	// ShellCompDirectiveFilterDirs completes directory names only.
	ShellCompDirectiveFilterDirs
	// ShellCompDirectiveKeepOrder preserves completion order.
	ShellCompDirectiveKeepOrder
	// ShellCompDirectiveDefault lets the shell use default behavior.
	ShellCompDirectiveDefault ShellCompDirective = 0
)

// CompletionFunc computes positional completions. It drops cobra's
// *Command parameter: handlers complete from args and the prefix only,
// the framework owns the rest.
type CompletionFunc func(args []string, toComplete string) ([]string, ShellCompDirective)
