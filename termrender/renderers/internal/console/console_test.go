package console

import (
	"os"
	"testing"
)

// Pipes have no terminal height on any platform.
func TestTerminalHeightPipeIsUnknown(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	if got := TerminalHeight(w); got != -1 {
		t.Fatalf("TerminalHeight(pipe) = %d, want -1", got)
	}
}
