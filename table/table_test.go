package table

import (
	"strings"
	"testing"
)

func render(t *testing.T, style Style, header []string, rows ...[]string) string {
	t.Helper()

	var b strings.Builder
	tb := New(&b, WithStyle(style))
	tb.SetHeaders(header...)
	for _, row := range rows {
		tb.AddRow(row...)
	}
	if err := tb.Render(); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return b.String()
}

func TestDefaultIsAqua(t *testing.T) {
	t.Parallel()

	// Empty tables render nothing instead of tripping engine edge cases.
	var empty strings.Builder
	if err := New(&empty).Render(); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if empty.String() != "" {
		t.Fatalf("empty Render() = %q, want silence", empty.String())
	}

	var def strings.Builder
	tb := New(&def)
	tb.SetHeaders("NAME", "STATUS")
	tb.AddRow("web", "ok")
	if err := tb.Render(); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	got := def.String()
	for _, want := range []string{"NAME", "STATUS", "web", "ok", "┌", "│", "─"} {
		if !strings.Contains(got, want) {
			t.Fatalf("aqua output %q does not contain %q", got, want)
		}
	}
}

func TestRodaineIsBorderless(t *testing.T) {
	t.Parallel()

	got := render(t, StyleRodaine, []string{"NAME", "STATUS"}, []string{"web", "ok"}, []string{"日本db", "ok"})
	for _, want := range []string{"NAME", "STATUS", "web", "ok", "日本db"} {
		if !strings.Contains(got, want) {
			t.Fatalf("rodaine output %q does not contain %q", got, want)
		}
	}
	for _, box := range []string{"┌", "┐", "└", "┘", "├", "┤", "─", "│", "+", "-", "|"} {
		if strings.Contains(got, box) {
			t.Fatalf("rodaine output %q contains border %q, want borderless", got, box)
		}
	}
}

func TestUnknownStyleErrors(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	tb := New(&b, WithStyle(Style(99)))
	if err := tb.Render(); err == nil {
		t.Fatal("Render(unknown) = nil, want error")
	}
}

func TestStyleString(t *testing.T) {
	t.Parallel()

	if StyleAqua.String() != "aqua" || StyleRodaine.String() != "rodaine" {
		t.Fatal("Style.String mismatch")
	}
	if Style(99).String() == "" {
		t.Fatal("unknown style string is empty")
	}
}
