package style

import (
	"os"

	"github.com/CHE3MZ/terminull/core"
)

// IsTerminal reports whether f is an interactive terminal.
func IsTerminal(f *os.File) bool {
	return core.IsTerminal(f)
}

// Color reports whether ANSI color codes may be written to f.
func Color(f *os.File, noColor bool) bool {
	return core.Color(f, noColor)
}

// Emoji reports whether emoji glyphs may be used.
func Emoji() bool {
	return core.Emoji()
}
