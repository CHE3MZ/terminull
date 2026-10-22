package termcore

import (
	"os"
	"testing"
)

func TestTerminalWidthPipeUsesDefault(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	if got := TerminalWidth(w); got != DefaultTerminalWidth {
		t.Fatalf("TerminalWidth(pipe) = %d, want %d", got, DefaultTerminalWidth)
	}
}

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

func TestCellPixelsPipeIsZero(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	if w, h := CellPixels(w); w != 0 || h != 0 {
		t.Fatalf("CellPixels(pipe) = %dx%d, want 0x0", w, h)
	}
}
