package table

import (
	"io"
	"strings"

	"github.com/CHE3MZ/terminull/termcore"
)

const plainPadding = 2 // spaces after each column

// Plain writes header and rows as padded columns with no borders, the
// shape scripts read best. Every cell widens its column to the longest
// visible value plus padding, so ansi-styled cells stay aligned, and
// multi-line cells split into extra physical rows.
func Plain(w io.Writer, header []string, rows [][]string) error {
	cols := len(header)
	if cols == 0 {
		return nil
	}

	expanded := expandRows(cols, rows)
	widths := make([]int, cols)
	for c, h := range header {
		if size := termcore.VisibleWidth(h) + plainPadding; size > widths[c] {
			widths[c] = size
		}
	}
	for _, row := range expanded {
		for c, cell := range row {
			if size := termcore.VisibleWidth(cell) + plainPadding; size > widths[c] {
				widths[c] = size
			}
		}
	}

	var b strings.Builder
	writeRow := func(row []string) {
		for c := 0; c < cols; c++ {
			cell := cellAt(row, c)
			b.WriteString(cell)
			if pad := widths[c] - termcore.VisibleWidth(cell); pad > 0 {
				b.WriteString(strings.Repeat(" ", pad))
			}
		}
		b.WriteString("\n")
	}

	writeRow(header)
	for _, row := range expanded {
		writeRow(row)
	}

	_, err := io.WriteString(w, b.String())
	return err
}

// expandRows splits cells on newlines into extra physical rows: row i
// of a block holds line i of every cell, blanks when a cell is
// shorter.
func expandRows(cols int, rows [][]string) [][]string {
	var out [][]string
	for _, row := range rows {
		maxLines := 0
		for c := 0; c < cols; c++ {
			if n := strings.Count(cellAt(row, c), "\n"); n > maxLines {
				maxLines = n
			}
		}
		for i := 0; i <= maxLines; i++ {
			physical := make([]string, cols)
			for c := 0; c < cols; c++ {
				physical[c] = cellLine(cellAt(row, c), i)
			}
			out = append(out, physical)
		}
	}
	return out
}

func cellLine(cell string, i int) string {
	if i == 0 && !strings.Contains(cell, "\n") {
		return cell
	}
	lines := strings.Split(cell, "\n")
	if i < len(lines) {
		return lines[i]
	}
	return ""
}
