package termcore

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

// TerminalHeight returns console rows for f, or -1 when unknown (pipes,
// files, and dumb terminals). Callers treat -1 as "draw unbounded".
func TerminalHeight(f *os.File) int {
	if rows, ok := terminalRows(f); ok {
		return rows
	}
	return -1
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
