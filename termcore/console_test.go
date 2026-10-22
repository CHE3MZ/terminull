package termcore

import "testing"

func TestEnsureConsoleIdempotent(t *testing.T) {
	t.Parallel()

	// Best effort by design: pipes and legacy systems may error, but it
	// must never panic and repeated calls must behave the same.
	_ = EnsureConsole()
	_ = EnsureConsole()
}
