package style

import (
	"github.com/CHE3MZ/terminull/core"
)

// EnsureConsole prepares the console for ANSI output.
func EnsureConsole() error {
	return core.EnsureConsole()
}
