package table

import (
	"reflect"
	"testing"
)

const emojiFamily = "\U0001F468\u200D\U0001F469\u200D\U0001F467\u200D\U0001F466"

func TestWrapCell(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		in    string
		width int
		want  []string
	}{
		{name: "fits", in: "hello world", width: 20, want: []string{"hello world"}},
		{name: "collapses spaces", in: "hello   world", width: 20, want: []string{"hello world"}},
		{name: "newlines break", in: "a\nb", width: 20, want: []string{"a", "b"}},
		{name: "empty cell", in: "", width: 5, want: []string{""}},
		{name: "cjk exact", in: "日本語", width: 6, want: []string{"日本語"}},
		{name: "word wrap", in: "hello world", width: 5, want: []string{"hello", "world"}},
		{
			name:  "hyphenates long words",
			in:    "supercalifragilistic",
			width: 10,
			want:  []string{"supercali-", "fragilist-", "ic"},
		},
		{name: "keeps ansi intact", in: "\x1b[4mNAME\x1b[0m", width: 10, want: []string{"\x1b[4mNAME\x1b[0m"}},
		{
			name:  "ansi reopens when wrapped",
			in:    "\x1b[4mhello world\x1b[0m",
			width: 5,
			want:  []string{"\x1b[4mhello", "\x1b[4mworld\x1b[0m"},
		},
		{name: "never splits wide runes", in: "日本語", width: 3, want: []string{"日-", "本-", "語"}},
		{
			name:  "keeps emoji cluster whole",
			in:    emojiFamily + "xy",
			width: 2,
			want:  []string{emojiFamily, "xy"},
		},
		{
			name:  "narrow words make progress",
			in:    "abc",
			width: 1,
			want:  []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := wrapCell(tt.in, tt.width); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("wrapCell(%q, %d) = %q, want %q", tt.in, tt.width, got, tt.want)
			}
		})
	}
}

func TestCutCellsEmoji(t *testing.T) {
	t.Parallel()

	head, tail := cutCells(emojiFamily+"x", 1)
	if head != emojiFamily || tail != "x" {
		t.Fatalf("cutCells(emoji) = %q, %q", head, tail)
	}

	head, tail = cutCells("\x1b[4m"+emojiFamily+"x\x1b[0m", 2)
	if head != "\x1b[4m"+emojiFamily+"\x1b[0m" || tail != "\x1b[4mx\x1b[0m" {
		t.Fatalf("cutCells(styled emoji) = %q, %q", head, tail)
	}
}

func TestCutCells(t *testing.T) {
	t.Parallel()

	head, tail := cutCells("\x1b[4mabcdef\x1b[0m", 3)
	if head != "\x1b[4mabc\x1b[0m" {
		t.Fatalf("cutCells head = %q", head)
	}
	if tail != "\x1b[4mdef\x1b[0m" {
		t.Fatalf("cutCells tail = %q", tail)
	}

	head, tail = cutCells("日本語", 3)
	if head != "日" || tail != "本語" {
		t.Fatalf("cutCells(cjk) = %q, %q", head, tail)
	}
}
