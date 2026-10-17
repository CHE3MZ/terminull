package termtable

import (
	"errors"
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

func TestStyleMatrix(t *testing.T) {
	t.Parallel()

	header := []string{"NAME", "STATUS"}
	row := []string{"web", "ok"}

	tests := []struct {
		name  string
		style Style
		want  string
	}{
		{
			name:  "bordered",
			style: StyleBordered,
			want: "" +
				"┌──────┬────────┐\n" +
				"│ NAME │ STATUS │\n" +
				"├──────┼────────┤\n" +
				"│ web  │ ok     │\n" +
				"└──────┴────────┘\n",
		},
		{
			name:  "rounded",
			style: StyleRounded,
			want: "" +
				"╭──────┬────────╮\n" +
				"│ NAME │ STATUS │\n" +
				"├──────┼────────┤\n" +
				"│ web  │ ok     │\n" +
				"╰──────┴────────╯\n",
		},
		{
			name:  "ascii",
			style: StyleASCII,
			want: "" +
				"+------+--------+\n" +
				"| NAME | STATUS |\n" +
				"+------+--------+\n" +
				"| web  | ok     |\n" +
				"+------+--------+\n",
		},
		{
			name:  "markdown",
			style: StyleMarkdown,
			want: "" +
				"| NAME | STATUS |\n" +
				"|------|--------|\n" +
				"| web  | ok     |\n",
		},
		{
			name:  "borderless",
			style: StyleBorderless,
			want: "" +
				"NAME  STATUS  \n" +
				"web   ok      \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := render(t, tt.style, header, row)
			if got != tt.want {
				t.Fatalf("Render(%s) = %q, want %q", tt.style, got, tt.want)
			}
		})
	}
}

func TestDefaultIsBordered(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	tb := New(&b)
	tb.SetHeaders("NAME", "STATUS")
	tb.AddRow("web", "ok")
	if err := tb.Render(); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	got := b.String()
	for _, want := range []string{"NAME", "STATUS", "web", "ok", "┌", "│", "─", "└"} {
		if !strings.Contains(got, want) {
			t.Fatalf("default output %q does not contain %q", got, want)
		}
	}
}

func TestRenderNormalizesRaggedRows(t *testing.T) {
	t.Parallel()

	t.Run("bordered", func(t *testing.T) {
		got := render(t, StyleBordered, []string{"NAME", "STATUS"}, []string{"web"}, []string{"proxy", "down", "extra"})

		want := "" +
			"┌───────┬────────┐\n" +
			"│ NAME  │ STATUS │\n" +
			"├───────┼────────┤\n" +
			"│ web   │        │\n" +
			"├───────┼────────┤\n" +
			"│ proxy │ down   │\n" +
			"└───────┴────────┘\n"

		if got != want {
			t.Fatalf("Render() = %q, want %q", got, want)
		}
	})

	t.Run("borderless", func(t *testing.T) {
		got := render(t, StyleBorderless, []string{"NAME", "STATUS"}, []string{"web"}, []string{"proxy", "down", "extra"})

		want := "" +
			"NAME   STATUS  \n" +
			"web            \n" +
			"proxy  down    \n"

		if got != want {
			t.Fatalf("Render() = %q, want %q", got, want)
		}
	})
}

func TestEmptyRendersNothing(t *testing.T) {
	t.Parallel()

	for style := StyleBordered; style <= StyleBorderless; style++ {
		var b strings.Builder
		if err := New(&b, WithStyle(style)).Render(); err != nil {
			t.Fatalf("Render(%s) error = %v", style, err)
		}
		if b.String() != "" {
			t.Fatalf("empty Render(%s) = %q, want silence", style, b.String())
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

func TestRenderPropagatesWriteError(t *testing.T) {
	t.Parallel()

	for style := StyleBordered; style <= StyleBorderless; style++ {
		tb := New(errorWriter{}, WithStyle(style))
		tb.SetHeaders("NAME")
		tb.AddRow("web")
		if err := tb.Render(); err == nil {
			t.Fatalf("Render(%s) = nil, want write error", style)
		}
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("boom")
}

func TestStyleString(t *testing.T) {
	t.Parallel()

	want := map[Style]string{
		StyleBordered:   "bordered",
		StyleRounded:    "rounded",
		StyleASCII:      "ascii",
		StyleMarkdown:   "markdown",
		StyleBorderless: "borderless",
	}
	for style, name := range want {
		if got := style.String(); got != name {
			t.Fatalf("Style(%d).String() = %q, want %q", int(style), got, name)
		}
	}
	if Style(99).String() == "" {
		t.Fatal("unknown style string is empty")
	}
}
