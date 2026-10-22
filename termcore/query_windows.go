package termcore

import (
	"bytes"
	"os"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modkernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	procGetCurrentConsoleFontEx = modkernel32.NewProc("GetCurrentConsoleFontExW")
)

// TerminalQuery writes query to stdout, switches console input to VT
// translation, and returns the reply up to terminator, timeout, or error.
// Detection callers use it for capability probes so no other package
// touches console modes. Returns nil when f is not a console or the probe
// fails.
func TerminalQuery(f *os.File, query []byte, timeout time.Duration, terminator byte) []byte {
	if f == nil {
		return nil
	}

	in := windows.Handle(f.Fd())

	var mode uint32
	if err := windows.GetConsoleMode(in, &mode); err != nil {
		return nil
	}

	vtMode := mode | windows.ENABLE_VIRTUAL_TERMINAL_INPUT | windows.ENABLE_PROCESSED_INPUT
	if err := windows.SetConsoleMode(in, vtMode); err != nil {
		return nil
	}
	defer func() { _ = windows.SetConsoleMode(in, mode) }()

	if _, err := os.Stdout.Write(query); err != nil {
		return nil
	}

	deadline := time.Now().Add(timeout)
	var resp []byte

	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}

		ms := remaining.Milliseconds()
		if ms < 1 {
			ms = 1
		}
		if max := int64(timeout / time.Millisecond); ms > max {
			ms = max
		}

		ev, err := windows.WaitForSingleObject(in, uint32(ms))
		if err != nil || ev != 0 {
			break
		}

		buf := make([]byte, 64)
		n, _ := f.Read(buf)
		if n > 0 {
			resp = append(resp, buf[:n]...)
			if bytes.IndexByte(resp, terminator) >= 0 {
				break
			}
		}
	}

	return resp
}

// consoleFontInfoEx mirrors CONSOLE_FONT_INFOEX for the font query.
type consoleFontInfoEx struct {
	cbSize     uint32
	nFont      uint32
	dwFontSize windows.Coord
	fontFamily uint32
	fontWeight uint32
	faceName   [32]uint16
}

// CellPixels returns the console font cell size for f, or zeros when f
// is not attached to a console.
func CellPixels(f *os.File) (int, int) {
	if f == nil {
		return 0, 0
	}
	if err := procGetCurrentConsoleFontEx.Find(); err != nil {
		return 0, 0
	}

	out := windows.Handle(f.Fd())
	var info consoleFontInfoEx
	info.cbSize = uint32(unsafe.Sizeof(info))

	r, _, _ := procGetCurrentConsoleFontEx.Call(uintptr(out), 0, uintptr(unsafe.Pointer(&info))) // #nosec G103 -- Win32 fills a caller-owned stack struct that outlives the syscall.
	if r == 0 {
		return 0, 0
	}

	return int(info.dwFontSize.X), int(info.dwFontSize.Y)
}
