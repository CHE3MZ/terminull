package style

import (
	"testing"

	"github.com/fatih/color"
)

func TestDimCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		attr  color.Attribute
		depth int
		want  string
	}{
		{color.FgRed, depthTrue, "38;2;102;0;0"},
		{color.FgHiGreen, depthTrue, "38;2;0;127;0"},
		{color.FgWhite, depthTrue, "38;2;114;114;114"},
		{color.FgRed, depth256, "38;5;52"},
		{color.FgHiGreen, depth256, "38;5;28"},
		{color.FgHiRed, depth16, "31"},
		{color.FgRed, depth16, "31"},
		{color.Bold, depthTrue, ""},
		{color.Bold, depth16, ""},
	}

	for _, tt := range tests {
		if got := dimCode(tt.attr, tt.depth); got != tt.want {
			t.Errorf("dimCode(%d, %d) = %q, want %q", int(tt.attr), tt.depth, got, tt.want)
		}
	}
}

func TestColorDepth(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TERM", "xterm")
	if got := colorDepth(); got != depthTrue {
		t.Fatalf("colorDepth() = %d, want truecolor", got)
	}

	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "xterm-256color")
	if got := colorDepth(); got != depth256 {
		t.Fatalf("colorDepth() = %d, want 256", got)
	}

	t.Setenv("TERM", "xterm")
	if got := colorDepth(); got != depth16 {
		t.Fatalf("colorDepth() = %d, want 16", got)
	}
}

func TestFaintEnabled(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TERM", "xterm")

	s := NewScheme(true)
	got := s.Faint(color.FgRed, "x")
	want := "\x1b[38;2;102;0;0mx\x1b[0m"
	if got != want {
		t.Fatalf("Faint() = %q, want %q", got, want)
	}
}

func TestFaintDisabled(t *testing.T) {
	t.Parallel()

	s := NewScheme(false)
	if got := s.Faint(color.FgRed, "x"); got != "x" {
		t.Fatalf("Faint(disabled) = %q", got)
	}
}
