package renderers

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/CHE3MZ/terminull/v2/termrender"
	"github.com/CHE3MZ/terminull/v2/termrender/terminal"
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

func TestSimpleRendererStylePrefixes(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 10, 26, 15, 4, 5, 123456000, time.UTC)
	tests := []struct {
		name  string
		style termrender.LogStyle
		at    time.Time
		level termrender.LogLevel
		want  string
	}{
		{"plain", termrender.LogStylePlain, base, termrender.InfoLevel, ""},
		{"time", termrender.LogStyleTime, base, termrender.InfoLevel, "[15:04:05] "},
		{"time-level", termrender.LogStyleTimeLevel, base, termrender.WarnLevel, "[15:04:05] WARN "},
		{"logrus-zero", termrender.LogStyleLogrus, base, termrender.InfoLevel, "INFO[0000] "},
		{"logrus-elapsed", termrender.LogStyleLogrus, base.Add(46 * time.Second), termrender.InfoLevel, "INFO[0046] "},
		{"kubernetes", termrender.LogStyleKubernetes, base, termrender.ErrorLevel, "E1026 15:04:05.123456 "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := SimpleRenderer{style: tt.style, startedAt: base}
			if got := r.prefix(tt.at, tt.level); got != tt.want {
				t.Fatalf("prefix() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSimpleRendererStyledMessage(t *testing.T) {
	t.Parallel()

	var out safeBuffer
	renderer := NewSimpleRendererWithStyle(&out, terminal.NoColorSchema(), termrender.LogStyleLogrus)
	logger := termrender.NewLogger(termrender.InfoLevel, renderer)
	logger.Infof("starting")
	logger.Sync()

	got := out.String()
	if !strings.HasPrefix(got, "INFO[") || !strings.HasSuffix(got, "] starting\n") {
		t.Fatalf("styled output = %q, want logrus prefix plus message", got)
	}
}

func TestSimpleRendererStyledScope(t *testing.T) {
	t.Parallel()

	var out safeBuffer
	renderer := NewSimpleRendererWithStyle(&out, terminal.NoColorSchema(), termrender.LogStyleTime)
	logger := termrender.NewLogger(termrender.InfoLevel, renderer)
	task := logger.Scoped("build")
	task.Infof("compiling")
	task.Finish(true)
	logger.Sync()

	got := out.String()
	for _, want := range []string{"Started 'build'", "compiling", "'build' succeeded in "} {
		if !strings.Contains(got, want) {
			t.Fatalf("styled output %q does not contain %q", got, want)
		}
	}
	if strings.Count(got, "[") < 3 {
		t.Fatalf("styled output %q does not prefix each line", got)
	}
}

func TestSimpleRendererStyledMultiline(t *testing.T) {
	t.Parallel()

	var out safeBuffer
	renderer := NewSimpleRendererWithStyle(&out, terminal.NoColorSchema(), termrender.LogStyleTimeLevel)
	logger := termrender.NewLogger(termrender.InfoLevel, renderer)
	logger.Infof("one\ntwo")
	logger.Sync()

	got := out.String()
	if !strings.Contains(got, "INFO one\n") || !strings.Contains(got, "INFO two\n") {
		t.Fatalf("multiline output = %q, want both lines decorated", got)
	}
}

func TestSimpleRendererStyledRawUnchanged(t *testing.T) {
	t.Parallel()

	var out safeBuffer
	renderer := NewSimpleRendererWithStyle(&out, terminal.NoColorSchema(), termrender.LogStyleLogrus)
	logger := termrender.NewLogger(termrender.InfoLevel, renderer)
	if _, err := logger.AsWriter(termrender.InfoLevel).Write([]byte("raw output")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	logger.Sync()

	if got := out.String(); got != "raw output" {
		t.Fatalf("raw styled output = %q, want undecorated", got)
	}
}
