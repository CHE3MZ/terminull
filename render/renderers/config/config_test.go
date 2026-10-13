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

	cfg := &InteractiveRendererConfig{
		ProgressIndicatorFrames:        []string{"a", "b"},
		ProgressIndicatorCycleDuration: time.Hour,
	}

	// With an hour-long cycle the frame stays "a" for the whole test run.
	if got := cfg.CurrentProgressIndicatorFrame(); got != "a" {
		t.Fatalf("frame = %q, want a", got)
	}
}
