package termrender

import (
	"strings"
	"testing"
)

func TestEntryTimes(t *testing.T) {
	t.Parallel()

	started := NewLogScopeStarted("build")
	message := NewLogEntryMessage(nil, InfoLevel, "working")
	finished := NewLogScopeFinished(FinishTypeSucceeded, "build")

	if started.GetTime().IsZero() {
		t.Fatal("scope start time is zero")
	}
	if message.GetTime().IsZero() {
		t.Fatal("message time is zero")
	}
	if finished.GetTime().IsZero() {
		t.Fatal("scope finish time is zero")
	}
	if message.IsRaw() {
		t.Fatal("formatted message marked raw")
	}
}

func TestEntryMessageNormalizesTrailingNewlines(t *testing.T) {
	t.Parallel()

	message := NewLogEntryMessage(nil, InfoLevel, "hello\n\n")
	if got := message.GetMessage(); got != "hello\n" {
		t.Fatalf("GetMessage() = %q, want one trailing newline", got)
	}

	raw := NewLogEntryMessage(nil, InfoLevel, "hello\n\n")
	raw.raw = true
	if got := raw.GetMessage(); got != "hello\n\n" {
		t.Fatalf("raw GetMessage() = %q, want bytes unchanged", got)
	}
}

func TestLogLevelStringRoundTrip(t *testing.T) {
	t.Parallel()

	levels := []LogLevel{ErrorLevel, WarnLevel, InfoLevel, DebugLevel, TraceLevel}
	for _, level := range levels {
		got, err := ParseLogLevel(level.String())
		if err != nil {
			t.Fatalf("ParseLogLevel(%q) error = %v", level, err)
		}
		if got != level {
			t.Fatalf("round trip %v -> %q -> %v", level, level.String(), got)
		}
	}
}

func TestParseLogLevelAliases(t *testing.T) {
	t.Parallel()

	tests := map[string]LogLevel{
		"ERROR":   ErrorLevel,
		"err":     ErrorLevel,
		"warning": WarnLevel,
		" INFO ":  InfoLevel,
		"debug":   DebugLevel,
		"trace":   TraceLevel,
	}
	for in, want := range tests {
		got, err := ParseLogLevel(in)
		if err != nil {
			t.Fatalf("ParseLogLevel(%q) error = %v", in, err)
		}
		if got != want {
			t.Fatalf("ParseLogLevel(%q) = %v, want %v", in, got, want)
		}
	}

	_, err := ParseLogLevel("loud")
	if err == nil || !strings.Contains(err.Error(), "error|warn|info|debug|trace") {
		t.Fatalf("ParseLogLevel(loud) error = %v, want choices", err)
	}
}
