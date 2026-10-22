package termstyle

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/CHE3MZ/terminull/v2/termstyle/internal/color"
)

func TestIsTerminal(t *testing.T) {
	t.Parallel()

	if IsTerminal(nil) {
		t.Fatal("IsTerminal(nil) = true, want false")
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	if IsTerminal(r) || IsTerminal(w) {
		t.Fatal("IsTerminal(pipe) = true, want false: pipes are not terminals")
	}
}

func TestColorNeverOnPipe(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm")
	t.Setenv("CI", "")

	if err := os.Unsetenv("NO_COLOR"); err != nil {
		t.Fatalf("Unsetenv(NO_COLOR) error = %v", err)
	}

	if err := os.Unsetenv("CI"); err != nil {
		t.Fatalf("Unsetenv(CI) error = %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	if Color(w, false) {
		t.Fatal("Color(pipe) = true, want false: output is piped, not a terminal")
	}
}

func TestColorRespectsOverrides(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	t.Setenv("TERM", "xterm")

	if Color(w, true) {
		t.Fatal("Color(noColor=true) = true, want false: --no-color wins")
	}

	t.Setenv("NO_COLOR", "1")

	if Color(w, false) {
		t.Fatal("Color with NO_COLOR set = true, want false")
	}

	if err := os.Unsetenv("NO_COLOR"); err != nil {
		t.Fatalf("Unsetenv(NO_COLOR) error = %v", err)
	}

	t.Setenv("TERM", "dumb")

	if Color(w, false) {
		t.Fatal("Color with TERM=dumb = true, want false")
	}
}

func TestEmoji(t *testing.T) {
	t.Setenv("TERM", "xterm")
	t.Setenv("CI", "")

	if err := os.Unsetenv("CI"); err != nil {
		t.Fatalf("Unsetenv(CI) error = %v", err)
	}

	if _, ok := os.LookupEnv("WT_SESSION"); ok {
		if err := os.Unsetenv("WT_SESSION"); err != nil {
			t.Fatalf("Unsetenv(WT_SESSION) error = %v", err)
		}
	}

	t.Setenv("TERM", "dumb")

	if Emoji() {
		t.Fatal("Emoji() with TERM=dumb = true, want false")
	}

	t.Setenv("TERM", "xterm")
	t.Setenv("CI", "true")

	if Emoji() {
		t.Fatal("Emoji() with CI=true = true, want false")
	}

	if err := os.Unsetenv("CI"); err != nil {
		t.Fatalf("Unsetenv(CI) error = %v", err)
	}

	// Legacy Windows consoles cannot render emoji; Windows Terminal
	// (WT_SESSION set) and all other platforms can. WT_SESSION was
	// unset above, so plain Windows expects false here.
	want := runtime.GOOS != "windows"

	if got := Emoji(); got != want {
		t.Fatalf("Emoji() = %v, want %v on %s", got, want, runtime.GOOS)
	}
}

func TestWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"hello", 5},
		{"日本語", 6},
		{"a日b", 4},
	}

	for _, tt := range tests {
		if got := Width(tt.in); got != tt.want {
			t.Errorf("Width(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	if got := Truncate("hi", 8); got != "hi" {
		t.Fatalf("Truncate short = %q, want unchanged", got)
	}

	got := Truncate("hello world", 8)
	if got != "hello..." {
		t.Fatalf("Truncate = %q, want %q", got, "hello...")
	}

	for _, max := range []int{0, 1, 2, 3, 5, 7, 20} {
		res := Truncate("日本語テスト hello", max)
		if Width(res) > max && max >= 3 {
			t.Fatalf("Truncate(max=%d) width = %d, want <= %d (%q)", max, Width(res), max, res)
		}
	}
}

func TestWritersNonNil(t *testing.T) {
	t.Parallel()

	if Stdout() == nil {
		t.Fatal("Stdout() = nil")
	}

	if Stderr() == nil {
		t.Fatal("Stderr() = nil")
	}
}

func TestPaint(t *testing.T) {
	t.Parallel()

	const input = "hello"

	if got := Paint(false, color.FgGreen, input); got != input {
		t.Fatalf("Paint(disabled) = %q, want unchanged", got)
	}

	got := Paint(true, color.FgGreen, input)
	if !strings.Contains(got, input) {
		t.Fatalf("Paint(enabled) = %q, want to contain input", got)
	}

	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("Paint(enabled) = %q, want ANSI escape", got)
	}
}

func TestPaintExactSGR(t *testing.T) {
	t.Parallel()

	if got := Paint(false, FgRed, "x"); got != "x" {
		t.Fatalf("Paint(disabled) = %q, want x", got)
	}

	if got := Paint(true, FgRed, "x"); got != "\x1b[31mx\x1b[0m" {
		t.Fatalf("Paint(FgRed) = %q", got)
	}

	if got := Paint(true, color.Bold, "x"); got != "\x1b[1mx\x1b[22m" {
		t.Fatalf("Paint(Bold) = %q", got)
	}
}
