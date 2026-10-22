package style

import (
	"github.com/CHE3MZ/terminull/core"
)

// VisibleWidth returns the terminal column width of s ignoring ANSI
// escape sequences.
func VisibleWidth(s string) int {
	return core.VisibleWidth(s)
}

// EscapeLen returns the byte length of the ANSI escape sequence
// beginning at index i, or 0 when no escape starts there.
func EscapeLen(s string, i int) int {
	return core.EscapeLen(s, i)
}
