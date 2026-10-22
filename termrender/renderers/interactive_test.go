package renderers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/CHE3MZ/terminull/v2/termrender"
	"github.com/CHE3MZ/terminull/v2/termrender/renderers/config"
)

// frameFile builds an interactive renderer on a temp file (a real
// *os.File, readable back deterministically) without starting the draw
// loop, so DrawFrame calls below are fully synchronous.
func frameFile(t *testing.T) (*InteractiveRenderer, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "frame.txt")
	// #nosec G304 -- test frame file under t.TempDir().
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	cfg := config.NewDefaultSymbolsOnlyRenderingConfig()
	return NewInteractiveRenderer(f, cfg), path
}

func readFrame(t *testing.T, path string) string {
	t.Helper()

	// #nosec G304 -- test frame file under t.TempDir().
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	return string(content)
}

func TestPauseSuppressesDrawFrame(t *testing.T) {
	t.Parallel()

	r, path := frameFile(t)
	r.RenderScopeStarted(termrender.NewLogScopeStarted("a"))
	r.RenderMessage(termrender.NewLogEntryMessage([]string{"a"}, termrender.InfoLevel, "working"))

	r.DrawFrame()
	grown := readFrame(t, path)
	if grown == "" {
		t.Fatal("DrawFrame wrote nothing with an open scope")
	}

	r.PauseDrawing()
	r.RenderScopeStarted(termrender.NewLogScopeStarted("b"))
	r.DrawFrame()
	if got := readFrame(t, path); got != grown {
		t.Fatalf("DrawFrame while paused changed output %q to %q", grown, got)
	}

	r.ResumeDrawing()
	r.DrawFrame()
	got := readFrame(t, path)
	if len(got) <= len(grown) {
		t.Fatalf("DrawFrame after resume did not grow %q", got)
	}
	for _, want := range []string{"a", "b"} {
		if !strings.Contains(got, want) {
			t.Fatalf("resumed frame %q misses %q", got, want)
		}
	}
}

func TestStopDrawingFinalizesPaused(t *testing.T) {
	t.Parallel()

	r, path := frameFile(t)
	r.RenderScopeStarted(termrender.NewLogScopeStarted("a"))
	r.PauseDrawing()
	r.StopDrawing()

	if got := readFrame(t, path); !strings.Contains(got, "a") {
		t.Fatalf("final frame %q misses scope", got)
	}
}

// stopWithTimeout fails instead of hanging forever if StopDrawing deadlocks.
func stopWithTimeout(t *testing.T, r *InteractiveRenderer) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		defer close(done)
		r.StopDrawing()
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("StopDrawing did not return")
	}
}

func TestStopDrawingJoinsLoop(t *testing.T) {
	t.Parallel()

	r, path := frameFile(t)
	r.RenderScopeStarted(termrender.NewLogScopeStarted("a"))
	go r.StartDrawing()
	stopWithTimeout(t, r)

	// The joined loop must not write after StopDrawing returns.
	before := readFrame(t, path)
	time.Sleep(300 * time.Millisecond)
	if got := readFrame(t, path); got != before {
		t.Fatalf("loop wrote after join:\n%q\nvs\n%q", before, got)
	}
	if !strings.Contains(before, "a") {
		t.Fatalf("final frame %q misses scope", before)
	}
}

func TestStopDrawingWithoutStart(t *testing.T) {
	t.Parallel()

	r, _ := frameFile(t)
	stopWithTimeout(t, r)
	stopWithTimeout(t, r)
}

func TestSetupWritesNoWrapOnce(t *testing.T) {
	t.Parallel()

	r, path := frameFile(t)
	r.Setup()
	r.Setup()

	got := readFrame(t, path)
	if got != "\x1b[?7l\x1b[?7l" {
		t.Fatalf("Setup output = %q, want two nowrap escapes", got)
	}
}

func TestLateStartAfterStopSilent(t *testing.T) {
	t.Parallel()

	r, path := frameFile(t)
	r.RenderScopeStarted(termrender.NewLogScopeStarted("a"))
	r.StopDrawing()
	before := readFrame(t, path)

	// A start racing in after a completed stop must not write: the
	// stopping gate returns before any setup or loop ownership.
	r.StartDrawing()
	time.Sleep(200 * time.Millisecond)
	if got := readFrame(t, path); got != before {
		t.Fatalf("late StartDrawing changed %q to %q", before, got)
	}
}
