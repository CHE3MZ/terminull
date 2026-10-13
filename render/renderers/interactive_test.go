package renderers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CHE3MZ/terminull/render"
	"github.com/CHE3MZ/terminull/render/renderers/config"
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
	r.RenderScopeStarted(render.NewLogScopeStarted("a"))
	r.RenderMessage(render.NewLogEntryMessage([]string{"a"}, render.InfoLevel, "working"))

	r.DrawFrame()
	grown := readFrame(t, path)
	if grown == "" {
		t.Fatal("DrawFrame wrote nothing with an open scope")
	}

	r.PauseDrawing()
	r.RenderScopeStarted(render.NewLogScopeStarted("b"))
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
	r.RenderScopeStarted(render.NewLogScopeStarted("a"))
	r.PauseDrawing()
	r.StopDrawing()

	if got := readFrame(t, path); !strings.Contains(got, "a") {
		t.Fatalf("final frame %q misses scope", got)
	}
}
