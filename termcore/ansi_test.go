package termcore

import "testing"

func TestVisibleWidth(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"plain", "hello", 5},
		{"sgr", "\x1b[31mred\x1b[0m", 3},
		{"truecolor", "\x1b[38;2;255;0;0mx\x1b[0m", 1},
		{"osc bel", "\x1b]8;;https://example.com\x07link\x1b]8;;\x07", 4},
		{"osc st", "\x1b]8;;https://example.com\x1b\\link\x1b]8;;\x1b\\", 4},
		{"cjk", "你好", 4},
		{"mixed", "a\x1b[1mb你\x1b[22m", 4},
	}
	for _, c := range cases {
		if got := VisibleWidth(c.in); got != c.want {
			t.Errorf("%s: VisibleWidth(%q) = %d, want %d", c.name, c.in, got, c.want)
		}
	}
}

func TestRuneWidth(t *testing.T) {
	cases := []struct {
		r    rune
		want int
	}{
		{'a', 1},
		{'你', 2},
		{'\u0301', 0},
	}
	for _, c := range cases {
		if got := RuneWidth(c.r); got != c.want {
			t.Errorf("RuneWidth(%q) = %d, want %d", c.r, got, c.want)
		}
	}
}

func TestEscapeLen(t *testing.T) {
	cases := []struct {
		name string
		in   string
		i    int
		want int
	}{
		{"not an escape", "abc", 0, 0},
		{"out of range", "abc", 5, 0},
		{"sgr", "\x1b[0m", 0, 4},
		{"truecolor", "\x1b[38;2;1;2;3m", 0, 13},
		{"lone esc", "\x1b", 0, 1},
		{"unterminated csi", "\x1b[31", 0, 4},
		{"osc bel", "\x1b]8;;x\x07", 0, 7},
		{"osc st", "\x1b]8;;x\x1b\\", 0, 8},
		{"unterminated osc", "\x1b]8;;x", 0, 6},
		{"two byte", "\x1bM", 0, 2},
		{"offset", "a\x1b[0m", 1, 4},
	}
	for _, c := range cases {
		if got := EscapeLen(c.in, c.i); got != c.want {
			t.Errorf("%s: EscapeLen(%q, %d) = %d, want %d", c.name, c.in, c.i, got, c.want)
		}
	}
}
