package termoutput

import (
	"strings"
	"testing"

	"github.com/CHE3MZ/terminull/termstyle"
	"github.com/CHE3MZ/terminull/termtable"
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
	// Default style is bordered: unicode borders.
	for _, want := range []string{"┌", "│", "─"} {
		if !strings.Contains(got, want) {
			t.Fatalf("table %q does not contain border %q", got, want)
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
	if err := TableWith(&plain, tv, termstyle.NewScheme(false)); err != nil {
		t.Fatalf("TableWith() error = %v", err)
	}
	// Disabled scheme keeps headers verbatim with no escapes.
	if got := plain.String(); !strings.Contains(got, "name") || strings.Contains(got, "\x1b[") {
		t.Fatalf("TableWith(disabled) = %q, want verbatim header, no escapes", got)
	}

	var styled strings.Builder
	if err := TableWith(&styled, tv, termstyle.NewScheme(true)); err != nil {
		t.Fatalf("TableWith() error = %v", err)
	}
	got := styled.String()
	if !strings.Contains(got, "NAME") || !strings.Contains(got, "\x1b[") {
		t.Fatalf("TableWith(enabled) = %q, want uppercased styled header", got)
	}
}

func TestTableStyleWithScheme(t *testing.T) {
	t.Parallel()

	tv := TableValue{Header: []string{"name", "status"}, Rows: [][]string{{"web", "ok"}}}

	var styled strings.Builder
	if err := TableStyleWith(&styled, tv, termtable.StyleBorderless, termstyle.NewScheme(true)); err != nil {
		t.Fatalf("TableStyleWith() error = %v", err)
	}
	got := styled.String()
	if !strings.Contains(got, "\x1b[4mNAME") || !strings.Contains(got, "web") {
		t.Fatalf("TableStyleWith(enabled) = %q, want styled header", got)
	}

	var plain strings.Builder
	if err := TableStyleWith(&plain, tv, termtable.StyleBorderless, termstyle.NewScheme(false)); err != nil {
		t.Fatalf("TableStyleWith() error = %v", err)
	}
	if got := plain.String(); !strings.Contains(got, "name") || strings.Contains(got, "\x1b[") {
		t.Fatalf("TableStyleWith(disabled) = %q, want verbatim header, no escapes", got)
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

func TestTableStyleBorderless(t *testing.T) {
	t.Parallel()

	tv := TableValue{Header: []string{"NAME", "STATUS"}, Rows: [][]string{{"web", "ok"}}}

	var b strings.Builder
	if err := TableStyle(&b, tv, termtable.StyleBorderless); err != nil {
		t.Fatalf("TableStyle() error = %v", err)
	}
	got := b.String()
	for _, want := range []string{"NAME", "STATUS", "web", "ok"} {
		if !strings.Contains(got, want) {
			t.Fatalf("borderless table %q does not contain %q", got, want)
		}
	}
	for _, box := range []string{"┌", "┐", "└", "┘", "├", "┤", "─", "│", "+"} {
		if strings.Contains(got, box) {
			t.Fatalf("borderless table %q contains border %q, want borderless", got, box)
		}
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
