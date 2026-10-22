package termstyle

import (
	"github.com/CHE3MZ/terminull/termcore"
)

// EnsureConsole prepares the console for ANSI output.
func EnsureConsole() error {
	return termcore.EnsureConsole()
}
