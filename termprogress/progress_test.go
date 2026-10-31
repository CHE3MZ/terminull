package termprogress

import (
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/CHE3MZ/terminull/v2/termrender"
	"github.com/CHE3MZ/terminull/v2/termrender/renderers"
	"github.com/CHE3MZ/terminull/v2/termrender/terminal"
	"github.com/CHE3MZ/terminull/v2/termstyle"
)

// guarded serializes concurrent widget writes with test reads.
type guarded struct {
	mu sync.Mutex
	b  strings.Builder
}

func (g *guarded) Write(p []byte) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.b.Write(p)
}

func (g *guarded) String() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.b.String()
}

func TestSpinnerBufferNoop(t *testing.T) {
	t.Parallel()

	// Buffers are not terminals: Start is a safe no-op writing nothing,
	// and Stop on a never-started spinner stays silent.
	for _, frames := range [][]string{nil, CharSets[CharSetBraille]} {
		var out guarded
		s := NewSpinner(&out, "spin-msg", frames)
		s.Start()
		if s.Active() {
			t.Fatal("spinner active on a buffer, want quiet no-op")
		}
		time.Sleep(3 * spinnerSpeed)
		s.Stop()
		if s.Active() {
			t.Fatal("spinner active after Stop")
		}
		if got := out.String(); got != "" {
			t.Fatalf("spinner output %q, want nothing on non-terminals", got)
		}
	}
}

// TestSpinnerPipeFileStaysQuiet pins the terminal check to a pipe, which
// is never a terminal on any platform: Start must be a no-op and Stop
// safe, deterministically, without depending on the test runner's own
// stdout.
func TestSpinnerPipeFileStaysQuiet(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	s := NewSpinner(w, "spin-msg", nil)
	s.Start()
	if s.Active() {
		t.Fatal("spinner active on a pipe, want quiet no-op")
	}
	time.Sleep(2 * spinnerSpeed)
	s.Stop()
	if s.Active() {
		t.Fatal("spinner active after Stop")
	}
}

