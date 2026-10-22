package renderers

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/CHE3MZ/terminull/termrender"
	"github.com/CHE3MZ/terminull/termrender/terminal"
)

func TestStubRendererSilent(t *testing.T) {
	t.Parallel()

	stub := &StubRenderer{}
	logger := termrender.NewLogger(termrender.TraceLevel, stub)
	task := logger.Scoped("quiet")
	task.Debugf("dropped")
	task.Finish(true)

	if !logger.IsLogLevelEnabled(termrender.TraceLevel) {
		t.Fatal("IsLogLevelEnabled(trace) = false, want true at trace max")
	}
}

func TestQuotedIfNeeded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"foo", "'foo'"},
		{"task 'foo'", "task 'foo'"},
		{"'foo' task", "'foo' task"},
		{"task \"foo\"", "task \"foo\""},
		{"\"foo\" task", "\"foo\" task"},
	}

	for _, tt := range tests {
		if got := quotedIfNeeded(tt.in); got != tt.want {
			t.Errorf("quotedIfNeeded(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// safeBuffer serializes the logger goroutine's writes with the test's reads.
type safeBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (s *safeBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func TestSimpleRendererGolden(t *testing.T) {
	t.Parallel()

	var out safeBuffer
	renderer := NewSimpleRenderer(&out, terminal.NoColorSchema())

	logger := termrender.NewLogger(termrender.InfoLevel, renderer)
	task := logger.Scoped("build")
	task.Infof("compiling")
	task.Finish(true)

	// The logger drains its channel on a separate goroutine: poll.
	deadline := time.Now().Add(5 * time.Second)
	for {
		got := out.String()
		if strings.Count(got, "\n") >= 3 {
			for _, want := range []string{"Started 'build'", "compiling", "'build' succeeded in "} {
				if !strings.Contains(got, want) {
					t.Fatalf("output %q does not contain %q", got, want)
				}
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for simple output, got %q", got)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
