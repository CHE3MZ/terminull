package core

import (
	width "github.com/CHE3MZ/terminull/core/internal/width"
)

// Width returns the terminal column width of s.
// East-Asian wide runes count as 2, so CJK text never breaks tables.
func Width(s string) int {
	return width.StringWidth(s)
}

// RuneWidth returns the terminal column width of one rune: 2 for CJK
// wide runes, 0 for combining marks, 1 otherwise. Wrappers use it to
// never split a wide rune across lines.
func RuneWidth(r rune) int {
	return width.RuneWidth(r)
}

// Truncate shortens s to at most max columns, appending "..." when cut.
// A max below 4 still returns a valid (shorter) string, never panics.
func Truncate(s string, max int) string {
	if max < 0 {
		max = 0
	}

	return width.Truncate(s, max, "...")
}
