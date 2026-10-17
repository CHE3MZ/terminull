package style

import (
	"strings"
)

// VisibleWidth returns the terminal column width of s ignoring ANSI
// escape sequences (CSI colors, OSC hyperlinks): styled text measures
// by its glyphs, so it embeds in boxes and tables without skewing
// alignment.
func VisibleWidth(s string) int {
	return Width(stripANSI(s))
}

// EscapeLen returns the byte length of the ANSI escape sequence
// beginning at index i, or 0 when no escape starts there (i out of
// range or not an ESC byte). It covers CSI, OSC (BEL- or ST-terminated),
// and two-byte escapes, so walkers can skip sequences while measuring
// or slicing styled text.
func EscapeLen(s string, i int) int {
	if i < 0 || i >= len(s) || s[i] != 0x1b {
		return 0
	}
	if i+1 >= len(s) {
		return 1
	}
	switch s[i+1] {
	case '[': // CSI: parameters, then a final byte 0x40-0x7E.
		j := i + 2
		for j < len(s) && (s[j] < 0x40 || s[j] > 0x7e) {
			j++
		}
		if j < len(s) {
			j++
		}
		return j - i
	case ']': // OSC: terminated by BEL or ST.
		j := i + 2
		for j < len(s) {
			if s[j] == 0x07 {
				return j + 1 - i
			}
			if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '\\' {
				return j + 2 - i
			}
			j++
		}
		return len(s) - i
	default: // Two-byte escape, e.g. ESC M.
		return 2
	}
}

// stripANSI removes ANSI escape sequences, leaving plain glyphs.
func stripANSI(s string) string {
	if !strings.ContainsRune(s, 0x1b) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if n := EscapeLen(s, i); n > 0 {
			i += n
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
