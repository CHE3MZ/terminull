package console

import (
	"os"

	"golang.org/x/sys/windows"
)

// PrepareTerminalEnvironment enables ANSI processing on Windows consoles
// (a no-op where already on, an error where unsupported).
func PrepareTerminalEnvironment() error {
	// enable handling ASCII codes
	err := addConsoleMode(windows.Stdout, windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	if err != nil {
		return err
	}
	return addConsoleMode(windows.Stderr, windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}

func addConsoleMode(handle windows.Handle, flags uint32) error {
	var mode uint32

	err := windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return err
	}
	return windows.SetConsoleMode(handle, mode|flags)
}

// TerminalHeight is unimplemented on Windows: figure out how to query
// console height there, until then callers draw unbounded frames.
func TerminalHeight(file *os.File) int {
	// todo: figure out how to find out console height on Windows
	return -1
}
