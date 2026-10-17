package termtable

import (
	"fmt"
	"io"

	"github.com/CHE3MZ/terminull/termtable/internal/bordered"
	"github.com/CHE3MZ/terminull/termtable/internal/plain"
)

// Style selects the table look. New styles arrive as new values; unknown
// values fail Render instead of silently picking one.
type Style int

const (
	// StyleBordered is a unicode box table with full borders. Default.
	StyleBordered Style = iota
	// StyleRounded is StyleBordered with rounded corners.
	StyleRounded
	// StyleASCII is StyleBordered with plain ascii glyphs.
	StyleASCII
	// StyleMarkdown renders a pipe table with a header rule, for docs.
	StyleMarkdown
	// StyleBorderless renders padded columns without borders for scripts.
	StyleBorderless
)

// String names the style for errors and docs.
func (s Style) String() string {
	switch s {
	case StyleBordered:
		return "bordered"
	case StyleRounded:
		return "rounded"
	case StyleASCII:
		return "ascii"
	case StyleMarkdown:
		return "markdown"
	case StyleBorderless:
		return "borderless"
	default:
		return "unknown"
	}
}

// Option tunes a Table at construction.
type Option func(*Table)

// WithStyle selects the look. Default StyleBordered.
func WithStyle(s Style) Option {
	return func(t *Table) {
		t.style = s
	}
}

// Table buffers a header plus string rows and renders on demand.
type Table struct {
	w      io.Writer
	style  Style
	header []string
	rows   [][]string
}

// New builds a Table writing to w with StyleBordered unless overridden.
func New(w io.Writer, opts ...Option) *Table {
	t := &Table{w: w, style: StyleBordered}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// SetHeaders replaces the header row.
func (t *Table) SetHeaders(cols ...string) {
	t.header = cols
}

// AddRow appends one row.
func (t *Table) AddRow(cols ...string) {
	t.rows = append(t.rows, cols)
}

// AddRows appends many rows.
func (t *Table) AddRows(rows ...[]string) {
	t.rows = append(t.rows, rows...)
}

// Render draws the table with the selected style. Rows normalize to the
// header width (short rows pad empty, long rows truncate), matching
// script-friendly expectations on every style.
func (t *Table) Render() error {
	if t.style < StyleBordered || t.style > StyleBorderless {
		return fmt.Errorf("termtable: unknown style %d", int(t.style))
	}
	if len(t.header) == 0 && len(t.rows) == 0 {
		return nil
	}

	rows := normalize(t.header, t.rows)
	if t.style == StyleBorderless {
		pt := plain.New(stringsToAny(t.header)...).WithWriter(t.w)
		for _, row := range rows {
			pt.AddRow(stringsToAny(row)...)
		}
		return pt.Print()
	}

	bt := bordered.New(t.w)
	bt.SetHeaders(t.header...)
	bt.AddRows(rows...)
	switch t.style {
	case StyleRounded:
		bt.SetDividers(bordered.UnicodeRoundedDividers)
	case StyleASCII:
		bt.SetDividers(bordered.ASCIIDividers)
	case StyleMarkdown:
		bt.SetDividers(bordered.MarkdownDividers)
		bt.SetBorderTop(false)
		bt.SetBorderBottom(false)
		bt.SetRowLines(false)
		left := make([]bordered.Alignment, len(t.header))
		for i := range left {
			left[i] = bordered.AlignLeft
		}
		bt.SetHeaderAlignment(left...)
	}
	bt.Render()
	return nil
}

// normalize pads short rows and truncates long ones to the header width.
func normalize(header []string, rows [][]string) [][]string {
	out := make([][]string, 0, len(rows))
	for _, row := range rows {
		fixed := make([]string, len(header))
		copy(fixed, row)
		out = append(out, fixed)
	}
	return out
}

func stringsToAny(values []string) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}
