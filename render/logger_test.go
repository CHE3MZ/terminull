package render

import (
	"slices"
	"sync"
	"testing"
	"time"
)

type recordedEvent struct {
	kind    string // "start", "finish", "msg"
	scopes  []string
	finish  FinishType
	level   LogLevel
	message string
}

type recorder struct {
	mu     sync.Mutex
	events []recordedEvent
}

func (r *recorder) RenderScopeStarted(entry *LogScopeStarted) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, recordedEvent{kind: "start", scopes: entry.GetScopes()})
}

func (r *recorder) RenderScopeFinished(entry *LogScopeFinished) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, recordedEvent{kind: "finish", scopes: entry.GetScopes(), finish: entry.FinishType()})
}

func (r *recorder) RenderMessage(entry *LogEntryMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, recordedEvent{kind: "msg", scopes: entry.GetScopes(), level: entry.Level, message: entry.GetMessage()})
}

func (r *recorder) snapshot() []recordedEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]recordedEvent(nil), r.events...)
}

// waitFor polls until cond holds or the deadline hits. The logger drains
// its channel on a separate goroutine, so tests must synchronize.
func (r *recorder) waitFor(t *testing.T, cond func([]recordedEvent) bool) []recordedEvent {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if got := r.snapshot(); cond(got) {
			return got
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for events, got %+v", r.snapshot())
	return nil
}

func TestLoggerScopeMessageFinishOrder(t *testing.T) {
	t.Parallel()

	rec := &recorder{}
	root := NewLogger(InfoLevel, rec)
	child := root.Scoped("build")
	child.Infof("hello %s", "world")
	child.Finish(true)

	got := rec.waitFor(t, func(events []recordedEvent) bool { return len(events) == 3 })
	if len(got) != 3 {
		t.Fatalf("got %d events, want 3", len(got))
	}

	if got[0].kind != "start" || !slices.Equal(got[0].scopes, []string{"build"}) {
		t.Fatalf("event[0] = %+v, want start [build]", got[0])
	}
	if got[1].kind != "msg" || got[1].message != "hello world\n" || got[1].level != InfoLevel {
		t.Fatalf("event[1] = %+v, want msg hello world at info", got[1])
	}
	if got[2].kind != "finish" || got[2].finish != FinishTypeSucceeded {
		t.Fatalf("event[2] = %+v, want succeeded finish", got[2])
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	t.Parallel()

	rec := &recorder{}
	logger := NewLogger(WarnLevel, rec)
	logger.Debugf("dropped")
	logger.Infof("dropped")
	logger.Warnf("kept")
	logger.Errorf("kept")

	got := rec.waitFor(t, func(events []recordedEvent) bool { return len(events) == 2 })

	for _, e := range got {
		if e.kind != "msg" || (e.level != WarnLevel && e.level != ErrorLevel) {
			t.Fatalf("unexpected event %+v, want only warn/error", e)
		}
	}
}

func TestLoggerAsWriterRaw(t *testing.T) {
	t.Parallel()

	rec := &recorder{}
	logger := NewLogger(InfoLevel, rec)

	w := logger.AsWriter(InfoLevel)
	n, err := w.Write([]byte("raw%content"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != len("raw%content") {
		t.Fatalf("Write() = %d, want %d", n, len("raw%content"))
	}

	// Raw writer output keeps % and gains no trailing newline.
	got := rec.waitFor(t, func(events []recordedEvent) bool { return len(events) == 1 })
	if got[0].kind != "msg" || got[0].message != "raw%content" {
		t.Fatalf("event = %+v, want raw message verbatim", got[0])
	}
}

func TestLoggerFinishSkipped(t *testing.T) {
	t.Parallel()

	rec := &recorder{}
	logger := NewLogger(InfoLevel, rec).Scoped("task")
	logger.FinishWithType(FinishTypeSkipped)

	got := rec.waitFor(t, func(events []recordedEvent) bool { return len(events) == 2 })
	if got[1].kind != "finish" || got[1].finish != FinishTypeSkipped {
		t.Fatalf("event[1] = %+v, want skipped finish", got[1])
	}
}
