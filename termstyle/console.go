package termstyle

import (
	"github.com/CHE3MZ/terminull/v2/termcore"
)

// EnsureConsole prepares the console for ANSI output.
func EnsureConsole() error {
	return termcore.EnsureConsole()
}
