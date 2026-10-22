package termbox

import (
	"strings"
	"testing"

	"github.com/CHE3MZ/terminull/termcore"
	"github.com/CHE3MZ/terminull/termstyle"
)

func TestStyleString(t *testing.T) {
	cases := map[Style]string{
		StyleRounded: "rounded",
		StyleHard:    "hard",
		StyleDouble:  "double",
		StyleASCII:   "ascii",
		StyleNone:    "none",
		Style(99):    "unknown",
	}
	for s, want := range cases {
		if got := s.String(); got != want {
			t.Errorf("Style(%d).String() = %q, want %q", s, got, want)
		}
	}
}

func TestStylesGolden(t *testing.T) {
	cases := []struct {
		name    string
		content string
		opts    Options
		want    string
	}{
		{
			name:    "zero value",
			content: "hello",
			want:    "╭─────╮\n│hello│\n╰─────╯\n",
		},
		{
			name:    "rounded padded",
			content: "go",
			opts:    Options{Padding: 1},
			want:    "╭────╮\n│ go │\n╰────╯\n",
		},
		{
			name:    "hard padded",
			content: "go",
			opts:    Options{Style: StyleHard, Padding: 1},
			want:    "┌────┐\n│ go │\n└────┘\n",
		},
		{
			name:    "double padded",
			content: "go",
			opts:    Options{Style: StyleDouble, Padding: 1},
			want:    "╔════╗\n║ go ║\n╚════╝\n",
		},
		{
			name:    "ascii padded",
			content: "go",
			opts:    Options{Style: StyleASCII, Padding: 1},
			want:    "+----+\n| go |\n+----+\n",
		},
		{
			name:    "none",
			content: "a\nb",
			opts:    Options{Style: StyleNone},
			want:    "a\nb\n",
		},
		{
			name:    "empty content",
			content: "",
			want:    "╭─╮\n│ │\n╰─╯\n",
		},
		{
			name:    "negative padding clamps",
			content: "x",
			opts:    Options{Padding: -3},
			want:    "╭─╮\n│x│\n╰─╯\n",
		},
		{
			name:    "unknown style is rounded",
			content: "x",
			opts:    Options{Style: Style(99)},
			want:    "╭─╮\n│x│\n╰─╯\n",
		},
	}
	for _, c := range cases {
		if got := String(c.content, c.opts, termstyle.Scheme{}); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
}

func TestTitle(t *testing.T) {
	got := String("hi", Options{Title: "Note", Padding: 1}, termstyle.Scheme{})
	want := "╭─ Note ╮\n│ hi    │\n╰───────╯\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTitleTruncation(t *testing.T) {
	got := String("x", Options{Title: "a very long title indeed", Width: 20}, termstyle.Scheme{})
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3: %q", len(lines), got)
	}
	for i, line := range lines {
		if w := termcore.VisibleWidth(line); w != 20 {
			t.Errorf("line %d width = %d, want 20: %q", i, w, line)
		}
	}
	if !strings.HasPrefix(lines[0], "╭─ a very long") || !strings.Contains(lines[0], "...") {
		t.Errorf("title not truncated with ellipsis: %q", lines[0])
	}
}

func TestTitleNarrowBoxDropsIt(t *testing.T) {
	got := String("x", Options{Title: "won't fit", Width: 5}, termstyle.Scheme{})
	if want := "╭───╮\n│x  │\n╰───╯\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapFixedWidth(t *testing.T) {
	got := String("hello world foo", Options{Padding: 1, Width: 12}, termstyle.Scheme{})
	want := "╭──────────╮\n│ hello    │\n│ world    │\n│ foo      │\n╰──────────╯\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapHardBreaksLongWords(t *testing.T) {
	got := String("abcdefghijkl", Options{Style: StyleHard, Width: 6}, termstyle.Scheme{})
	want := "┌────┐\n│abcd│\n│efgh│\n│ijkl│\n└────┘\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWideRunesFitAndWrap(t *testing.T) {
	got := String("你好世界", Options{Padding: 1}, termstyle.Scheme{})
	want := "╭──────────╮\n│ 你好世界 │\n╰──────────╯\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got = String("ab你好cd", Options{Style: StyleHard, Width: 6}, termstyle.Scheme{})
	want = "┌────┐\n│ab你│\n│好cd│\n└────┘\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestANSIContentMeasuredByGlyphs(t *testing.T) {
	inner := termstyle.NewScheme(true).Paint(termstyle.FgRed, "hi")
	got := String(inner, Options{Padding: 1}, termstyle.Scheme{})
	want := "╭────╮\n│ " + inner + " │\n╰────╯\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	for i, line := range lines {
		if w := termcore.VisibleWidth(line); w != 6 {
			t.Errorf("line %d visible width = %d, want 6: %q", i, w, line)
		}
	}
}

func TestOSCHyperlinkContent(t *testing.T) {
	link := "\x1b]8;;https://example.com\x1b\\link\x1b]8;;\x1b\\"
	got := String(link, Options{}, termstyle.Scheme{})
	want := "╭────╮\n│" + link + "│\n╰────╯\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFrameColor(t *testing.T) {
	scheme := termstyle.NewScheme(true)
	got := String("hi", Options{Padding: 1, Frame: termstyle.FgCyan}, scheme)
	if !strings.Contains(got, "\x1b[36m╭────╮\x1b[0m") {
		t.Errorf("top border not colored: %q", got)
	}
	if !strings.Contains(got, "\x1b[36m│\x1b[0m") {
		t.Errorf("sides not colored: %q", got)
	}
	if !strings.Contains(got, "\x1b[36m╰────╯\x1b[0m") {
		t.Errorf("bottom border not colored: %q", got)
	}

	plain := String("hi", Options{Padding: 1, Frame: termstyle.FgCyan}, termstyle.Scheme{})
	if strings.ContainsRune(plain, 0x1b) {
		t.Errorf("disabled scheme emitted escape bytes: %q", plain)
	}
}

func TestTabsAndTrailingNewlines(t *testing.T) {
	got := String("a\tb\n\n", Options{Padding: 1}, termstyle.Scheme{})
	want := "╭────────────╮\n│ a        b │\n╰────────────╯\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestOpenStyleResetsAtLineEnds(t *testing.T) {
	got := String("\x1b[31mhello world", Options{Width: 7}, termstyle.Scheme{})
	want := "╭─────╮\n│\x1b[31mhello\x1b[0m│\n│\x1b[31mworld\x1b[0m│\n╰─────╯\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestClosedStyleIsNotRewritten(t *testing.T) {
	inner := termstyle.NewScheme(true).Paint(termstyle.FgRed, "hi")
	got := String(inner, Options{}, termstyle.Scheme{})
	want := "╭──╮\n│" + inner + "│\n╰──╯\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderMatchesString(t *testing.T) {
	var b strings.Builder
	if err := Render(&b, "x", Options{}, termstyle.Scheme{}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if want := String("x", Options{}, termstyle.Scheme{}); b.String() != want {
		t.Errorf("Render wrote %q, want %q", b.String(), want)
	}
}
