package style

import (
	"io"
	"os"
)

// DefaultTerminalWidth is used when detection fails (pipes, files,
// dumb terminals): the classic 80 columns.
const DefaultTerminalWidth = 80

// TerminalWidth returns console columns for f, or DefaultTerminalWidth
// when unknown. Never returns zero or negative.
func TerminalWidth(f *os.File) int {
	if cols, ok := terminalColumns(f); ok {
		return cols
	}
	return DefaultTerminalWidth
}

// ColorFor reports whether ANSI may be written to w: false for
// --no-color, non-file writers (buffers, pipes without files), dumb
// terminals, and CI; true only for capable terminal files.
func ColorFor(w io.Writer, noColor bool) bool {
	if noColor {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return Color(f, false)
}
