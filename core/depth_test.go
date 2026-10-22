package core

import "testing"

func TestColorDepthFrom(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		env  map[string]string
		win  bool
		want int
	}{
		{"truecolor", map[string]string{"COLORTERM": "truecolor"}, false, DepthTrue},
		{"24bit upcase", map[string]string{"COLORTERM": "24BIT"}, false, DepthTrue},
		{"windows console", nil, true, DepthTrue},
		{"console beats 256 term", map[string]string{"TERM": "xterm-256color"}, true, DepthTrue},
		{"256color term", map[string]string{"TERM": "xterm-256color"}, false, Depth256},
		{"base term", map[string]string{"TERM": "xterm"}, false, Depth16},
		{"empty", nil, false, Depth16},
	} {
		getenv := func(k string) string { return tc.env[k] }
		if got := colorDepthFrom(getenv, tc.win); got != tc.want {
			t.Fatalf("%s: colorDepthFrom = %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestColorDepthTier(t *testing.T) {
	t.Parallel()

	switch ColorDepth() {
	case Depth16, Depth256, DepthTrue:
	default:
		t.Fatalf("ColorDepth() = %d, want a known tier", ColorDepth())
	}
}
