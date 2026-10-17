//go:build !windows

package style

// EnsureConsole prepares the console for ANSI output. No-op on Unix.
func EnsureConsole() error {
	return nil
}

// isWindowsConsole is always false off Windows.
func isWindowsConsole() bool { return false }
