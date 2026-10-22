package core

import (
	"os"
	"runtime"
	"strings"

	"github.com/CHE3MZ/terminull/core/internal/isatty"
)

// IsTerminal reports whether f is an interactive terminal.
// Pipes, files, and CI log collectors are not terminals.
func IsTerminal(f *os.File) bool {
	if f == nil {
		return false
	}

	return isatty.IsTerminal(f.Fd())
}

// Color reports whether ANSI color codes may be written to f.
// Precedence: noColor flag, NO_COLOR env, TERM=dumb, not-a-tty, CI.
//
// Pass the --no-color flag value as noColor. Never read os.Stdout
// directly for this decision; always pass the actual destination file.
func Color(f *os.File, noColor bool) bool {
	if noColor {
		return false
	}

	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}

	if strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}

	if !IsTerminal(f) {
		return false
	}

	if os.Getenv("CI") == "true" {
		return false
	}

	return true
}

// Emoji reports whether emoji glyphs may be used.
// False on dumb terminals, in CI, and on legacy Windows consoles
// (conhost without Windows Terminal, detected via missing WT_SESSION),
// where raster fonts cannot render them and they break column layout.
func Emoji() bool {
	if strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}

	if os.Getenv("CI") == "true" {
		return false
	}

	if runtime.GOOS == "windows" {
		if _, ok := os.LookupEnv("WT_SESSION"); !ok {
			return false
		}
	}

	return true
}
