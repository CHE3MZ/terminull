//go:build !windows

package termimage

import (
	"os"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// probeSixel asks the terminal directly with DA1 and watches the reply
// for sixel support. A select timeout keeps the read from ever stealing
// a later keystroke.
func probeSixel() bool {
	f := os.Stdin
	fd := int(f.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return false
	}
	defer func() { _ = term.Restore(fd, old) }()

	if _, err := unix.Write(fd, []byte("\x1b[c")); err != nil {
		return false
	}
	var set unix.FdSet
	set.Set(fd)
	tv := unix.NsecToTimeval(probeTimeout.Nanoseconds())
	n, err := unix.Select(fd+1, &set, nil, nil, &tv)
	if err != nil || n == 0 {
		return false
	}
	buf := make([]byte, 64)
	m, err := unix.Read(fd, buf)
	if err != nil || m <= 0 {
		return false
	}
	return hasSixelDA(buf[:m])
}

// queryCellPixels asks TIOCGWINSZ for the window's pixel size; terminals
// that do not track pixels report zeros and the caller falls back.
func queryCellPixels() (int, int) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 || ws.Row == 0 || ws.Xpixel == 0 || ws.Ypixel == 0 {
		return 0, 0
	}
	return int(ws.Xpixel) / int(ws.Col), int(ws.Ypixel) / int(ws.Row)
}

// isWindowsConsole is always false off Windows.
func isWindowsConsole() bool { return false }
