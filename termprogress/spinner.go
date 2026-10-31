package termprogress

import (
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/CHE3MZ/terminull/v2/termcore"
	"github.com/CHE3MZ/terminull/v2/termstyle"
)

// asciiFrames render on every console, including raster fonts.
var asciiFrames = []string{"|", "/", "-", "\\"}

const (
	spinnerSpeed = 100 * time.Millisecond
	hideCursor   = "\x1b[?25l"
	showCursor   = "\x1b[?25h"
)

// Option configures a Spinner built by New.
type Option func(*Spinner)

// WithWriter sets the destination. A writer that is not an *os.File is
// treated as non-interactive, so Start stays a safe no-op; pair it with
// WithTerminal (or use WithWriterFile) when wrapping a real terminal.
func WithWriter(w io.Writer) Option {
	return func(s *Spinner) {
		if w == nil {
			s.out = io.Discard
			s.file = nil
			return
		}
		s.out = w
		if f, ok := w.(*os.File); ok {
			s.file = f
			return
		}
		s.file = nil
	}
}

// WithWriterFile sets both the destination and the file used for
// terminal and width detection. Use it for direct file writers.
func WithWriterFile(f *os.File) Option {
	return func(s *Spinner) {
		if f == nil {
			s.out = io.Discard
			s.file = nil
			return
		}
		s.out = f
		s.file = f
	}
}

// WithTerminal supplies the terminal file used for gating and width
// detection while leaving the writer untouched. Use it when writing
// through a colorable or other wrapper around a terminal.
func WithTerminal(f *os.File) Option {
	return func(s *Spinner) {
		s.file = f
	}
}

// WithPrefix prepends text before every frame.
func WithPrefix(prefix string) Option {
	return func(s *Spinner) {
		s.prefix = prefix
	}
}

// WithSuffix appends text after every frame.
func WithSuffix(suffix string) Option {
	return func(s *Spinner) {
		s.suffix = suffix
	}
}

// WithFinalMessage replaces the spinner line with message when Stop is
// called. Stop adds a newline. An empty message leaves a cleared line.
func WithFinalMessage(message string) Option {
	return func(s *Spinner) {
		s.finalMessage = message
	}
}

// WithFinalMSG is the briandowns-spinner spelling of WithFinalMessage.
func WithFinalMSG(message string) Option {
	return WithFinalMessage(message)
}

// WithHiddenCursor hides the cursor while the spinner is active when the
// destination supports ANSI. The cursor is restored by Stop.
func WithHiddenCursor(hide bool) Option {
	return func(s *Spinner) {
		s.hideCursor = hide
	}
}

// WithDelay sets the frame interval. Non-positive values keep the
// default 100ms.
func WithDelay(delay time.Duration) Option {
	return func(s *Spinner) {
		if delay > 0 {
			s.delay = delay
		}
	}
}

// WithColor paints frames with attr. Without it, frames use the scheme's
// cyan highlight. The zero Scheme still renders plain text.
func WithColor(attr termstyle.Attribute) Option {
	return func(s *Spinner) {
		s.color = attr
		s.hasColor = true
	}
}

// Spinner draws frames from a caller-provided character set.
//
// It animates only when the destination is a terminal. On capable
// terminals redraws clear wrapped lines and hide the cursor; legacy
// Windows consoles fall back to plain carriage-return redraws. Stop
// always clears the spinner region before writing FinalMessage.
type Spinner struct {
	out          io.Writer
	file         *os.File
	frames       []string
	delay        time.Duration
	prefix       string
	suffix       string
	finalMessage string
	hideCursor   bool
	scheme       termstyle.Scheme
	color        termstyle.Attribute
	hasColor     bool

	mu           sync.Mutex
	tty          bool
	ansi         bool
	enabled      bool
	active       bool
	stopping     bool
	cursorHidden bool
	stop         chan struct{}
	done         chan struct{}
	frameIndex   int
	lastLines    int
	lastWidth    int
}

// New builds a stopped spinner. frames may be any character set; an
// empty slice falls back to ASCII. The writer defaults to stdout, like
// briandowns/spinner, and options can replace it.
func New(frames []string, delay time.Duration, opts ...Option) *Spinner {
	s := &Spinner{
		out:        os.Stdout,
		file:       os.Stdout,
		frames:     cloneFrames(frames),
		delay:      delay,
		hideCursor: true,
		enabled:    true,
	}
	if len(s.frames) == 0 {
		s.frames = cloneFrames(asciiFrames)
	}
	if s.delay <= 0 {
		s.delay = spinnerSpeed
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	s.applyWriter()
	return s
}

// NewSpinner builds a stopped spinner with caller-provided frames and
// options. Empty frames fall back to ASCII. The message is the suffix while
// animating and the final line after Stop. Callers using emoji or wide
// glyphs must gate them with termcore.Emoji.
func NewSpinner(w io.Writer, message string, frames []string, opts ...Option) *Spinner {
	return New(frames, spinnerSpeed,
		append([]Option{
			WithWriter(w),
			WithSuffix(" " + message),
			WithFinalMessage(message),
		}, opts...)...,
	)
}

// WithScheme paints the frame through scheme; the zero Scheme keeps
// frames plain, so widgets stay plain unless callers opt in.
func (s *Spinner) WithScheme(scheme termstyle.Scheme) *Spinner {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scheme = scheme
	return s
}

// Start begins animation. The first frame draws immediately. Safe to
// call twice; only the first starts the loop, and non-terminals stay a
// silent no-op.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active || s.stopping || !s.enabled || !s.tty {
		s.mu.Unlock()
		return
	}

	s.ansi = termcore.SupportsANSI(s.file)
	s.active = true
	s.stop = make(chan struct{})
	s.done = make(chan struct{})
	s.frameIndex = 0
	s.lastLines = 0
	s.lastWidth = 0
	if s.hideCursor && s.ansi {
		_, _ = io.WriteString(s.out, hideCursor)
		s.cursorHidden = true
	}
	s.drawLocked()
	go s.loop(s.stop, s.done)
	s.mu.Unlock()
}

