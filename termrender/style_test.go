package termrender

import "testing"

func TestParseLogStyle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want LogStyle
	}{
		{"tree", LogStyleTree},
		{" plain ", LogStylePlain},
		{"TIME", LogStyleTime},
		{"time-level", LogStyleTimeLevel},
		{"clock", LogStyleTimeLevel},
		{"logrus", LogStyleLogrus},
		{"k8s", LogStyleKubernetes},
		{"kubernetes", LogStyleKubernetes},
	}

	for _, tt := range tests {
		got, err := ParseLogStyle(tt.in)
		if err != nil {
			t.Fatalf("ParseLogStyle(%q) error = %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("ParseLogStyle(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}

	if _, err := ParseLogStyle("rainbow"); err == nil {
		t.Fatal("ParseLogStyle(rainbow) = nil error, want error")
	}
}

func TestLogStyleStringRoundTrip(t *testing.T) {
	t.Parallel()

	styles := []LogStyle{
		LogStyleTree,
		LogStylePlain,
		LogStyleTime,
		LogStyleTimeLevel,
		LogStyleLogrus,
		LogStyleKubernetes,
	}
	for _, style := range styles {
		got, err := ParseLogStyle(style.String())
		if err != nil {
			t.Fatalf("ParseLogStyle(%q) error = %v", style, err)
		}
		if got != style {
			t.Fatalf("round trip %v -> %q -> %v", style, style.String(), got)
		}
	}
}
