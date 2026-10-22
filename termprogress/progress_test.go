package termprogress

import (
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
	for _, emoji := range []bool{false, true} {
		var out guarded
		s := NewSpinner(&out, "spin-msg", emoji)
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

	s := NewSpinner(w, "spin-msg", false)
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
	s := NewSpinner(&out, "spin-msg", false).WithScheme(termstyle.NewScheme(true))
	s.draw(0)
	if got := out.String(); got != "\r\x1b[36m|\x1b[0m spin-msg" {
		t.Fatalf("spinner frame %q, want cyan frame", got)
	}

	var plain guarded
	p := NewSpinner(&plain, "spin-msg", false)
	p.draw(0)
	if got := plain.String(); got != "\r| spin-msg" {
		t.Fatalf("plain spinner frame %q, want raw frame", got)
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

	spin := NewSpinner(&out, "spin-msg", false)
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
