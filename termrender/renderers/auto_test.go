package renderers

import (
	"os"
	"testing"

	"github.com/CHE3MZ/terminull/termrender"
)

// Pipes are never terminals: AutoRenderer must pick simple output that
// needs no drawing loop and stops trivially.
func TestAutoRendererPipeIsSimple(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	renderer, stop := AutoRenderer(w, false)
	defer stop()

	if _, ok := renderer.(*SimpleRenderer); !ok {
		t.Fatalf("AutoRenderer(pipe) = %T, want *SimpleRenderer", renderer)
	}

	// Round-trip through the picked renderer must not block or panic.
	logger := termrender.NewLogger(termrender.InfoLevel, renderer)
	task := logger.Scoped("smoke")
	task.Infof("hello")
	task.Finish(true)
}

func TestAutoRendererNoColorPipeIsSimple(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	renderer, stop := AutoRenderer(w, true)
	defer stop()

	if _, ok := renderer.(*SimpleRenderer); !ok {
		t.Fatalf("AutoRenderer(pipe, noColor) = %T, want *SimpleRenderer", renderer)
	}
}