func TestBarAddFinish(t *testing.T) {
	t.Parallel()

	var out guarded
	bar := NewBar(&out, 10, "bar-desc")
	if err := bar.Add(5); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err := bar.Add(5); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err := bar.Finish(); err != nil {
		t.Fatalf("Finish() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{"bar-desc", "100%"} {
		if !strings.Contains(got, want) {
			t.Fatalf("bar output %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "\x1b") {
		t.Fatalf("bar output %q has escapes without a scheme", got)
	}
}

func TestSpinnerSchemeFrame(t *testing.T) {
	t.Parallel()

	var out guarded
	s := NewSpinner(&out, "spin-msg", nil).WithScheme(termstyle.NewScheme(true))
	if got := s.line("|"); got != "\x1b[36m|\x1b[0m spin-msg" {
		t.Fatalf("spinner frame %q, want cyan frame", got)
	}

	var plain guarded
	p := NewSpinner(&plain, "spin-msg", nil)
	if got := p.line("|"); got != "| spin-msg" {
		t.Fatalf("plain spinner frame %q, want raw frame", got)
	}
}

func TestSpinnerOptionsAndCharsets(t *testing.T) {
	t.Parallel()

	var out guarded
	s := New([]string{"a", "b"}, 25*time.Millisecond,
		WithWriter(&out),
		WithPrefix("pre "),
		WithSuffix(" post"),
		WithFinalMessage("done"),
		WithHiddenCursor(false),
		WithColor(termstyle.FgRed),
	)
	if got := s.line("a"); got != "pre a post" {
		t.Fatalf("spinner line %q, want prefix and suffix", got)
	}
	if s.delay != 25*time.Millisecond {
		t.Fatalf("spinner delay = %v, want 25ms", s.delay)
	}
	if s.hideCursor {
		t.Fatal("spinner hides cursor despite WithHiddenCursor(false)")
	}
	if got := s.scheme.Paint(termstyle.FgRed, "a"); got != "a" {
		t.Fatalf("zero scheme painted %q, want plain text", got)
	}

	if got := CharSets[CharSetBraille][0]; got != "⠋" {
		t.Fatalf("charset 14 frame = %q, want braille dot", got)
	}
	if got := CharSets[CharSetASCII]; len(got) != 4 || got[0] != "|" {
		t.Fatalf("charset 9 = %v, want ASCII spinner", got)
	}
	if got := CharSets[39]; len(got) == 0 {
		t.Fatal("charset 39 is empty, want emoji globe frames")
	}
}

func TestSpinnerFramesAndOptions(t *testing.T) {
	t.Parallel()

	var out guarded
	s := NewSpinner(&out, "spin-msg", CharSets[CharSetBraille],
		WithDelay(25*time.Millisecond),
		WithPrefix("pre "),
		WithHiddenCursor(false),
	)
	if len(s.frames) != len(CharSets[CharSetBraille]) {
		t.Fatalf("frames = %v, want charset 14", s.frames)
	}
	if got := s.line(s.frames[0]); got != "pre "+CharSets[CharSetBraille][0]+" spin-msg" {
		t.Fatalf("spinner line %q, want prefix and message", got)
	}
	if s.delay != 25*time.Millisecond {
		t.Fatalf("spinner delay = %v, want 25ms", s.delay)
	}
	if s.finalMessage != "spin-msg" {
		t.Fatalf("final message = %q, want spin-msg", s.finalMessage)
	}
	if s.hideCursor {
		t.Fatal("spinner hides cursor despite WithHiddenCursor(false)")
	}

	var fallback guarded
	f := NewSpinner(&fallback, "fallback", nil)
	if len(f.frames) != len(asciiFrames) || f.frames[0] != "|" {
		t.Fatalf("frames = %v, want ASCII fallback", f.frames)
	}
}

func TestSpinnerWithTerminalKeepsWriter(t *testing.T) {
	t.Parallel()

	var out guarded
	s := New([]string{"x"}, time.Second,
		WithWriter(&out),
		WithTerminal(os.Stdout),
		WithHiddenCursor(false),
	)
	if s.out != &out {
		t.Fatal("WithTerminal replaced the writer")
	}
	if s.file != os.Stdout {
		t.Fatal("WithTerminal did not set the detection file")
	}
}

func TestSpinnerUpdateAndReverse(t *testing.T) {
	t.Parallel()

	s := New([]string{"a", "b", "c"}, time.Second,
		WithWriter(io.Discard),
		WithHiddenCursor(false),
	)
	s.Reverse()
	if got := strings.Join(s.frames, ""); got != "cba" {
		t.Fatalf("reversed frames %q, want cba", got)
	}

	s.UpdateCharset([]string{"x", "y"})
	if got := strings.Join(s.frames, ""); got != "xy" {
		t.Fatalf("updated frames %q, want xy", got)
	}

	s.UpdateCharset(nil)
	if got := strings.Join(s.frames, ""); got != "xy" {
		t.Fatalf("empty update changed frames to %q, want xy", got)
	}

	s.UpdateSpeed(10 * time.Millisecond)
	if s.delay != 10*time.Millisecond {
		t.Fatalf("delay = %v, want 10ms", s.delay)
	}
}

func TestSpinnerDisplayLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		line string
		cols int
		want int
	}{
		{name: "empty", line: "", cols: 80, want: 1},
		{name: "short", line: "hello", cols: 80, want: 1},
		{name: "exact", line: strings.Repeat("x", 80), cols: 80, want: 1},
		{name: "wrapped", line: strings.Repeat("x", 81), cols: 80, want: 2},
		{name: "newline", line: "a\nb", cols: 80, want: 2},
		{name: "ansi", line: "\x1b[36mhello\x1b[0m", cols: 80, want: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := displayLines(tc.line, tc.cols); got != tc.want {
				t.Fatalf("displayLines(%q, %d) = %d, want %d", tc.line, tc.cols, got, tc.want)
			}
		})
	}
}

