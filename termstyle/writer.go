package termstyle

import (
	"io"

	"github.com/CHE3MZ/terminull/v2/termcore"
)

// Stdout returns a writer for standard output that translates ANSI codes
// on legacy Windows consoles and passes through everywhere else.
func Stdout() io.Writer {
	return termcore.Stdout()
}

// Stderr returns a writer for standard error with the same guarantee as Stdout.
func Stderr() io.Writer {
	return termcore.Stderr()
}
