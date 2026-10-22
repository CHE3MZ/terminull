package termretry

import (
	"errors"
	"testing"
	"time"
)

func TestDoSuccessFirstTry(t *testing.T) {
	t.Parallel()

	calls := 0
	err := Do(func() error {
		calls++
		return nil
	}, Options{Delay: time.Millisecond})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestDoEventualSuccess(t *testing.T) {
	t.Parallel()

	calls := 0
	err := Do(func() error {
		calls++
		if calls < 3 {
			return errors.New("flake")
		}
		return nil
	}, Options{Attempts: 3, Delay: time.Millisecond})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
}

func TestDoExhausted(t *testing.T) {
	t.Parallel()

	calls := 0
	err := Do(func() error {
		calls++
		return errors.New("down")
	}, Options{Attempts: 2, Delay: time.Millisecond})
	if err == nil || err.Error() != "down" {
		t.Fatalf("Do() error = %v, want last error unwrapped", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestDoRetryIfVeto(t *testing.T) {
	t.Parallel()

	calls := 0
	boom := errors.New("permanent")
	err := Do(func() error {
		calls++
		return boom
	}, Options{
		Attempts: 5,
		Delay:    time.Millisecond,
		RetryIf:  func(err error) bool { return !errors.Is(err, boom) },
	})
	if !errors.Is(err, boom) {
		t.Fatalf("Do() error = %v, want permanent error", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want veto after first try", calls)
	}
}

func TestDoDefaults(t *testing.T) {
	calls := 0
	err := Do(func() error {
		calls++
		return errors.New("down")
	}, Options{Delay: time.Millisecond})
	if err == nil {
		t.Fatal("Do() = nil, want error")
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want default 3 attempts", calls)
	}
}
