package table

import (
	"errors"
	"strings"
	"testing"
)

func grid(t *testing.T, cfg GridConfig, header []string, rows ...[]string) string {
	t.Helper()

	var b strings.Builder
	if err := Grid(&b, cfg, header, rows); err != nil {
		t.Fatalf("Grid() error = %v", err)
	}
	return b.String()
}

func TestGridGoldens(t *testing.T) {
	t.Parallel()

	header := []string{"NAME", "STATUS"}
	row := []string{"web", "ok"}

	tests := []struct {
		name string
		cfg  GridConfig
		want string
	}{
		{
			name: "bordered",
			cfg:  Bordered,
			want: "" +
				"┌──────┬────────┐\n" +
				"│ NAME │ STATUS │\n" +
				"├──────┼────────┤\n" +
				"│ web  │ ok     │\n" +
				"└──────┴────────┘\n",
		},
		{
			name: "rounded",
			cfg:  Rounded,
			want: "" +
				"╭──────┬────────╮\n" +
				"│ NAME │ STATUS │\n" +
				"├──────┼────────┤\n" +
				"│ web  │ ok     │\n" +
				"╰──────┴────────╯\n",
		},
		{
			name: "ascii",
			cfg:  ASCII,
			want: "" +
				"+------+--------+\n" +
				"| NAME | STATUS |\n" +
				"+------+--------+\n" +
				"| web  | ok     |\n" +
				"+------+--------+\n",
		},
		{
			name: "markdown",
			cfg:  Markdown,
			want: "" +
				"| NAME | STATUS |\n" +
				"|------|--------|\n" +
				"| web  | ok     |\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := grid(t, tt.cfg, header, row); got != tt.want {
				t.Fatalf("Grid(%s) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestGridRaggedRows(t *testing.T) {
	t.Parallel()

	got := grid(t, Bordered, []string{"NAME", "STATUS"}, []string{"web"}, []string{"proxy", "down", "extra"})
	want := "" +
		"┌───────┬────────┐\n" +
		"│ NAME  │ STATUS │\n" +
		"├───────┼────────┤\n" +
		"│ web   │        │\n" +
		"├───────┼────────┤\n" +
		"│ proxy │ down   │\n" +
		"└───────┴────────┘\n"

	if got != want {
		t.Fatalf("Grid(ragged) = %q, want %q", got, want)
	}
}

func TestGridHeaderOnly(t *testing.T) {
	t.Parallel()

	got := grid(t, Bordered, []string{"NAME"})
	want := "" +
		"┌──────┐\n" +
		"│ NAME │\n" +
		"└──────┘\n"

	if got != want {
		t.Fatalf("Grid(header only) = %q, want %q", got, want)
	}
}

func TestGridCJK(t *testing.T) {
	t.Parallel()

	got := grid(t, Bordered, []string{"名前", "状態"}, []string{"日本db", "稼働"})
	want := "" +
		"┌────────┬──────┐\n" +
		"│  名前  │ 状態 │\n" +
		"├────────┼──────┤\n" +
		"│ 日本db │ 稼働 │\n" +
		"└────────┴──────┘\n"

	if got != want {
		t.Fatalf("Grid(cjk) = %q, want %q", got, want)
	}
}

func TestGridANSIHeaderAligns(t *testing.T) {
	t.Parallel()

	got := grid(t, Bordered, []string{"\x1b[4mNAME\x1b[0m", "STATUS"}, []string{"web", "ok"})
	want := "" +
		"┌──────┬────────┐\n" +
		"│ \x1b[4mNAME\x1b[0m │ STATUS │\n" +
		"├──────┼────────┤\n" +
		"│ web  │ ok     │\n" +
		"└──────┴────────┘\n"

	if got != want {
		t.Fatalf("Grid(ansi header) = %q, want %q", got, want)
	}
}

func TestGridMultiLineCell(t *testing.T) {
	t.Parallel()

	got := grid(t, Bordered, []string{"A", "B"}, []string{"x\ny", "z"})
	want := "" +
		"┌───┬───┐\n" +
		"│ A │ B │\n" +
		"├───┼───┤\n" +
		"│ x │ z │\n" +
		"│ y │   │\n" +
		"└───┴───┘\n"

	if got != want {
		t.Fatalf("Grid(multiline) = %q, want %q", got, want)
	}
}

func TestGridMarkdownMultiLine(t *testing.T) {
	t.Parallel()

	got := grid(t, Markdown, []string{"A", "B"}, []string{"x\ny", "z"})
	want := "" +
		"| A | B |\n" +
		"|---|---|\n" +
		"| x | z |\n" +
		"| y |   |\n"

	if got != want {
		t.Fatalf("Grid(markdown multiline) = %q, want %q", got, want)
	}
}

func TestGridWrapsWhenTooWide(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	if err := gridTo(&b, 12, 8, Bordered, []string{"N"}, [][]string{{"alpha beta gamma"}}); err != nil {
		t.Fatalf("gridTo() error = %v", err)
	}
	want := "" +
		"┌───────┐\n" +
		"│   N   │\n" +
		"├───────┤\n" +
		"│ alpha │\n" +
		"│ beta  │\n" +
		"│ gamma │\n" +
		"└───────┘\n"

	if got := b.String(); got != want {
		t.Fatalf("gridTo(wrapped) = %q, want %q", got, want)
	}
}

func TestGridHyphenatesLongWords(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	if err := gridTo(&b, 20, 10, Bordered, []string{"T"}, [][]string{{"supercalifragilistic"}}); err != nil {
		t.Fatalf("gridTo() error = %v", err)
	}
	want := "" +
		"┌────────────┐\n" +
		"│     T      │\n" +
		"├────────────┤\n" +
		"│ supercali- │\n" +
		"│ fragilist- │\n" +
		"│ ic         │\n" +
		"└────────────┘\n"

	if got := b.String(); got != want {
		t.Fatalf("gridTo(hyphen) = %q, want %q", got, want)
	}
}

func TestGridNoColumnsIsSilent(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	if err := Grid(&b, Bordered, nil, [][]string{{"x"}}); err != nil {
		t.Fatalf("Grid() error = %v", err)
	}
	if b.String() != "" {
		t.Fatalf("Grid(no header) = %q, want silence", b.String())
	}
}

func TestPlainGoldens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header []string
		rows   [][]string
		want   string
	}{
		{
			name:   "basic",
			header: []string{"NAME", "STATUS"},
			rows:   [][]string{{"web", "ok"}},
			want:   "NAME  STATUS  \nweb   ok      \n",
		},
		{
			name:   "ragged",
			header: []string{"NAME", "STATUS"},
			rows:   [][]string{{"web"}, {"proxy", "down", "extra"}},
			want: "NAME   STATUS  \n" +
				"web            \n" +
				"proxy  down    \n",
		},
		{
			name:   "multiline",
			header: []string{"A", "B"},
			rows:   [][]string{{"x\ny", "z"}},
			want: "A  B  \n" +
				"x  z  \n" +
				"y     \n",
		},
		{
			name:   "cjk",
			header: []string{"名前"},
			rows:   [][]string{{"日本db"}},
			want:   "名前    \n日本db  \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var b strings.Builder
			if err := Plain(&b, tt.header, tt.rows); err != nil {
				t.Fatalf("Plain() error = %v", err)
			}
			if got := b.String(); got != tt.want {
				t.Fatalf("Plain(%s) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestPlainNoColumnsIsSilent(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	if err := Plain(&b, nil, [][]string{{"x"}}); err != nil {
		t.Fatalf("Plain() error = %v", err)
	}
	if b.String() != "" {
		t.Fatalf("Plain(no header) = %q, want silence", b.String())
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("boom")
}

func TestWriteErrorsPropagate(t *testing.T) {
	t.Parallel()

	if err := Grid(errorWriter{}, Bordered, []string{"A"}, [][]string{{"1"}}); err == nil {
		t.Fatal("Grid(errorWriter) = nil, want error")
	}
	if err := Plain(errorWriter{}, []string{"A"}, [][]string{{"1"}}); err == nil {
		t.Fatal("Plain(errorWriter) = nil, want error")
	}
}
