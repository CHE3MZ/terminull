package style

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestSchemeDisabledIsIdentity(t *testing.T) {
	t.Parallel()

	s := NewScheme(false)
	if got := s.Bold("x"); got != "x" {
		t.Fatalf("Bold disabled = %q", got)
	}
	if got := s.TableHeader("Name"); got != "Name" {
		t.Fatalf("TableHeader disabled = %q, want verbatim", got)
	}
	for _, tc := range []struct {
		name string
		got  string
	}{
		{"muted", s.Muted("x")},
		{"red", s.Red("x")},
		{"yellow", s.Yellow("x")},
		{"green", s.Green("x")},
		{"cyan", s.Cyan("x")},
		{"success", s.SuccessIcon()},
		{"failure", s.FailureIcon()},
		{"warning", s.WarningIcon()},
	} {
		if strings.Contains(tc.got, "\x1b[") {
			t.Fatalf("%s disabled = %q, want no escapes", tc.name, tc.got)
		}
	}
	if got := s.SuccessIcon(); got != "✓" {
		t.Fatalf("SuccessIcon disabled = %q", got)
	}
}

func TestSchemeEnabledPaints(t *testing.T) {
	t.Parallel()

	s := NewScheme(true)
	got := s.Bold("x")
	if !strings.Contains(got, "x") || !strings.Contains(got, "\x1b[") {
		t.Fatalf("Bold enabled = %q, want ANSI around content", got)
	}
	if got := s.TableHeader("name"); got == "name" || got == "NAME" {
		t.Fatalf("TableHeader enabled = %q, want styled upper", got)
	}
	if got := s.TableHeader("name"); !strings.Contains(got, "NAME") {
		t.Fatalf("TableHeader enabled = %q, want uppercased", got)
	}
	if s.Paint(color.FgRed, "x") == "x" {
		t.Fatal("Paint enabled returned input unchanged")
	}
}

func TestTerminalWidthFallback(t *testing.T) {
	t.Parallel()

	if got := TerminalWidth(nil); got != DefaultTerminalWidth {
		t.Fatalf("TerminalWidth(nil) = %d, want %d", got, DefaultTerminalWidth)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})
	if got := TerminalWidth(w); got != DefaultTerminalWidth {
		t.Fatalf("TerminalWidth(pipe) = %d, want fallback %d", got, DefaultTerminalWidth)
	}
}

func TestColorFor(t *testing.T) {
	var buf bytes.Buffer
	if ColorFor(&buf, false) {
		t.Fatal("ColorFor(buffer) = true, want false for non-files")
	}
	if ColorFor(os.Stdout, true) {
		t.Fatal("ColorFor(no-color) = true, want false")
	}
}