// Stop ends animation, clears the spinner region, restores the cursor,
// and writes the final message. Safe to call twice and without Start.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	s.stopping = true
	done := s.done
	close(s.stop)
	s.mu.Unlock()

	<-done

	s.mu.Lock()
	defer s.mu.Unlock()

	var b strings.Builder
	if seq := s.eraseStringLocked(); seq != "" {
		b.WriteString("\r")
		b.WriteString(seq)
	}
	if s.cursorHidden {
		b.WriteString(showCursor)
		s.cursorHidden = false
	}
	if s.finalMessage != "" {
		b.WriteString(s.finalMessage)
		b.WriteByte('\n')
	}
	if b.Len() > 0 {
		_, _ = io.WriteString(s.out, b.String())
	}
	s.stopping = false
}

// Active reports whether the spinner is animating.
func (s *Spinner) Active() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// Enabled reports whether Start will animate.
func (s *Spinner) Enabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enabled
}

// Enable enables the spinner and restarts it.
func (s *Spinner) Enable() {
	s.mu.Lock()
	s.enabled = true
	s.mu.Unlock()
	s.Restart()
}

// Disable stops the spinner and makes Start a no-op until Enable.
func (s *Spinner) Disable() {
	s.mu.Lock()
	s.enabled = false
	s.mu.Unlock()
	s.Stop()
}

// Restart stops and starts the spinner with the current settings.
func (s *Spinner) Restart() {
	s.Stop()
	s.Start()
}

// Reverse reverses the frame order.
func (s *Spinner) Reverse() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, j := 0, len(s.frames)-1; i < j; i, j = i+1, j-1 {
		s.frames[i], s.frames[j] = s.frames[j], s.frames[i]
	}
}

// UpdateSpeed changes the frame interval. A non-positive value is
// ignored; the active loop picks up the new delay after the next frame.
func (s *Spinner) UpdateSpeed(delay time.Duration) {
	if delay <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delay = delay
}

// UpdateCharset changes the frames. An empty slice is ignored.
func (s *Spinner) UpdateCharset(frames []string) {
	if len(frames) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.frames = cloneFrames(frames)
}

// UpdateCharSet is the briandowns-spinner spelling of UpdateCharset.
func (s *Spinner) UpdateCharSet(frames []string) {
	s.UpdateCharset(frames)
}

func (s *Spinner) applyWriter() {
	if s.file == nil {
		if f, ok := s.out.(*os.File); ok {
			s.file = f
		}
	}
	s.tty = s.file != nil && termcore.IsTerminal(s.file)
}

func (s *Spinner) loop(stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	for {
		s.mu.Lock()
		delay := s.delay
		s.mu.Unlock()
		if delay <= 0 {
			delay = spinnerSpeed
		}

		timer := time.NewTimer(delay)
		select {
		case <-stop:
			stopTimer(timer)
			return
		case <-timer.C:
			s.drawIfActive()
		}
	}
}

func stopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

func (s *Spinner) drawIfActive() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active || s.stopping {
		return
	}
	s.drawLocked()
}

// drawLocked writes one frame line. Caller must hold s.mu.
func (s *Spinner) drawLocked() {
	if !s.active || len(s.frames) == 0 {
		return
	}
	if s.frameIndex >= len(s.frames) {
		s.frameIndex = 0
	}
	line := s.line(s.frames[s.frameIndex])
	s.frameIndex++

	var b strings.Builder
	b.WriteString("\r")
	b.WriteString(s.eraseStringLocked())
	b.WriteString(line)
	_, _ = io.WriteString(s.out, b.String())
	s.lastWidth = termcore.VisibleWidth(line)
	s.lastLines = displayLines(line, s.columns())
}

func (s *Spinner) line(frame string) string {
	return s.prefix + s.frameText(frame) + s.suffix
}

func (s *Spinner) frameText(frame string) string {
	if s.hasColor {
		return s.scheme.Paint(s.color, frame)
	}
	return s.scheme.Cyan(frame)
}

func (s *Spinner) columns() int {
	if s.file == nil {
		return termcore.DefaultTerminalWidth
	}
	return termcore.TerminalWidth(s.file)
}

// eraseStringLocked returns the sequence that clears the previous frame
// and leaves the cursor at column zero. Caller must hold s.mu.
func (s *Spinner) eraseStringLocked() string {
	if s.lastLines == 0 {
		return ""
	}
	lines, width := s.lastLines, s.lastWidth
	s.lastLines = 0
	s.lastWidth = 0
	if s.ansi {
		var b strings.Builder
		b.WriteString("\x1b[K")
		for i := 1; i < lines; i++ {
			b.WriteString("\x1b[A\r\x1b[K")
		}
		return b.String()
	}
	return strings.Repeat(" ", width) + "\r"
}

func displayLines(line string, cols int) int {
	if cols < 1 {
		cols = termcore.DefaultTerminalWidth
	}
	total := 0
	for _, part := range strings.Split(line, "\n") {
		width := termcore.VisibleWidth(part)
		lines := 1
		if width > 0 {
			lines = (width + cols - 1) / cols
		}
		total += lines
	}
	if total < 1 {
		return 1
	}
	return total
}

func cloneFrames(frames []string) []string {
	if len(frames) == 0 {
		return nil
	}
	out := make([]string, len(frames))
	copy(out, frames)
	return out
}
