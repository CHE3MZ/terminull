package termimage

import (
	"os"
	"strings"
	"time"

	"github.com/CHE3MZ/terminull/style"
)

// Detect picks the richest backend the current terminal can host:
// environment signals first, a sixel DA1 probe on real consoles next,
// ANSI half blocks last.
func Detect() Backend {
	if b, ok := backendFromEnv(os.Getenv); ok {
		return b
	}
	if style.IsTerminal(os.Stdin) && probeSixel() {
		return Sixel
	}
	return HalfBlock
}

// backendFromEnv maps terminal identity variables to protocols. Sixel
// cannot be trusted from environment alone on Windows Terminal (support
// arrived in 1.22), so those consoles go through the probe instead.
func backendFromEnv(getenv func(string) string) (Backend, bool) {
	if getenv("KITTY_WINDOW_ID") != "" {
		return Kitty, true
	}
	term := strings.ToLower(getenv("TERM"))
	prog := getenv("TERM_PROGRAM")
	switch {
	case strings.Contains(term, "kitty"), strings.Contains(term, "ghostty"):
		return Kitty, true
	case prog == "WezTerm":
		return Kitty, true
	case prog == "iTerm.app":
		return ITerm2, true
	case strings.Contains(term, "mlterm"), strings.Contains(term, "foot"), strings.Contains(term, "contour"):
		return Sixel, true
	}
	return Auto, false
}

// probeTimeout bounds how long detection waits for a DA1 reply.
const probeTimeout = 200 * time.Millisecond

// hasSixelDA scans a DA1 reply (ESC [ ? 62 ; ... c) for attribute 4,
// which marks sixel support.
func hasSixelDA(resp []byte) bool {
	s := string(resp)
	i := strings.Index(s, "\x1b[?")
	if i < 0 {
		return false
	}
	j := strings.IndexByte(s[i+3:], 'c')
	if j < 0 {
		return false
	}
	for _, attr := range strings.Split(s[i+3:i+3+j], ";") {
		if attr == "4" {
			return true
		}
	}
	return false
}

// cellPixels estimates the console cell size in device pixels, falling
// back to 8x16 when the terminal will not say.
func cellPixels() (int, int) {
	w, h := queryCellPixels()
	if w < 2 || h < 2 || w > 64 || h > 128 {
		return 8, 16
	}
	return w, h
}
