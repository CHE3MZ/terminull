package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/CHE3MZ/terminull/style"
)

// defaultBarWidth is used when the destination width is unknown.
const defaultBarWidth = 40

// Bar draws a labeled progress bar to w. Bars render everywhere,
// including pipes and CI logs: every Add prints one updated line, so
// output stays readable without cursor control.
type Bar struct {
	out         io.Writer
	description string
	total       int
	width       int

	mu      sync.Mutex
	scheme  style.Scheme
	current int
	done    bool
}

// NewBar builds a bar for total units with a description prefix.
// Narrow terminals shrink the bar to fit instead of wrapping.
func NewBar(w io.Writer, total int, description string) *Bar {
	width := defaultBarWidth
	if f, ok := w.(*os.File); ok {
		if cols := style.TerminalWidth(f); cols < 80 && cols >= 10 {
			width = cols
		}
	}
	return &Bar{out: w, description: description, total: total, width: width}
}

// WithScheme paints the filled run through scheme; the zero Scheme keeps
// the bar plain, so pipes and CI logs stay raw text.
func (b *Bar) WithScheme(scheme style.Scheme) *Bar {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.scheme = scheme
	return b
}

// Add advances n units, redrawing the bar. Negative totals clamp at zero;
// overshoot clamps at the total.
func (b *Bar) Add(n int) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current += n
	if b.current < 0 {
		b.current = 0
	}
	if b.total > 0 && b.current > b.total {
		b.current = b.total
	}
	_, err := fmt.Fprintln(b.out, b.line())
	return err
}

// Finish fills the bar and ends the line. Safe to call twice.
func (b *Bar) Finish() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.total > 0 {
		b.current = b.total
	}
	if b.done {
		return nil
	}
	b.done = true
	_, err := fmt.Fprintln(b.out, b.line())
	return err
}

func (b *Bar) line() string {
	percent := 100
	filled := b.width
	if b.total > 0 {
		percent = b.current * 100 / b.total
		filled = b.current * b.width / b.total
	}
	fill := strings.Repeat("=", filled)
	if filled > 0 {
		fill = b.scheme.Cyan(fill)
	}
	bar := fill + strings.Repeat(" ", b.width-filled)
	count := ""
	if b.total > 0 {
		count = fmt.Sprintf(" (%d/%d)", b.current, b.total)
	}
	desc := b.description
	if desc != "" {
		desc += " "
	}
	return fmt.Sprintf("%s[%s] %d%%%s", desc, bar, percent, count)
}
