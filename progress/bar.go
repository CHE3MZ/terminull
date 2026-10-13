package progress

import (
	"io"

	"github.com/schollz/progressbar/v3"
)

// Bar draws a labeled progress bar to w.
type Bar struct {
	inner *progressbar.ProgressBar
}

// NewBar builds a bar for total units with a description prefix.
func NewBar(w io.Writer, total int, description string) *Bar {
	return &Bar{
		inner: progressbar.NewOptions(total,
			progressbar.OptionSetWriter(w),
			progressbar.OptionSetDescription(description),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetPredictTime(false),
		),
	}
}

// Add advances n units, redrawing the bar.
func (b *Bar) Add(n int) error {
	return b.inner.Add(n)
}

// Finish fills the bar and ends the line.
func (b *Bar) Finish() error {
	return b.inner.Finish()
}
