package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/google/shlex"
	"gopkg.in/yaml.v3"

	"github.com/CHE3MZ/terminull/style"
	"github.com/CHE3MZ/terminull/table"
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

// Table writes a bordered table (StyleAqua default) with a header row.
func Table(w io.Writer, header []string, rows [][]string) error {
	return TableStyle(w, TableValue{Header: header, Rows: rows}, table.StyleAqua)
}

// TableWith renders like Table but styles the header through scheme:
// uppercased and underlined when enabled, verbatim otherwise so piped
// output stays script-friendly. The aqua engine measures ANSI-aware,
// so styled headers stay aligned.
func TableWith(w io.Writer, tv TableValue, scheme style.Scheme) error {
	header := tv.Header
	if scheme.Enabled {
		header = make([]string, len(tv.Header))
		for i, h := range tv.Header {
			header[i] = scheme.TableHeader(h)
		}
	}
	tb := table.New(w)
	tb.SetHeaders(header...)
	for _, row := range tv.Rows {
		tb.AddRow(row...)
	}
	return tb.Render()
}

// TableStyle renders with an explicit table style: StyleAqua (bordered,
// the default look) or StyleRodaine (borderless, most script-friendly).
func TableStyle(w io.Writer, tv TableValue, s table.Style) error {
	tb := table.New(w, table.WithStyle(s))
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
func YAML(w io.Writer, v any) error {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("output: encode yaml: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("output: close yaml: %w", err)
	}
	_, err := io.Copy(w, &buf)
	return err
}

// Graph renders values as a plain-ASCII point chart with a labeled
// axis: pure ASCII, no color codes, safe on every console including
// raster fonts.
func Graph(values []float64) string {
	if len(values) == 0 {
		return ""
	}

	const height = 10
	min, max := values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	labels := make([]string, height)
	labelWidth := 0
	for r := 0; r < height; r++ {
		labels[r] = fmt.Sprintf("%.2f", max-(max-min)*float64(r)/float64(height-1))
		if len(labels[r]) > labelWidth {
			labelWidth = len(labels[r])
		}
	}

	rows := make([][]rune, height)
	for r := range rows {
		rows[r] = []rune(strings.Repeat(" ", len(values)))
	}
	for c, v := range values {
		r := height / 2
		if max > min {
			r = int(math.Round((max - v) * float64(height-1) / (max - min)))
		}
		rows[r][c] = '*'
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
	parts, err := shlex.Split(s)
	if err != nil {
		return nil, fmt.Errorf("output: split %q: %w", s, err)
	}
	return parts, nil
}
