package termstyle

import (
	"os"

	"github.com/CHE3MZ/terminull/termcore"
)

// IsTerminal reports whether f is an interactive terminal.
func IsTerminal(f *os.File) bool {
	return termcore.IsTerminal(f)
}

// Color reports whether ANSI color codes may be written to f.
func Color(f *os.File, noColor bool) bool {
	return termcore.Color(f, noColor)
}

// Emoji reports whether emoji glyphs may be used.
func Emoji() bool {
	return termcore.Emoji()
}
