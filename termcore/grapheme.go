package termcore

import (
	width "github.com/CHE3MZ/terminull/v2/termcore/internal/width"
)

// FirstGrapheme returns the first extended grapheme cluster in s and its
// display width. It is the unit wrappers and editors should move by when
// they need to keep emoji sequences, combining marks, flags, and Hangul
// syllables together. ANSI escape sequences are not special here; callers
// that handle styled text must skip them before calling.
func FirstGrapheme(s string) (string, int) {
	return width.FirstGrapheme(s)
}
