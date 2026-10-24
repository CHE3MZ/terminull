//go:build !windows

package termcore

import "os"

// EnsureConsole prepares the console for ANSI output. No-op on Unix.
func EnsureConsole() error {
	return nil
}

// SupportsANSI reports whether f can accept ANSI control sequences.
// Off Windows that means f is an interactive terminal and TERM is not
// a limited console such as dumb or linux.
func SupportsANSI(f *os.File) bool {
	return !plainTerminal() && IsTerminal(f)
}

// isWindowsConsole is always false off Windows.
func isWindowsConsole() bool { return false }
