package box

import (
	"reflect"
	"testing"
)

func TestBreakAt(t *testing.T) {
	cases := []struct {
		in   string
		cols int
		want int
	}{
		{"hello world", 8, 5},
		{"abcdefghij", 8, 8},
		{"ab你", 3, 2},
		{"a你b", 3, 4},
		{"hello", 3, 3},
		{"ab你cd", 3, 2},
	}
	for _, c := range cases {
		if got := breakAt(c.in, c.cols); got != c.want {
			t.Errorf("breakAt(%q, %d) = %d, want %d", c.in, c.cols, got, c.want)
		}
	}
}

func TestWrapLine(t *testing.T) {
	cases := []struct {
		name string
		in   string
		cols int
		want []string
	}{
		{"spaces", "hello world", 5, []string{"hello", "world"}},
		{"long word", "abcdefghijkl", 4, []string{"abcd", "efgh", "ijkl"}},
		{"fits", "short", 10, []string{"short"}},
		{
			"combining mark stays with base",
			"e\u0301x", 1,
			[]string{"e\u0301", "x"},
		},
	}
	for _, c := range cases {
		if got := wrapLine(c.in, c.cols); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: wrapLine(%q, %d) = %q, want %q", c.name, c.in, c.cols, got, c.want)
		}
	}
}

func TestWrapKeepsEscapesIntact(t *testing.T) {
	styled := "\x1b[31mhello world\x1b[0m"
	got := wrapLine(styled, 5)
	want := []string{"\x1b[31mhello", "world\x1b[0m"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wrapLine = %q, want %q", got, want)
	}
}

func TestCellsAttachEscapesToRune(t *testing.T) {
	cs := cells("\x1b[31mx")
	if len(cs) != 1 {
		t.Fatalf("got %d cells, want 1", len(cs))
	}
	if cs[0].start != 0 || cs[0].end != 6 || cs[0].r != 'x' || cs[0].w != 1 {
		t.Errorf("unexpected cell: %+v", cs[0])
	}
}
