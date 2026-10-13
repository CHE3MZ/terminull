package utils

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		duration time.Duration
		decimals bool
		want     string
	}{
		{1500 * time.Millisecond, true, "1.5s"},
		{42 * time.Second, true, "42s"},
		{42 * time.Second, false, "42s"},
		{65 * time.Second, true, "01:05"},
		{3661 * time.Second, true, "01:01:01"},
	}

	for _, tt := range tests {
		if got := FormatDuration(tt.duration, tt.decimals); got != tt.want {
			t.Errorf("FormatDuration(%v, %v) = %q, want %q", tt.duration, tt.decimals, got, tt.want)
		}
	}
}
