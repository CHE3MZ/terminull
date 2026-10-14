package output

import (
	"strings"
	"testing"

	"github.com/CHE3MZ/terminull/style"
)

func TestTableGolden(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	err := Table(&b, []string{"NAME", "STATUS"}, [][]string{
		{"web", "ok"},
		{"日本db", "ok"},
	})
	if err != nil {
		t.Fatalf("Table() error = %v", err)
	}

	got := b.String()
	for _, want := range []string{"NAME", "STATUS", "web", "ok", "日本db"} {
		if !strings.Contains(got, want) {
			t.Fatalf("table %q does not contain %q", got, want)
		}
	}
	for _, r := range []string{"│", "─", "┌", "┐", "└", "┘", "├", "┤", "┬", "┴", "┼"} {
		if strings.Contains(got, r) {
			t.Fatalf("table %q contains box-drawing %q, want ASCII only", got, r)
		}
	}
}

func TestJSONGolden(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	err := JSON(&b, struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}{Name: "x", Count: 2})
	if err != nil {
		t.Fatalf("JSON() error = %v", err)
	}

	want := "{\n  \"name\": \"x\",\n  \"count\": 2\n}\n"
	if b.String() != want {
		t.Fatalf("JSON() = %q, want %q", b.String(), want)
	}
}

func TestTableWithScheme(t *testing.T) {
	t.Parallel()

	tv := TableValue{Header: []string{"name"}, Rows: [][]string{{"x"}}}

	var plain strings.Builder
	if err := TableWith(&plain, tv, style.NewScheme(false)); err != nil {
		t.Fatalf("TableWith() error = %v", err)
	}
	// tablewriter itself uppercases headers; disabled only guarantees
	// no escapes (header styling is scheme-owned, unit-tested in style).
	if got := plain.String(); !strings.Contains(got, "NAME") || strings.Contains(got, "\x1b[") {
		t.Fatalf("TableWith(disabled) = %q, want plain header, no escapes", got)
	}

	var styled strings.Builder
	if err := TableWith(&styled, tv, style.NewScheme(true)); err != nil {
		t.Fatalf("TableWith() error = %v", err)
	}
	got := styled.String()
	if !strings.Contains(got, "NAME") || !strings.Contains(got, "\x1b[") {
		t.Fatalf("TableWith(enabled) = %q, want uppercased styled header", got)
	}
}

func TestYAMLGolden(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	err := YAML(&b, struct {
		Name  string `yaml:"name"`
		Count int    `yaml:"count"`
	}{Name: "x", Count: 2})
	if err != nil {
		t.Fatalf("YAML() error = %v", err)
	}

	want := "name: x\ncount: 2\n"
	if b.String() != want {
		t.Fatalf("YAML() = %q, want %q", b.String(), want)
	}
}

func TestRenderDispatch(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	err := Render(&b, FormatTable, TableValue{
		Header: []string{"A"},
		Rows:   [][]string{{"1"}},
	})
	if err != nil {
		t.Fatalf("Render(table) error = %v", err)
	}
	if !strings.Contains(b.String(), "A") {
		t.Fatalf("Render(table) = %q", b.String())
	}

	b.Reset()
	if err := Render(&b, FormatJSON, map[string]int{"b": 1}); err != nil {
		t.Fatalf("Render(json) error = %v", err)
	}

	if err := Render(&b, "xml", nil); err == nil {
		t.Fatal("Render(xml) = nil, want error")
	}

	if err := Render(&b, FormatTable, "nope"); err == nil {
		t.Fatal("Render(table, string) = nil, want type error")
	}
}

func TestGraphNonEmpty(t *testing.T) {
	t.Parallel()

	got := Graph([]float64{1, 5, 3, 9, 2})
	if got == "" || !strings.Contains(got, "\n") {
		t.Fatalf("Graph() = %q, want multi-line chart", got)
	}
	if !strings.Contains(got, "9.00") {
		t.Fatalf("Graph() = %q, want max label", got)
	}
}

func TestBytes(t *testing.T) {
	t.Parallel()

	if got := Bytes(0); got != "0 B" {
		t.Fatalf("Bytes(0) = %q", got)
	}
	if got := Bytes(1500); got != "1.5 kB" {
		t.Fatalf("Bytes(1500) = %q", got)
	}
	if got := IBytes(1048576); got != "1.0 MiB" {
		t.Fatalf("IBytes(1MiB) = %q", got)
	}
}

func TestSplit(t *testing.T) {
	t.Parallel()

	got, err := Split(`run "a b" c`)
	if err != nil {
		t.Fatalf("Split() error = %v", err)
	}
	if len(got) != 3 || got[0] != "run" || got[1] != "a b" || got[2] != "c" {
		t.Fatalf("Split() = %q", got)
	}

	if _, err := Split(`run "oops`); err == nil {
		t.Fatal("Split(unclosed) = nil, want error")
	}
}
