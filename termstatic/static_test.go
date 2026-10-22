package termstatic

import (
	"strings"
	"testing"
)

func TestBannerPlain(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	if err := Banner(&b, "demo", "1.2.3", false); err != nil {
		t.Fatalf("Banner() error = %v", err)
	}

	got := b.String()
	if got != "demo 1.2.3\n" {
		t.Fatalf("Banner(plain) = %q", got)
	}
}

func TestBannerColored(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	if err := Banner(&b, "demo", "1.2.3", true); err != nil {
		t.Fatalf("Banner() error = %v", err)
	}

	got := b.String()
	if !strings.Contains(got, "demo") || !strings.Contains(got, "1.2.3") {
		t.Fatalf("Banner(color) = %q, want name and version", got)
	}
	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("Banner(color) = %q, want ANSI escape", got)
	}
}
