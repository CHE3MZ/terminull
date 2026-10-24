package termcore

import (
	"os"

	"golang.org/x/sys/windows"
)

// EnsureConsole enables ANSI processing on Windows consoles so colors,
// spinners, and tree drawing render instead of printing raw escapes.
// Best effort: pipes and pre-VT systems return an error and callers
// proceed with plain output.
func EnsureConsole() error {
	if err := addConsoleMode(windows.Stdout, windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		return err
	}
	return addConsoleMode(windows.Stderr, windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}

func addConsoleMode(handle windows.Handle, flags uint32) error {
	var mode uint32

	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return err
	}
	return windows.SetConsoleMode(handle, mode|flags)
}

// SupportsANSI reports whether f can accept ANSI control sequences.
// Windows Terminal already enables virtual-terminal processing; legacy
// conhost is upgraded in place when the console API allows it. A pipe,
// file, or pre-VT console reports false so callers can use a plain
// fallback instead of printing raw escapes.
func SupportsANSI(f *os.File) bool {
	if f == nil || plainTerminal() {
		return false
	}

	handle := windows.Handle(f.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return false
	}
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0 {
		return true
	}
	return windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING) == nil
}

// isWindowsConsole reports whether stdout is a real console, which
// paints 24-bit color once VT is enabled.
func isWindowsConsole() bool {
	var mode uint32
	return windows.GetConsoleMode(windows.Stdout, &mode) == nil
}
