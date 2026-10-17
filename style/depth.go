package style

import (
	"os"
	"strings"
)

// Color depth tiers reported by ColorDepth.
const (
	Depth16   = 16
	Depth256  = 256
	DepthTrue = 1 << 24
)

// ColorDepth reports what ANSI color tiers the terminal can paint:
// COLORTERM truecolor/24bit wins (case-insensitive), a real Windows
// console paints 24-bit once VT is enabled, then 256-color terms, else
// the 16-color base set. One detector feeds text dimming and image
// backends alike.
func ColorDepth() int {
	return colorDepthFrom(os.Getenv, isWindowsConsole())
}

// colorDepthFrom is ColorDepth with the environment and console probe
// injected so tests can pin every tier.
func colorDepthFrom(getenv func(string) string, winConsole bool) int {
	switch strings.ToLower(getenv("COLORTERM")) {
	case "truecolor", "24bit":
		return DepthTrue
	}
	if winConsole {
		return DepthTrue
	}
	if strings.Contains(getenv("TERM"), "256color") {
		return Depth256
	}
	return Depth16
}
