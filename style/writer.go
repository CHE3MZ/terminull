package style

import (
	"io"

	"github.com/CHE3MZ/terminull/core"
)

// Stdout returns a writer for standard output that translates ANSI codes
// on legacy Windows consoles and passes through everywhere else.
func Stdout() io.Writer {
	return core.Stdout()
}

// Stderr returns a writer for standard error with the same guarantee as Stdout.
func Stderr() io.Writer {
	return core.Stderr()
}
