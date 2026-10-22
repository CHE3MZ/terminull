//go:build !windows

package termcore

import (
	"bytes"
	"os"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// TerminalQuery writes query to f, switches f to raw mode, and returns
// the reply up to terminator, timeout, or error. Detection callers use it
// for capability probes so no other package touches terminal modes.
// Returns nil when f is not a terminal or the probe fails.
func TerminalQuery(f *os.File, query []byte, timeout time.Duration, terminator byte) []byte {
	if f == nil {
		return nil
	}

	fd := int(f.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return nil
	}
	defer func() { _ = term.Restore(fd, old) }()

	if _, err := unix.Write(fd, query); err != nil {
		return nil
	}

	deadline := time.Now().Add(timeout)
	var resp []byte

	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}

		var set unix.FdSet
		set.Set(fd)

		tv := unix.NsecToTimeval(remaining.Nanoseconds())
		n, err := unix.Select(fd+1, &set, nil, nil, &tv)
		if err != nil || n == 0 {
			break
		}

		buf := make([]byte, 64)
		m, err := unix.Read(fd, buf)
		if err != nil || m <= 0 {
			break
		}

		resp = append(resp, buf[:m]...)
		if bytes.IndexByte(resp, terminator) >= 0 {
			break
		}
	}

	return resp
}

// CellPixels returns the terminal cell size in device pixels for f, or
// zeros when the terminal does not report it.
func CellPixels(f *os.File) (int, int) {
	if f == nil {
		return 0, 0
	}

	ws, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 || ws.Row == 0 || ws.Xpixel == 0 || ws.Ypixel == 0 {
		return 0, 0
	}

	return int(ws.Xpixel) / int(ws.Col), int(ws.Ypixel) / int(ws.Row)
}
