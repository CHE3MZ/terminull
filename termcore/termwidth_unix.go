//go:build !windows

package termcore

import (
	"os"

	"golang.org/x/sys/unix"
)

// terminalColumns queries the console width via ioctl, reporting false
// when f is not a terminal.
func terminalColumns(f *os.File) (int, bool) {
	if f == nil {
		return 0, false
	}
	ws, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 {
		return 0, false
	}
	return int(ws.Col), true
}
