package table

import (
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/CHE3MZ/terminull/termcore"
)

const (
	padding  = 1  // spaces around each cell
	wrapAt   = 60 // column wrap width once tables outgrow the screen
	fallback = 80 // assumed columns for non-tty writers
)

const (
	alignLeft = iota
	alignCenter
)

// glyphs are the pieces of one grid variant, each one cell wide.
type glyphs struct {
	topLeft     string
	topRight    string
	bottomLeft  string
	bottomRight string
	horizontal  string
	vertical    string
	middleLeft  string
	middleRight string
	topJoint    string
	bottomJoint string
	cross       string
}

// GridConfig selects a grid look.
type GridConfig struct {
	Top        bool // draw the outer top border
	Bottom     bool // draw the outer bottom border
	RowLines   bool // draw rules between data rows
	HeaderLeft bool // left-align headers instead of centering
	glyphs     glyphs
}

// Grid presets.
var (
	// Bordered draws a full unicode box.
	Bordered = GridConfig{Top: true, Bottom: true, RowLines: true, glyphs: glyphs{
		topLeft: "┌", topRight: "┐", bottomLeft: "└", bottomRight: "┘",
		horizontal: "─", vertical: "│",
		middleLeft: "├", middleRight: "┤", topJoint: "┬", bottomJoint: "┴", cross: "┼",
	}}
	// Rounded is Bordered with rounded corners.
	Rounded = GridConfig{Top: true, Bottom: true, RowLines: true, glyphs: glyphs{
		topLeft: "╭", topRight: "╮", bottomLeft: "╰", bottomRight: "╯",
		horizontal: "─", vertical: "│",
		middleLeft: "├", middleRight: "┤", topJoint: "┬", bottomJoint: "┴", cross: "┼",
	}}
	// ASCII is Bordered with plain ascii pieces.
	ASCII = GridConfig{Top: true, Bottom: true, RowLines: true, glyphs: glyphs{
		topLeft: "+", topRight: "+", bottomLeft: "+", bottomRight: "+",
		horizontal: "-", vertical: "|",
		middleLeft: "+", middleRight: "+", topJoint: "+", bottomJoint: "+", cross: "+",
	}}
	// Markdown draws a pipe table with a header rule, for docs.
	Markdown = GridConfig{HeaderLeft: true, glyphs: glyphs{
		horizontal: "-", vertical: "|",
		middleLeft: "|", middleRight: "|", cross: "|",
	}}
)

// Grid writes header and rows as a grid table. Cells wrap to their
// contents unless the table outgrows the writer's width: os.Stdout
// measures the terminal, anything else assumes eighty columns, and
// oversized cells then wrap at word boundaries with hyphens.
func Grid(w io.Writer, cfg GridConfig, header []string, rows [][]string) error {
	return gridTo(w, availableWidth(w), wrapAt, cfg, header, rows)
}

func availableWidth(w io.Writer) int {
	if w == os.Stdout {
		if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
			return width
		}
	}
	return fallback
}

func gridTo(w io.Writer, available, wrapWidth int, cfg GridConfig, header []string, rows [][]string) error {
	cols := len(header)
	if cols == 0 {
		return nil
	}

	source := make([][]string, 0, len(rows)+1)
	source = append(source, header)
	source = append(source, rows...)

	natural := 0
	for _, row := range source {
		width := 1
		for c := 0; c < cols; c++ {
			width += termcore.Width(cellAt(row, c)) + 2*padding + 1
		}
		if width > natural {
			natural = width
		}
	}
	wrap := available < natural

	var cellRows [][][]string
	for _, row := range source {
		cells := make([][]string, cols)
		for c := 0; c < cols; c++ {
			text := cellAt(row, c)
			limit := termcore.Width(text)
			if wrap {
				limit = wrapWidth
			}
			cells[c] = wrapCell(text, limit)
		}
		cellRows = append(cellRows, cells)
	}

	widths := make([]int, cols)
	heights := make([]int, len(cellRows))
	for r, cells := range cellRows {
		height := 0
		for c, lines := range cells {
			for _, line := range lines {
				if size := termcore.VisibleWidth(line); size > widths[c] {
					widths[c] = size
				}
			}
			if len(lines) > height {
				height = len(lines)
			}
		}
		heights[r] = height
	}

	var b strings.Builder
	rule := func(left, joint, right string) {
		b.WriteString(left)
		for c, width := range widths {
			if c > 0 {
				b.WriteString(joint)
			}
			b.WriteString(strings.Repeat(cfg.glyphs.horizontal, width+2*padding))
		}
		b.WriteString(right)
		b.WriteString("\n")
	}

	if cfg.Top {
		rule(cfg.glyphs.topLeft, cfg.glyphs.topJoint, cfg.glyphs.topRight)
	}
	for r, cells := range cellRows {
		if r == 1 || (r > 1 && cfg.RowLines) {
			rule(cfg.glyphs.middleLeft, cfg.glyphs.cross, cfg.glyphs.middleRight)
		}
		align := alignLeft
		if r == 0 && !cfg.HeaderLeft {
			align = alignCenter
		}
		for y := 0; y < heights[r]; y++ {
			b.WriteString(cfg.glyphs.vertical)
			for c, width := range widths {
				b.WriteString(strings.Repeat(" ", padding))
				b.WriteString(padLine(lineAt(cells[c], y), width, align))
				b.WriteString(strings.Repeat(" ", padding))
				b.WriteString(cfg.glyphs.vertical)
			}
			b.WriteString("\n")
		}
	}
	if cfg.Bottom {
		rule(cfg.glyphs.bottomLeft, cfg.glyphs.bottomJoint, cfg.glyphs.bottomRight)
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func cellAt(row []string, c int) string {
	if c >= len(row) {
		return ""
	}
	return row[c]
}

func lineAt(lines []string, y int) string {
	if y >= len(lines) {
		return ""
	}
	return lines[y]
}

// padLine widens line to width cells; centered lines bias the odd
// spare cell to the right, matching long-standing table output.
func padLine(line string, width, align int) string {
	pad := width - termcore.VisibleWidth(line)
	if pad <= 0 {
		return line
	}
	if align == alignLeft {
		return line + strings.Repeat(" ", pad)
	}
	left := pad / 2
	return strings.Repeat(" ", left) + line + strings.Repeat(" ", pad-left)
}
