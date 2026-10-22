package termstyle

import (
	"github.com/CHE3MZ/terminull/v2/termcore"
)

// Width returns the terminal column width of s.
func Width(s string) int {
	return termcore.Width(s)
}

// RuneWidth returns the terminal column width of one rune.
func RuneWidth(r rune) int {
	return termcore.RuneWidth(r)
}

// Truncate shortens s to at most max columns, appending "..." when cut.
func Truncate(s string, max int) string {
	return termcore.Truncate(s, max)
}
