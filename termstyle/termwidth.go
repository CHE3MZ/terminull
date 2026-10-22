package termstyle

import (
	"io"
	"os"

	"github.com/CHE3MZ/terminull/termcore"
)

// DefaultTerminalWidth is used when detection fails.
const DefaultTerminalWidth = termcore.DefaultTerminalWidth

// TerminalWidth returns console columns for f, or DefaultTerminalWidth
// when unknown.
func TerminalWidth(f *os.File) int {
	return termcore.TerminalWidth(f)
}

// ColorFor reports whether ANSI may be written to w.
func ColorFor(w io.Writer, noColor bool) bool {
	return termcore.ColorFor(w, noColor)
}
