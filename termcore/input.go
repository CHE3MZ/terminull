package termcore

import "golang.org/x/term"

// ReadPassword reads a line from the terminal without echo. Callers keep
// the prompt and cancellation policy; this only owns the platform read.
func ReadPassword(fd int) ([]byte, error) {
	return term.ReadPassword(fd)
}
