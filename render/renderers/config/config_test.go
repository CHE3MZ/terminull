package config

import (
	"testing"
	"time"
)

func TestDefaultConfigs(t *testing.T) {
	t.Parallel()

	emoji := NewDefaultEmojiRenderingConfig()
	if len(emoji.ProgressIndicatorFrames) == 0 || emoji.SuccessStatus == "" {
		t.Fatal("emoji config has no frames or status")
	}

	symbols := NewDefaultSymbolsOnlyRenderingConfig()
	for _, frame := range symbols.ProgressIndicatorFrames {
		if len(frame) != 1 {
			t.Fatalf("symbols frame %q is not single-width ASCII", frame)
		}
	}

	def := NewDefaultRenderingConfig()
	if def == nil || def.Colors == nil || def.RefreshRate <= 0 {
		t.Fatal("default config is not usable")
	}
}

func TestCurrentProgressIndicatorFrameCycles(t *testing.T) {
	t.Parallel()

	// A single frame always wins regardless of wall-clock time.
	single := &InteractiveRendererConfig{
		ProgressIndicatorFrames:        []string{"only"},
		ProgressIndicatorCycleDuration: time.Second,
	}
	if got := single.CurrentProgressIndicatorFrame(); got != "only" {
		t.Fatalf("frame = %q, want only", got)
	}

	// Multiple frames rotate with the clock: assert membership, never
	// an exact frame, or the test flakes by time of hour.
	multi := &InteractiveRendererConfig{
		ProgressIndicatorFrames:        []string{"a", "b"},
		ProgressIndicatorCycleDuration: time.Hour,
	}
	if got := multi.CurrentProgressIndicatorFrame(); got != "a" && got != "b" {
		t.Fatalf("frame = %q, want a or b", got)
	}
}
