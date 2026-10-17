//go:build windows

package termimage

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

// probeSixel asks Windows Terminal directly with DA1. Console input is
// switched to VT translation for the read and restored after; any
// failure or timeout means no sixel and the caller falls back.
func probeSixel() bool {
	in := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(in, &mode); err != nil {
		return false
	}
	vtMode := mode | windows.ENABLE_VIRTUAL_TERMINAL_INPUT | windows.ENABLE_PROCESSED_INPUT
	if err := windows.SetConsoleMode(in, vtMode); err != nil {
		return false
	}
	defer func() { _ = windows.SetConsoleMode(in, mode) }()

	if _, err := os.Stdout.WriteString("\x1b[c"); err != nil {
		return false
	}

	deadline := time.Now().Add(probeTimeout)
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
		if max := int64(probeTimeout / time.Millisecond); ms > max {
			ms = max
		}
		ev, err := windows.WaitForSingleObject(in, uint32(ms))
		if err != nil || ev != 0 {
			break
		}
		buf := make([]byte, 64)
		n, _ := os.Stdin.Read(buf)
		if n > 0 {
			resp = append(resp, buf[:n]...)
			if bytes.ContainsRune(resp, 'c') {
				break
			}
		}
	}
	return hasSixelDA(resp)
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

// queryCellPixels reads the console font cell size when attached to a
// console; shells without one get the default estimate.
func queryCellPixels() (int, int) {
	if err := procGetCurrentConsoleFontEx.Find(); err != nil {
		return 0, 0
	}
	out := windows.Handle(os.Stdout.Fd())
	var info consoleFontInfoEx
	info.cbSize = uint32(unsafe.Sizeof(info))
	r, _, _ := procGetCurrentConsoleFontEx.Call(uintptr(out), 0, uintptr(unsafe.Pointer(&info))) // #nosec G103 -- Win32 fills a caller-owned stack struct that outlives the syscall.
	if r == 0 {
		return 0, 0
	}
	return int(info.dwFontSize.X), int(info.dwFontSize.Y)
}

// sixelUnits returns the DEC VT240 reference cell of 10x20 pixels:
// Windows consoles (Windows Terminal and conhost) size sixel pixels
// with that fixed assumption instead of the real font cell, so images
// declare their raster in these units to keep aspect on any font.
func sixelUnits(_, _ int) (int, int) { return 10, 20 }

// isWindowsConsole reports whether stdout is a real console, which does
// 24-bit color once VT is enabled.
func isWindowsConsole() bool {
	var mode uint32
	return windows.GetConsoleMode(windows.Handle(os.Stdout.Fd()), &mode) == nil
}
