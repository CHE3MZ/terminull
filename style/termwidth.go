package style

import (
	"io"
	"os"

	"github.com/CHE3MZ/terminull/core"
)

// DefaultTerminalWidth is used when detection fails.
const DefaultTerminalWidth = core.DefaultTerminalWidth

// TerminalWidth returns console columns for f, or DefaultTerminalWidth
// when unknown.
func TerminalWidth(f *os.File) int {
	return core.TerminalWidth(f)
}

// ColorFor reports whether ANSI may be written to w.
func ColorFor(w io.Writer, noColor bool) bool {
	return core.ColorFor(w, noColor)
}
