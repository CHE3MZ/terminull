package progress

import (
	"io"
	"os"
	"time"

	"github.com/briandowns/spinner"
)

// asciiFrames render on every console, including raster fonts.
var asciiFrames = []string{"|", "/", "-", "\\"}

// dotFrames need a modern terminal font (see package doc for the rule).
var dotFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const spinnerSpeed = 100 * time.Millisecond

// Spinner draws a message with an animated prefix until Stop.
type Spinner struct {
	inner *spinner.Spinner
}

// NewSpinner builds a stopped spinner writing to w. Pass style.Emoji()
// as emoji to pick dots vs ASCII frames.
//
// Animation runs only when the destination is a terminal: Start on a
// pipe, file, or CI log is a safe no-op (nothing draws, nothing hangs),
// so callers never branch on interactivity themselves.
func NewSpinner(w io.Writer, message string, emoji bool) *Spinner {
	frames := asciiFrames
	if emoji {
		frames = dotFrames
	}
	opts := []spinner.Option{
		spinner.WithWriter(w),
		spinner.WithSuffix(" " + message),
	}
	// Gate animation on the actual destination when it is a file;
	// otherwise the spinner falls back to os.Stdout for its check.
	if f, ok := w.(*os.File); ok {
		opts = append(opts, spinner.WithWriterFile(f))
	}
	return &Spinner{inner: spinner.New(frames, spinnerSpeed, opts...)}
}

// Start begins animation. The message stays until Stop.
func (s *Spinner) Start() {
	s.inner.Start()
}

// Stop ends animation and restores the line.
func (s *Spinner) Stop() {
	s.inner.Stop()
}

// Active reports whether the spinner is animating.
func (s *Spinner) Active() bool {
	return s.inner.Active()
}
