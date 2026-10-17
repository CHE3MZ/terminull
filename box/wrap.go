package box

import (
	"strings"
	"unicode/utf8"

	"github.com/CHE3MZ/terminull/style"
)

// cell is one printable unit of a styled line: any ANSI prefix bytes
// starting at start, plus the base rune ending at end.
type cell struct {
	start int
	end   int
	r     rune
	w     int
}

// cells splits a styled line into printable units, attaching leading
// escape sequences to the rune they precede so slices never orphan
// color codes mid-sequence.
func cells(s string) []cell {
	var out []cell
	pending := -1
	for i := 0; i < len(s); {
		if n := style.EscapeLen(s, i); n > 0 {
			if pending < 0 {
				pending = i
			}
			i += n
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		start := i
		if pending >= 0 {
			start = pending
			pending = -1
		}
		out = append(out, cell{start: start, end: i + size, r: r, w: style.RuneWidth(r)})
		i += size
	}
	if pending >= 0 {
		if len(out) > 0 {
			out[len(out)-1].end = len(s)
		} else {
			out = append(out, cell{start: pending, end: len(s)})
		}
	}
	return out
}

// wrapLines wraps every line to at most cols visible columns, breaking
// at spaces where possible and hard-cutting words longer than a line.
// Continuation lines lose the spaces they break on.
func wrapLines(lines []string, cols int) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, wrapLine(line, cols)...)
	}
	return out
}

func wrapLine(s string, cols int) []string {
	if cols < 1 {
		cols = 1
	}
	var out []string
	for style.VisibleWidth(s) > cols {
		cut := breakAt(s, cols)
		out = append(out, strings.TrimRight(s[:cut], " "))
		s = strings.TrimLeft(s[cut:], " ")
	}
	return append(out, s)
}

// breakAt returns the byte offset where s should split so the first
// part holds at most cols columns: before the last fitting space when
// one exists, else after the last fitting unit.
func breakAt(s string, cols int) int {
	cs := cells(s)
	width := 0
	lastSpace := -1
	i := 0
	for i < len(cs) && width+cs[i].w <= cols {
		if cs[i].r == ' ' {
			lastSpace = i
		}
		width += cs[i].w
		i++
	}
	if i == 0 {
		return cs[0].end
	}
	if lastSpace >= 0 {
		return cs[lastSpace].start
	}
	return cs[i-1].end
}