func TestSpinnerEraseANSI(t *testing.T) {
	t.Parallel()

	var out guarded
	s := New([]string{"x"}, time.Second, WithWriter(&out), WithHiddenCursor(false))
	s.ansi = true
	s.lastLines = 2

	if got, want := s.eraseStringLocked(), "\x1b[K\x1b[A\r\x1b[K"; got != want {
		t.Fatalf("erase sequence %q, want %q", got, want)
	}
	if s.lastLines != 0 {
		t.Fatalf("erase left lastLines = %d, want 0", s.lastLines)
	}
}

type writeCounter struct {
	writes int
}

func (w *writeCounter) Write(p []byte) (int, error) {
	w.writes++
	return len(p), nil
}

func TestSpinnerDrawIsSingleWrite(t *testing.T) {
	t.Parallel()

	var out writeCounter
	s := New([]string{"a", "b"}, time.Second, WithWriter(&out), WithHiddenCursor(false))
	s.active = true
	s.drawLocked()
	s.drawLocked()

	if out.writes != 2 {
		t.Fatalf("draws = %d writes, want 2 (one per frame)", out.writes)
	}
}

func TestSpinnerStartStopFallback(t *testing.T) {
	t.Parallel()

	var out guarded
	s := New([]string{"|", "/"}, 5*time.Millisecond,
		WithWriter(&out),
		WithFinalMessage("done"),
		WithHiddenCursor(false),
	)
	s.tty = true
	s.Start()
	if !s.Active() {
		t.Fatal("spinner not active after Start on simulated terminal")
	}
	time.Sleep(3 * 5 * time.Millisecond)
	s.Stop()

	got := out.String()
	for _, want := range []string{"\r|", "\r \r", "done\n"} {
		if !strings.Contains(got, want) {
			t.Fatalf("spinner output %q does not contain %q", got, want)
		}
	}
	if s.Active() {
		t.Fatal("spinner active after Stop")
	}
}

func TestBarSchemeFill(t *testing.T) {
	t.Parallel()

	var out guarded
	bar := NewBar(&out, 4, "bar-desc").WithScheme(termstyle.NewScheme(true))
	if err := bar.Add(2); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "\x1b[36m") || !strings.Contains(got, "50%") {
		t.Fatalf("bar output %q misses cyan fill or percent", got)
	}
	if strings.Contains(got, "\x1b[36m ") {
		t.Fatalf("bar output %q paints padding", got)
	}
}

func TestBarWidthNarrowTerminal(t *testing.T) {
	t.Parallel()

	// Narrow terminals shrink the bar; the exact width follows the
	// console, so assert shape (one bracketed run + percent + counts).
	var out guarded
	bar := NewBar(&out, 4, "bar-desc")
	if err := bar.Add(2); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	got := out.String()
	for _, want := range []string{"bar-desc", "[", "]", "50%", "(2/4)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("bar output %q does not contain %q", got, want)
		}
	}
}

// TestSharedWriterSmoke runs a spinner, a bar, and a simple logger on one
// writer concurrently: all three markers must land, proving the shared
// writer contract at smoke level.
func TestSharedWriterSmoke(t *testing.T) {
	t.Parallel()

	var out guarded
	renderer := renderers.NewSimpleRenderer(&out, terminal.NoColorSchema())
	logger := termrender.NewLogger(termrender.InfoLevel, renderer)

	spin := NewSpinner(&out, "spin-msg", nil)
	spin.Start()
	bar := NewBar(&out, 4, "bar-desc")
	task := logger.Scoped("smoke")
	task.Infof("log-line")

	if err := bar.Add(4); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err := bar.Finish(); err != nil {
		t.Fatalf("Finish() error = %v", err)
	}
	task.Finish(true)
	time.Sleep(3 * spinnerSpeed)
	spin.Stop()

	// Let the logger goroutine drain. The spinner is quiet here (shared
	// buffer, no terminal gate): the smoke proves safe coexistence and
	// asserts the always-rendering widgets.
	deadline := time.Now().Add(5 * time.Second)
	for {
		got := out.String()
		if strings.Contains(got, "bar-desc") && strings.Contains(got, "log-line") {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("shared output %q misses a marker", got)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
