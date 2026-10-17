package box

import (
	"io"
	"strings"

	"github.com/CHE3MZ/terminull/style"
)

// Style selects the frame glyph set. Unknown values render as rounded.
type Style int

// Frame styles.
const (
	StyleRounded Style = iota
	StyleHard
	StyleDouble
	StyleASCII
	StyleNone
)

// String names the style.
func (s Style) String() string {
	switch s {
	case StyleRounded:
		return "rounded"
	case StyleHard:
		return "hard"
	case StyleDouble:
		return "double"
	case StyleASCII:
		return "ascii"
	case StyleNone:
		return "none"
	default:
		return "unknown"
	}
}

// Options tunes a box. The zero value renders a rounded frame sized to
// the content with no padding.
type Options struct {
	// Style picks the frame glyphs; zero is StyleRounded.
	Style Style
	// Title renders in the top border, e.g. ╭─ Warning ─╮, and is
	// truncated with an ellipsis when the box is too narrow.
	Title string
	// Padding is the inner horizontal spaces between frame and
	// content; negative values clamp to zero.
	Padding int
	// Width is the total outer width including the frame. Zero fits
	// the content; a positive value wraps content to fit, clamping to
	// one content column when smaller than the frame. StyleNone
	// measures Width against the content alone.
	Width int
	// Frame colors the frame glyphs. Zero (the engine's Reset
	// attribute) leaves the frame plain.
	Frame style.Attribute
}

// String renders content as a framed box and always ends the result
// with a newline. Tabs expand to 8 columns and trailing newlines are
// dropped before framing. scheme gates frame colors, so a disabled
// scheme yields plain text.
func String(content string, opts Options, scheme style.Scheme) string {
	padding := opts.Padding
	if padding < 0 {
		padding = 0
	}
	lines := contentLines(content)
	glyphs := frameGlyphs(opts.Style)
	framed := glyphs.vertical != ""
	cols := contentWidth(lines, opts, padding, framed)
	wrapped := wrapLines(lines, cols)
	if !framed {
		return strings.Join(wrapped, "\n") + "\n"
	}
	paint := func(s string) string { return s }
	if opts.Frame != 0 {
		paint = func(s string) string { return scheme.Paint(opts.Frame, s) }
	}
	gap := strings.Repeat(" ", padding)
	left := paint(glyphs.vertical)
	right := paint(glyphs.vertical)
	span := cols + 2*padding
	out := make([]string, 0, len(wrapped)+2)
	out = append(out, paint(topLine(glyphs, opts.Title, span)))
	carry := ""
	for i, line := range wrapped {
		line = carry + line
		carry = ""
		if pending := pendingSGR(line); pending != "" {
			line += "\x1b[0m"
			if i < len(wrapped)-1 {
				carry = pending
			}
		}
		out = append(out, left+gap+padLine(line, cols)+gap+right)
	}
	out = append(out, paint(glyphs.bottomLeft+strings.Repeat(glyphs.horizontal, span)+glyphs.bottomRight))
	return strings.Join(out, "\n") + "\n"
}

// Render writes String(content, opts, scheme) to w.
func Render(w io.Writer, content string, opts Options, scheme style.Scheme) error {
	_, err := io.WriteString(w, String(content, opts, scheme))
	return err
}

type frame struct {
	topLeft     string
	topRight    string
	bottomLeft  string
	bottomRight string
	horizontal  string
	vertical    string
}

func frameGlyphs(s Style) frame {
	switch s {
	case StyleHard:
		return frame{"┌", "┐", "└", "┘", "─", "│"}
	case StyleDouble:
		return frame{"╔", "╗", "╚", "╝", "═", "║"}
	case StyleASCII:
		return frame{"+", "+", "+", "+", "-", "|"}
	case StyleNone:
		return frame{}
	default:
		return frame{"╭", "╮", "╰", "╯", "─", "│"}
	}
}

// topLine draws the top border with an optional title segment
// "─ Title ─" so the line keeps the box width.
func topLine(glyphs frame, title string, span int) string {
	if title != "" && span >= 4 {
		title = style.Truncate(title, span-3)
		tail := span - 3 - style.VisibleWidth(title)
		return glyphs.topLeft + glyphs.horizontal + " " + title + " " +
			strings.Repeat(glyphs.horizontal, tail) + glyphs.topRight
	}
	return glyphs.topLeft + strings.Repeat(glyphs.horizontal, span) + glyphs.topRight
}

// contentWidth picks the content column count: the fixed inner size
// when Width is set, else the widest visible line (at least 1 or the
// title segment, whichever needs more room).
func contentWidth(lines []string, opts Options, padding int, framed bool) int {
	extra := 0
	if framed {
		extra = 2 + 2*padding
	}
	if opts.Width > 0 {
		if cols := opts.Width - extra; cols > 1 {
			return cols
		}
		return 1
	}
	cols := 1
	for _, line := range lines {
		if w := style.VisibleWidth(line); w > cols {
			cols = w
		}
	}
	if framed && opts.Title != "" {
		if min := style.VisibleWidth(opts.Title) + 3 - 2*padding; min > cols {
			cols = min
		}
	}
	return cols
}

func padLine(line string, cols int) string {
	if pad := cols - style.VisibleWidth(line); pad > 0 {
		return line + strings.Repeat(" ", pad)
	}
	return line
}

const tabWidth = 8

func contentLines(content string) []string {
	content = strings.ReplaceAll(content, "\t", strings.Repeat(" ", tabWidth))
	content = strings.TrimRight(content, "\n")
	return strings.Split(content, "\n")
}
