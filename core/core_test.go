package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func TestExitCodeOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"plain", errors.New("boom"), 1},
		{"direct", NewExitCodeError(3, errors.New("boom")), 3},
		{"wrapped", fmt.Errorf("ctx: %w", NewExitCodeError(7, errors.New("boom"))), 7},
		{"canceled", context.Canceled, 130},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ExitCodeOf(tt.err); got != tt.want {
				t.Fatalf("ExitCodeOf() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestExitCodeErrorUnwrap(t *testing.T) {
	t.Parallel()

	cause := errors.New("root cause")
	err := NewExitCodeError(2, cause)

	if !errors.Is(err, cause) {
		t.Fatal("errors.Is should see through ExitCodeError to the cause")
	}

	if err.Error() != "root cause" {
		t.Fatalf("Error() = %q, want %q", err.Error(), "root cause")
	}
}

func TestSignalContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := SignalContext(context.Background())
	defer cancel()

	select {
	case <-ctx.Done():
		t.Fatal("fresh SignalContext should not be cancelled")
	default:
	}

	cancel()

	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("SignalContext did not honour cancel")
	}
}

func TestFullVersionFormat(t *testing.T) {
	t.Parallel()

	got := FullVersion()
	if got == "" || got == "-" {
		t.Fatalf("FullVersion() = %q, want non-empty Version-Commit", got)
	}
}

func TestSetDebug(t *testing.T) {
	previous := Level.Level()
	t.Cleanup(func() { Level.Set(previous) })

	Level.Set(slog.LevelInfo)
	SetDebug(true)

	if Level.Level() != slog.LevelDebug {
		t.Fatalf("SetDebug(true) level = %v, want debug", Level.Level())
	}
}
