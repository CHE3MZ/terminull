package config

import (
	"runtime"
	"time"

	"github.com/CHE3MZ/terminull/render/terminal"
)

const defaultVisibleLines = 5

// InteractiveRendererConfig tunes the live tree: spinner frames, status
// glyphs, refresh rate, and how much failed output stays visible.
type InteractiveRendererConfig struct {
	Colors                         *terminal.ColorSchema
	RefreshRate                    time.Duration
	ProgressIndicatorFrames        []string
	ProgressIndicatorCycleDuration time.Duration
	SuccessStatus                  string
	FailureStatus                  string
	SkippedStatus                  string
	DescriptionLinesWhenFailed     int
	DescriptionLinesWhenSkipped    int
	VisibleDescriptionLines        int
}

// NewDefaultRenderingConfig selects emoji on Unix and symbols-only on Windows.
func NewDefaultRenderingConfig() *InteractiveRendererConfig {
	if runtime.GOOS == "windows" {
		return NewDefaultSymbolsOnlyRenderingConfig()
	}
	return NewDefaultEmojiRenderingConfig()
}

// NewDefaultEmojiRenderingConfig builds the emoji config for modern terminals.
func NewDefaultEmojiRenderingConfig() *InteractiveRendererConfig {
	return &InteractiveRendererConfig{
		Colors:      terminal.DefaultColorSchema(),
		RefreshRate: 200 * time.Microsecond,
		ProgressIndicatorFrames: []string{
			"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛",
		},
		ProgressIndicatorCycleDuration: time.Second,
		SuccessStatus:                  "✅",
		FailureStatus:                  "❌",
		SkippedStatus:                  "⏩",
		DescriptionLinesWhenFailed:     100,
		DescriptionLinesWhenSkipped:    0,
		VisibleDescriptionLines:        defaultVisibleLines,
	}
}

// NewDefaultSymbolsOnlyRenderingConfig builds the ASCII config for legacy
// Windows consoles and anywhere else emoji cannot render.
func NewDefaultSymbolsOnlyRenderingConfig() *InteractiveRendererConfig {
	return &InteractiveRendererConfig{
		Colors:      terminal.DefaultColorSchema(),
		RefreshRate: 250 * time.Microsecond,
		ProgressIndicatorFrames: []string{
			"\\", "|", "/", "-",
		},
		ProgressIndicatorCycleDuration: time.Second,
		SuccessStatus:                  "+",
		FailureStatus:                  "-",
		SkippedStatus:                  "!",
		DescriptionLinesWhenFailed:     100,
		DescriptionLinesWhenSkipped:    0,
		VisibleDescriptionLines:        defaultVisibleLines,
	}
}

// CurrentProgressIndicatorFrame returns the spinner frame for right now.
func (config *InteractiveRendererConfig) CurrentProgressIndicatorFrame() string {
	amountOfFrames := int64(len(config.ProgressIndicatorFrames))
	nanosPerFrame := int64(config.ProgressIndicatorCycleDuration) / amountOfFrames
	currentNanosTail := time.Now().UnixNano() % int64(config.ProgressIndicatorCycleDuration)
	frameIndex := currentNanosTail / nanosPerFrame
	if frameIndex < amountOfFrames {
		return config.ProgressIndicatorFrames[frameIndex]
	}
	return config.ProgressIndicatorFrames[0]
}
