package termtable

import (
	"fmt"
	"io"

	"github.com/CHE3MZ/terminull/termtable/internal/table"
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
	switch t.style {
	case StyleBorderless:
		return table.Plain(t.w, t.header, rows)
	case StyleRounded:
		return table.Grid(t.w, table.Rounded, t.header, rows)
	case StyleASCII:
		return table.Grid(t.w, table.ASCII, t.header, rows)
	case StyleMarkdown:
		return table.Grid(t.w, table.Markdown, t.header, rows)
	default:
		return table.Grid(t.w, table.Bordered, t.header, rows)
	}
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
