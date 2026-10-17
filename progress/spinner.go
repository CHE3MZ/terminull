package progress

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/CHE3MZ/terminull/style"
)

// asciiFrames render on every console, including raster fonts.
var asciiFrames = []string{"|", "/", "-", "\\"}

// dotFrames need a modern terminal font (see package doc for the rule).
var dotFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const spinnerSpeed = 100 * time.Millisecond

// Spinner draws a message with an animated prefix until Stop.
type Spinner struct {
	out     io.Writer
	frames  []string
	message string
	tty     bool

	mu     sync.Mutex
	scheme style.Scheme
	active bool
	stop   chan struct{}
	done   chan struct{}
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
	tty := false
	if f, ok := w.(*os.File); ok {
		tty = style.IsTerminal(f)
	}
	return &Spinner{out: w, frames: frames, message: message, tty: tty}
}

// WithScheme paints the frame through scheme; the zero Scheme keeps
// frames plain, so widgets stay plain unless callers opt in.
func (s *Spinner) WithScheme(scheme style.Scheme) *Spinner {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scheme = scheme
	return s
}

// Start begins animation. The message stays until Stop. Safe to call
// twice; only the first starts the loop.
func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active || !s.tty {
		return
	}
	s.active = true
	s.stop = make(chan struct{})
	s.done = make(chan struct{})
	go s.loop()
}

// Stop ends animation and restores the line. Safe to call twice and
// without Start.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	done := s.done
	close(s.stop)
	s.mu.Unlock()

	<-done
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = fmt.Fprintf(s.out, "\r%s\n", s.message)
}

// Active reports whether the spinner is animating.
func (s *Spinner) Active() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

func (s *Spinner) loop() {
	defer close(s.done)
	ticker := time.NewTicker(spinnerSpeed)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.draw(i)
			i++
		}
	}
}

// draw writes one frame line.
func (s *Spinner) draw(i int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = fmt.Fprintf(s.out, "\r%s %s", s.scheme.Cyan(s.frames[i%len(s.frames)]), s.message)
}
