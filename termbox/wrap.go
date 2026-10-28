package termbox

import (
	"strings"

	"github.com/CHE3MZ/terminull/v2/termcore"
)

// cell is one printable unit of a styled line: any ANSI prefix bytes
// starting at start, plus the grapheme cluster ending at end.
type cell struct {
	start int
	end   int
	text  string
	w     int
}

// cells splits a styled line into printable grapheme clusters, attaching
// leading escape sequences to the cluster they precede so slices never
// orphan color codes or split an emoji sequence mid-cluster.
func cells(s string) []cell {
	var out []cell
	pending := -1
	for i := 0; i < len(s); {
		if n := termcore.EscapeLen(s, i); n > 0 {
			if pending < 0 {
				pending = i
			}
			i += n
			continue
		}
		cluster, width := termcore.FirstGrapheme(s[i:])
		if cluster == "" {
			break
		}
		start := i
		if pending >= 0 {
			start = pending
			pending = -1
		}
		out = append(out, cell{start: start, end: i + len(cluster), text: cluster, w: width})
		i += len(cluster)
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
	for termcore.VisibleWidth(s) > cols {
		cut := breakAt(s, cols)
		out = append(out, strings.TrimRight(s[:cut], " "))
		s = strings.TrimLeft(s[cut:], " ")
	}
	return append(out, s)
}

// pendingSGR returns the SGR sequences in s that are still in effect at
// its end: every SGR escape since the last reset. Replaying the result
// verbatim resumes the exact styling, so box borders never inherit
// content color across wrapped lines.
func pendingSGR(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		n := termcore.EscapeLen(s, i)
		if n == 0 {
			i++
			continue
		}
		esc := s[i : i+n]
		if isSGR(esc) {
			parts := strings.Split(esc[2:len(esc)-1], ";")
			last := -1
			for j, p := range parts {
				if p == "" || p == "0" {
					last = j
				}
			}
			if last >= 0 {
				b.Reset()
				if rest := parts[last+1:]; len(rest) > 0 {
					b.WriteString("\x1b[")
					b.WriteString(strings.Join(rest, ";"))
					b.WriteByte('m')
				}
			} else {
				b.WriteString(esc)
			}
		}
		i += n
	}
	return b.String()
}

func isSGR(esc string) bool {
	return len(esc) >= 3 && esc[0] == 0x1b && esc[1] == '[' && esc[len(esc)-1] == 'm'
}

// breakAt returns the byte offset where s should split so the first
// part holds at most cols columns: before the last fitting space when
// one exists, else after the last fitting unit.
func breakAt(s string, cols int) int {
	cs := cells(s)
	if len(cs) == 0 {
		return len(s)
	}
	width := 0
	lastSpace := -1
	i := 0
	for i < len(cs) && width+cs[i].w <= cols {
		if cs[i].text == " " {
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
