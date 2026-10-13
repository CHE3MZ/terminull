package style

import (
	"github.com/mattn/go-runewidth"
)

// Width returns the terminal column width of s.
// East-Asian wide runes count as 2, so CJK text never breaks tables.
func Width(s string) int {
	return runewidth.StringWidth(s)
}

// Truncate shortens s to at most max columns, appending "..." when cut.
// A max below 4 still returns a valid (shorter) string, never panics.
func Truncate(s string, max int) string {
	if max < 0 {
		max = 0
	}

	return runewidth.Truncate(s, max, "...")
}
