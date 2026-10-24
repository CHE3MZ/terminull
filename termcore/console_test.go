package termcore

import (
	"os"
	"testing"
)

func TestEnsureConsoleIdempotent(t *testing.T) {
	t.Parallel()

	// Best effort by design: pipes and legacy systems may error, but it
	// must never panic and repeated calls must behave the same.
	_ = EnsureConsole()
	_ = EnsureConsole()
}

func TestSupportsANSIRejectsNonTerminals(t *testing.T) {
	t.Parallel()

	if SupportsANSI(nil) {
		t.Fatal("SupportsANSI(nil) = true, want false")
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	if SupportsANSI(w) {
		t.Fatal("SupportsANSI(pipe) = true, want false")
	}
}
