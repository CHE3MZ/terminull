package termcore

import (
	"os"

	"golang.org/x/sys/windows"
)

// terminalColumns queries the console width, reporting false when f is
// not a console.
func terminalColumns(f *os.File) (int, bool) {
	if f == nil {
		return 0, false
	}
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(f.Fd()), &info); err != nil {
		return 0, false
	}
	cols := int(info.Window.Right-info.Window.Left) + 1
	if cols <= 0 {
		return 0, false
	}
	return cols, true
}

// terminalRows queries the console height, reporting false when f is not
// a console.
func terminalRows(f *os.File) (int, bool) {
	if f == nil {
		return 0, false
	}
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(f.Fd()), &info); err != nil {
		return 0, false
	}
	rows := int(info.Window.Bottom-info.Window.Top) + 1
	if rows <= 0 {
		return 0, false
	}
	return rows, true
}
