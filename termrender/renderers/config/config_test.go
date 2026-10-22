package config

import (
	"os"
	"runtime"
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

func TestDefaultFollowsEmojiCapability(t *testing.T) {
	t.Setenv("TERM", "xterm")
	t.Setenv("CI", "")
	previous, hadWT := os.LookupEnv("WT_SESSION")
	if err := os.Unsetenv("WT_SESSION"); err != nil {
		t.Fatalf("Unsetenv() error = %v", err)
	}
	t.Cleanup(func() {
		if hadWT {
			_ = os.Setenv("WT_SESSION", previous)
		}
	})

	// Without Windows Terminal, only non-Windows gets emoji glyphs.
	wantEmoji := runtime.GOOS != "windows"
	if got := NewDefaultRenderingConfig().SuccessStatus == "✅"; got != wantEmoji {
		t.Fatalf("default emoji = %v, want %v on %s", got, wantEmoji, runtime.GOOS)
	}

	// Windows Terminal joins the emoji set everywhere.
	t.Setenv("WT_SESSION", "test-session")
	if got := NewDefaultRenderingConfig().SuccessStatus; got != "✅" {
		t.Fatalf("default status with WT_SESSION = %q, want emoji", got)
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
