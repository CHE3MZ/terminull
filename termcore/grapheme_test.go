package termcore

import "testing"

const (
	familyEmoji = "\U0001F468\u200D\U0001F469\u200D\U0001F467\u200D\U0001F466"
	flagEmoji   = "\U0001F1FA\U0001F1F8"
	thumbEmoji  = "\U0001F44D\U0001F3FD"
	heartEmoji  = "\u2764\uFE0F"
)

func TestFirstGrapheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		in        string
		want      string
		wantWidth int
	}{
		{name: "empty", in: "", want: "", wantWidth: 0},
		{name: "ascii", in: "abc", want: "a", wantWidth: 1},
		{name: "combining mark", in: "e\u0301x", want: "e\u0301", wantWidth: 1},
		{name: "wide rune", in: "\u754Cx", want: "\u754C", wantWidth: 2},
		{name: "zwj family", in: familyEmoji + "x", want: familyEmoji, wantWidth: 2},
		{name: "flag", in: flagEmoji + "x", want: flagEmoji, wantWidth: 2},
		{name: "skin tone", in: thumbEmoji + "x", want: thumbEmoji, wantWidth: 2},
		{name: "variation selector", in: heartEmoji + "x", want: heartEmoji, wantWidth: 1},
		{name: "hangul jamo", in: "\u1100\u1161x", want: "\u1100\u1161", wantWidth: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, width := FirstGrapheme(tt.in)
			if got != tt.want || width != tt.wantWidth {
				t.Fatalf("FirstGrapheme(%q) = %q, %d; want %q, %d", tt.in, got, width, tt.want, tt.wantWidth)
			}
		})
	}
}
