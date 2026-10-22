package console

import (
	"os"

	"github.com/CHE3MZ/terminull/termcore"
)

// PrepareTerminalEnvironment enables ANSI processing on Windows consoles
// (a no-op where already on, an error where unsupported).
func PrepareTerminalEnvironment() error {
	return termcore.EnsureConsole()
}

// TerminalHeight is unimplemented on Windows: figure out how to query
// console height there, until then callers draw unbounded frames.
func TerminalHeight(file *os.File) int {
	// todo: figure out how to find out console height on Windows
	return -1
}
