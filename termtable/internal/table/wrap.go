package table

import (
	"strings"
	"unicode/utf8"

	"github.com/CHE3MZ/terminull/core"
)

// wrapCell lays a cell's text into display lines of at most width
// columns. Newlines always break, runs of spaces collapse to one, and
// words too long for a line are hyphenated. ANSI stays attached to its
// text: styles re-open on continuation lines and reset before the
// border, so colored cells never leak into the frame.
func wrapCell(s string, width int) []string {
	if width < 1 {
		width = 1
	}
	var lines []string
	for _, segment := range strings.Split(s, "\n") {
		lines = append(lines, wrapLine(segment, width)...)
	}
	if len(lines) == 0 {
		lines = []string{""}
	}
	return lines
}

func wrapLine(line string, width int) []string {
	words := splitWords(line)
	if len(words) == 0 {
		return []string{strings.TrimSpace(line)}
	}

	var lines []string
	current := ""
	currentWidth := 0
	push := func() {
		lines = append(lines, current)
		current = ""
		currentWidth = 0
	}

	for _, w := range words {
		for w.width > width {
			head, rest := cutCells(w.prefix+w.text, width-1)
			if currentWidth > 0 {
				push()
			}
			lines = append(lines, head+"-")
			w.prefix = ""
			w.text = rest
			w.width = core.VisibleWidth(rest)
		}
		token := w.prefix + w.text
		switch {
		case currentWidth == 0:
			current = token
			currentWidth = w.width
		case currentWidth+1+w.width <= width:
			current += " " + token
			currentWidth += 1 + w.width
		default:
			push()
			current = token
			currentWidth = w.width
		}
		current += w.suffix
	}
	if current != "" || len(lines) == 0 {
		lines = append(lines, current)
	}
	return lines
}

// word is one space-delimited token of a cell. prefix carries the
// style active when the token starts, suffix any style sequences seen
// right after it, so emission and wrapping keep exact escapes.
type word struct {
	prefix string
	text   string
	suffix string
	width  int
}

// splitWords tokenizes a line for wrapping, keeping ANSI sequences
// with their word.
func splitWords(line string) []word {
	var out []word
	var current word
	open := false
	active := ""

	flush := func() {
		if !open {
			return
		}
		current.width = core.VisibleWidth(current.text)
		out = append(out, current)
		current = word{}
		open = false
	}

	for i := 0; i < len(line); {
		if n := core.EscapeLen(line, i); n > 0 {
			seq := line[i : i+n]
			i += n
			switch {
			case open:
				current.text += seq
			case len(out) > 0:
				out[len(out)-1].suffix += seq
			}
			active = nextActive(active, seq)
			continue
		}
		r, size := utf8.DecodeRuneInString(line[i:])
		i += size
		if r == ' ' || r == '\t' {
			flush()
			continue
		}
		if !open {
			current = word{prefix: active}
			open = true
		}
		current.text += string(r)
	}
	flush()
	return out
}

// nextActive tracks the SGR state after seq: reset clears it, other
// SGR codes stack, and non-SGR escapes leave styling alone.
func nextActive(active, seq string) string {
	if !strings.HasPrefix(seq, "\x1b[") || !strings.HasSuffix(seq, "m") {
		return active
	}
	switch params := seq[2 : len(seq)-1]; params {
	case "", "0", "00", "000":
		return ""
	default:
		return active + seq
	}
}

// cutCells splits s after at most cells display columns, keeping ANSI
// sequences intact: the head resets the styles it opened and the tail
// re-opens them, so a cut never bleeds color across lines.
func cutCells(s string, cells int) (head, tail string) {
	var b strings.Builder
	active := ""
	width := 0
	i := 0
	for i < len(s) {
		if n := core.EscapeLen(s, i); n > 0 {
			seq := s[i : i+n]
			b.WriteString(seq)
			active = nextActive(active, seq)
			i += n
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if width+core.RuneWidth(r) > cells {
			break
		}
		b.WriteRune(r)
		width += core.RuneWidth(r)
		i += size
	}
	if i >= len(s) {
		return b.String(), ""
	}
	if active != "" {
		b.WriteString("\x1b[0m")
	}
	return b.String(), active + s[i:]
}
