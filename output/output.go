package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/dustin/go-humanize"
	"github.com/google/shlex"
	"github.com/guptarohit/asciigraph"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"gopkg.in/yaml.v3"
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

// Table writes an ASCII-bordered table with a header row.
func Table(w io.Writer, header []string, rows [][]string) error {
	table := tablewriter.NewTable(w,
		tablewriter.WithHeader(header),
		tablewriter.WithRendition(tw.Rendition{
			Symbols: tw.NewSymbols(tw.StyleASCII),
		}),
	)
	for _, row := range rows {
		if err := table.Append(row); err != nil {
			return fmt.Errorf("output: append table row: %w", err)
		}
	}
	return table.Render()
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

// Graph renders values as a plain-ASCII line chart (no color codes,
// safe on every console).
func Graph(values []float64) string {
	return asciigraph.Plot(values)
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
