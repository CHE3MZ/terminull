package liner

import "testing"

const (
	testFamily = "\U0001F468\u200D\U0001F469\u200D\U0001F467\u200D\U0001F466"
	testFlag   = "\U0001F1FA\U0001F1F8"
)

func TestCountGlyphsClusters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want int
	}{
		{in: "", want: 0},
		{in: "abc", want: 3},
		{in: "e\u0301x", want: 2},
		{in: testFamily, want: 2},
		{in: testFlag, want: 2},
	}
	for _, tt := range tests {
		if got := countGlyphs([]rune(tt.in)); got != tt.want {
			t.Errorf("countGlyphs(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestGetPrefixGlyphsClusters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		num  int
		want string
	}{
		{in: testFamily + "x", num: 1, want: testFamily},
		{in: testFamily + "x", num: 2, want: testFamily + "x"},
		{in: "e\u0301x", num: 1, want: "e\u0301"},
		{in: "x" + testFlag, num: 1, want: "x"},
		{in: "x" + testFlag, num: 2, want: "x" + testFlag},
	}
	for _, tt := range tests {
		if got := string(getPrefixGlyphs([]rune(tt.in), tt.num)); got != tt.want {
			t.Errorf("getPrefixGlyphs(%q, %d) = %q, want %q", tt.in, tt.num, got, tt.want)
		}
	}
}

func TestGetSuffixGlyphsClusters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		num  int
		want string
	}{
		{in: "x" + testFamily, num: 1, want: testFamily},
		{in: "x" + testFamily, num: 2, want: "x" + testFamily},
		{in: "e\u0301x", num: 1, want: "x"},
		{in: testFamily + "x", num: 1, want: "x"},
		{in: "x" + testFlag, num: 1, want: testFlag},
	}
	for _, tt := range tests {
		if got := string(getSuffixGlyphs([]rune(tt.in), tt.num)); got != tt.want {
			t.Errorf("getSuffixGlyphs(%q, %d) = %q, want %q", tt.in, tt.num, got, tt.want)
		}
	}
}
