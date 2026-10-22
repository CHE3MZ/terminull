package termoutput

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/CHE3MZ/terminull/termoutput/internal/humanize"
	"gopkg.in/yaml.v3"

	"github.com/CHE3MZ/terminull/termcore"
	"github.com/CHE3MZ/terminull/termstyle"
	"github.com/CHE3MZ/terminull/termtable"
)

// Formats for Render and the --output flag.
const (
	FormatTable = "table"
	FormatJSON  = "json"
	FormatYAML  = "yaml"
)

// Render writes v in the named format: table, json, or yaml.
// Tables need a header plus string rows: pass TableValue{Header, Rows}.
// Anything else returns an error (app validates --output earlier too).
func Render(w io.Writer, format string, v any) error {
	switch format {
	case FormatTable:
		tv, ok := v.(TableValue)
		if !ok {
			return fmt.Errorf("output: table format needs TableValue, got %T", v)
		}
		return Table(w, tv.Header, tv.Rows)
	case FormatJSON:
		return JSON(w, v)
	case FormatYAML:
		return YAML(w, v)
	default:
		return fmt.Errorf("output: unknown format %q: want table|json|yaml", format)
	}
}

// TableValue carries a table for Render.
type TableValue struct {
	Header []string
	Rows   [][]string
}

// Table writes a bordered table (StyleBordered default) with a header row.
func Table(w io.Writer, header []string, rows [][]string) error {
	return TableStyle(w, TableValue{Header: header, Rows: rows}, termtable.StyleBordered)
}

// TableWith renders like Table but styles the header through scheme:
// uppercased and underlined when enabled, verbatim otherwise so piped
// output stays script-friendly. termtable measures ANSI-aware, so
// styled headers stay aligned.
func TableWith(w io.Writer, tv TableValue, scheme termstyle.Scheme) error {
	return TableStyleWith(w, tv, termtable.StyleBordered, scheme)
}

// TableStyleWith renders like TableStyle but styles the header through
// scheme, on any termtable style. Borderless and grid styles both
// measure visible width, so styled headers stay aligned either way.
func TableStyleWith(w io.Writer, tv TableValue, s termtable.Style, scheme termstyle.Scheme) error {
	header := tv.Header
	if scheme.Enabled {
		header = make([]string, len(tv.Header))
		for i, h := range tv.Header {
			header[i] = scheme.TableHeader(h)
		}
	}
	tb := termtable.New(w, termtable.WithStyle(s))
	tb.SetHeaders(header...)
	for _, row := range tv.Rows {
		tb.AddRow(row...)
	}
	return tb.Render()
}

// TableStyle renders with an explicit termtable style: StyleBordered
// (bordered, the default look) or StyleBorderless (borderless, most
// script-friendly).
func TableStyle(w io.Writer, tv TableValue, s termtable.Style) error {
	tb := termtable.New(w, termtable.WithStyle(s))
	tb.SetHeaders(tv.Header...)
	for _, row := range tv.Rows {
		tb.AddRow(row...)
	}
	return tb.Render()
}

// JSON writes indented JSON plus a trailing newline. Map keys marshal
// in sorted order, so output is stable across runs.
func JSON(w io.Writer, v any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("output: encode json: %w", err)
	}
	_, err := io.Copy(w, &buf)
	return err
}

// YAML writes YAML plus a trailing newline with stable key order.
func YAML(w io.Writer, v any) (err error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("output: encode yaml: %v", r)
		}
	}()
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("output: encode yaml: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("output: close yaml: %w", err)
	}
	_, err = io.Copy(w, &buf)
	return err
}

// Graph renders values as a plain-ASCII point chart with a labeled
// axis: pure ASCII, no color codes, safe on every console including
// raster fonts. Non-finite values are marked ? and ignored when
// computing the axis.
func Graph(values []float64) string {
	if len(values) == 0 {
		return ""
	}

	const height = 10
	min, max := math.Inf(1), math.Inf(-1)
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	scale := 1.0
	if math.IsInf(min, 1) {
		min, max = 0, 0
	} else if math.IsInf(max-min, 0) {
		scale = math.Max(math.Abs(min), math.Abs(max))
		min /= scale
		max /= scale
	}

	labels := make([]string, height)
	labelWidth := 0
	for r := 0; r < height; r++ {
		label := (max - (max-min)*float64(r)/float64(height-1)) * scale
		labels[r] = fmt.Sprintf("%.2f", label)
		if len(labels[r]) > labelWidth {
			labelWidth = len(labels[r])
		}
	}

	rows := make([][]rune, height)
	for r := range rows {
		rows[r] = []rune(strings.Repeat(" ", len(values)))
	}
	for c, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			rows[height-1][c] = '?'
			continue
		}
		v /= scale
		r := height / 2
		if max > min {
			r = int(math.Round((max - v) * float64(height-1) / (max - min)))
			if r < 0 {
				r = 0
			} else if r >= height {
				r = height - 1
			}
		}
		if r >= 0 && r < height {
			rows[r][c] = '*' // #nosec G602 -- c ranges over values; rows are len(values) wide.
		}
	}

	var b strings.Builder
	for r := 0; r < height; r++ {
		fmt.Fprintf(&b, "%*s |%s\n", labelWidth, labels[r], string(rows[r]))
	}
	return b.String()
}

// Bytes renders a byte count for humans, e.g. "1.5 kB".
func Bytes(n uint64) string {
	return humanize.Bytes(n)
}

// IBytes renders a byte count in binary units, e.g. "1.0 MiB".
func IBytes(n uint64) string {
	return humanize.IBytes(n)
}

// Split shell-splits a command line, honouring quotes and escapes.
func Split(s string) ([]string, error) {
	parts, err := termcore.ShellSplit(s)
	if err != nil {
		return nil, fmt.Errorf("output: split %q: %w", s, err)
	}
	return parts, nil
}
