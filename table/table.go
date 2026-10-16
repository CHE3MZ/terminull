package table

import (
	"fmt"
	"io"

	"github.com/CHE3MZ/terminull/table/aqua"
	"github.com/CHE3MZ/terminull/table/rodaine"
)

// Style selects the rendering engine. New styles arrive as new values;
// unknown values fail Render instead of silently picking one.
type Style int

const (
	// StyleAqua is bordered unicode output, the default look.
	StyleAqua Style = iota
	// StyleRodaine is borderless padded columns for script-friendly output.
	StyleRodaine
)

// String names the style for errors and docs.
func (s Style) String() string {
	switch s {
	case StyleAqua:
		return "aqua"
	case StyleRodaine:
		return "rodaine"
	default:
		return "unknown"
	}
}

// Option tunes a Table at construction.
type Option func(*Table)

// WithStyle selects the engine. Default StyleAqua.
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

// New builds a Table writing to w with StyleAqua unless overridden.
func New(w io.Writer, opts ...Option) *Table {
	t := &Table{w: w, style: StyleAqua}
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

// Render draws the table with the selected engine. Rows normalize to
// the header width (short rows pad empty, long rows truncate), matching
// script-friendly expectations on both engines.
func (t *Table) Render() error {
	switch t.style {
	case StyleAqua, StyleRodaine:
		if len(t.header) == 0 && len(t.rows) == 0 {
			return nil
		}
	}
	switch t.style {
	case StyleAqua:
		at := aqua.New(t.w)
		at.SetHeaders(t.header...)
		at.AddRows(normalize(t.header, t.rows)...)
		at.Render()
		return nil
	case StyleRodaine:
		rt := rodaine.New(stringsToAny(t.header)...).WithWriter(t.w)
		for _, row := range normalize(t.header, t.rows) {
			rt.AddRow(stringsToAny(row)...)
		}
		return rt.Print()
	default:
		return fmt.Errorf("table: unknown style %d", int(t.style))
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

func stringsToAny(values []string) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}
