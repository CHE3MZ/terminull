package style

import (
	"io"

	"github.com/mattn/go-colorable"
)

// Stdout returns a writer for standard output that translates ANSI codes
// on legacy Windows consoles and passes through everywhere else.
// Always write through this, never os.Stdout directly.
func Stdout() io.Writer {
	return colorable.NewColorableStdout()
}

// Stderr returns a writer for standard error with the same guarantee as Stdout.
func Stderr() io.Writer {
	return colorable.NewColorableStderr()
}
