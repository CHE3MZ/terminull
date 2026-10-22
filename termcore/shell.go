package termcore

import (
	"github.com/CHE3MZ/terminull/termcore/internal/shlex"
)

// ShellSplit splits a command line into arguments using POSIX shell
// rules: quoting, escapes, and comments. It is the one shell-splitting
// implementation in the tree, shared by pager parsing and command
// output filtering.
func ShellSplit(s string) ([]string, error) {
	return shlex.Split(s)
}
