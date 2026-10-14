package progress

import (
	"io"
	"os"

	"github.com/schollz/progressbar/v3"

	"github.com/CHE3MZ/terminull/style"
)

// Bar draws a labeled progress bar to w.
type Bar struct {
	inner *progressbar.ProgressBar
}

// NewBar builds a bar for total units with a description prefix.
// Narrow terminals shrink the bar to fit instead of wrapping.
func NewBar(w io.Writer, total int, description string) *Bar {
	opts := []progressbar.Option{
		progressbar.OptionSetWriter(w),
		progressbar.OptionSetDescription(description),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(false),
	}
	if width, ok := barWidth(w); ok {
		opts = append(opts, progressbar.OptionSetWidth(width))
	}
	return &Bar{inner: progressbar.NewOptions(total, opts...)}
}

// barWidth fits the bar when the destination terminal is narrower than
// the default 80 columns. Non-file writers and wide terminals keep the
// default: false means untouched.
func barWidth(w io.Writer) (int, bool) {
	f, ok := w.(*os.File)
	if !ok {
		return 0, false
	}
	if cols := style.TerminalWidth(f); cols < 80 && cols >= 10 {
		return cols, true
	}
	return 0, false
}

// Add advances n units, redrawing the bar.
func (b *Bar) Add(n int) error {
	return b.inner.Add(n)
}

// Finish fills the bar and ends the line.
func (b *Bar) Finish() error {
	return b.inner.Finish()
}
