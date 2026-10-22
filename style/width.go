package style

import (
	"github.com/CHE3MZ/terminull/core"
)

// Width returns the terminal column width of s.
func Width(s string) int {
	return core.Width(s)
}

// RuneWidth returns the terminal column width of one rune.
func RuneWidth(r rune) int {
	return core.RuneWidth(r)
}

// Truncate shortens s to at most max columns, appending "..." when cut.
func Truncate(s string, max int) string {
	return core.Truncate(s, max)
}
