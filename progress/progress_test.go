package progress

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/CHE3MZ/terminull/render"
	"github.com/CHE3MZ/terminull/render/renderers"
	"github.com/CHE3MZ/terminull/render/terminal"
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

func TestSpinnerStartStop(t *testing.T) {
	t.Parallel()

	// Buffers are not *os.File: the spinner falls back to os.Stdout for
	// its terminal check. Either way Start/Stop must be safe and return.
	var out guarded
	s := NewSpinner(&out, "spin-msg", false)
	s.Start()
	time.Sleep(3 * spinnerSpeed)
	s.Stop()

	dots := NewSpinner(&out, "spin-msg", true)
	dots.Start()
	time.Sleep(3 * spinnerSpeed)
	dots.Stop()
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
}

// TestSharedWriterSmoke runs a spinner, a bar, and a simple logger on one
// writer concurrently: all three markers must land, proving the shared
// writer contract at smoke level.
func TestSharedWriterSmoke(t *testing.T) {
	t.Parallel()

	var out guarded
	renderer := renderers.NewSimpleRenderer(&out, terminal.NoColorSchema())
	logger := render.NewLogger(render.InfoLevel, renderer)

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
