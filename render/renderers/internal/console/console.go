//go:build !windows

package console

import (
	"os"

	"golang.org/x/sys/unix"

	"github.com/CHE3MZ/terminull/style"
)

// PrepareTerminalEnvironment is a no-op on Unix: ANSI just works.
func PrepareTerminalEnvironment() error {
	return style.EnsureConsole()
}

// TerminalHeight returns the terminal row count, or -1 when unknown
// (piped, dumb, or query failed).
func TerminalHeight(file *os.File) int {
	ws, err := unix.IoctlGetWinsize(int(file.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return -1
	}

	return int(ws.Row)
}
