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

func TestAttributesMatchEngine(t *testing.T) {
	t.Parallel()

	pairs := map[Attribute]color.Attribute{
		FgBlack: color.FgBlack, FgRed: color.FgRed,
		FgGreen: color.FgGreen, FgYellow: color.FgYellow,
		FgBlue: color.FgBlue, FgMagenta: color.FgMagenta,
		FgCyan: color.FgCyan, FgWhite: color.FgWhite,
		FgHiBlack: color.FgHiBlack, FgHiRed: color.FgHiRed,
		FgHiGreen: color.FgHiGreen, FgHiYellow: color.FgHiYellow,
		FgHiBlue: color.FgHiBlue, FgHiMagenta: color.FgHiMagenta,
		FgHiCyan: color.FgHiCyan, FgHiWhite: color.FgHiWhite,
	}
	for ours, engine := range pairs {
		if color.Attribute(ours) != engine {
			t.Fatalf("attribute %d != engine %d", int(ours), int(engine))
		}
	}
}

func TestFaintDisabled(t *testing.T) {
	t.Parallel()

	s := NewScheme(false)
	if got := s.Faint(color.FgRed, "x"); got != "x" {
		t.Fatalf("Faint(disabled) = %q", got)
	}
	if got := s.FaintRGB(255, 0, 0, "x"); got != "x" {
		t.Fatalf("FaintRGB(disabled) = %q", got)
	}
}

func TestFaintRGB(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TERM", "xterm")

	s := NewScheme(true)
	if got := s.FaintRGB(255, 0, 0, "x"); got != "\x1b[38;2;127;0;0mx\x1b[0m" {
		t.Fatalf("FaintRGB(true) = %q", got)
	}

	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "xterm-256color")
	if got := s.FaintRGB(255, 0, 0, "x"); got != "\x1b[38;5;88mx\x1b[0m" {
		t.Fatalf("FaintRGB(256) = %q", got)
	}

	t.Setenv("TERM", "xterm")
	if got := s.FaintRGB(255, 165, 0, "x"); got != "\x1b[31mx\x1b[0m" {
		t.Fatalf("FaintRGB(16) = %q, want nearest base red", got)
	}

	t.Setenv("COLORTERM", "truecolor")
	if got := s.FaintRGB(300, -10, 128, "x"); got != "\x1b[38;2;127;0;64mx\x1b[0m" {
		t.Fatalf("FaintRGB(clamped) = %q", got)
	}
}
