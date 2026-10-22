package termimage

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CHE3MZ/terminull/v2/termcore"
)

// Detect picks the richest backend the current terminal can host:
// environment signals first, a sixel DA1 probe on real consoles next,
// ANSI half blocks last.
func Detect() Backend {
	if b, ok := backendFromEnv(os.Getenv); ok {
		return b
	}
	if termcore.IsTerminal(os.Stdin) && probeSixel() {
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
		// WezTerm implements kitty graphics, but its placements bind to
		// cells at insert time, and on Windows the escapes cross ConPTY,
		// which runs its own screen model and mangles them (wezterm's own
		// imgcat compensates explicitly). The iterm2 protocol is linear
		// and passes ConPTY intact, so WezTerm always takes that path.
		return ITerm2, true
	case prog == "iTerm.app":
		return ITerm2, true
	case konsoleKitty(getenv("KONSOLE_VERSION")):
		return Kitty, true
	case strings.Contains(term, "mlterm"), strings.Contains(term, "foot"), strings.Contains(term, "contour"):
		return Sixel, true
	}
	return Auto, false
}

// konsoleKitty reports whether KONSOLE_VERSION names a Konsole release
// with working kitty graphics: the protocol landed in 22.04, and the
// variable packs the version as YYMMPP (22.04.0 -> 220400).
func konsoleKitty(v string) bool {
	n, err := strconv.Atoi(v)
	return err == nil && n >= 220400
}

// probeTimeout bounds how long detection waits for a DA1 reply.
const probeTimeout = 200 * time.Millisecond

// probeSixel asks the terminal directly with DA1; termcore owns the raw
// mode, console mode, and timed read so image code stays protocol-only.
func probeSixel() bool {
	return hasSixelDA(termcore.TerminalQuery(os.Stdin, []byte("\x1b[c"), probeTimeout, 'c'))
}

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
	w, h := termcore.CellPixels(os.Stdout)
	if w < 2 || h < 2 || w > 64 || h > 128 {
		return 8, 16
	}
	return w, h
}
