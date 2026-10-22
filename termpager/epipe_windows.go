package termpager

import (
	"errors"
	"syscall"
)

// Windows pipe errors not exported by the syscall package.
const (
	errorNoData     syscall.Errno = 232 // ERROR_NO_DATA
	errorBrokenPipe syscall.Errno = 109 // ERROR_BROKEN_PIPE
)

// Closed reports a broken downstream pipe: the pager (or piped reader)
// went away, typically because the user quit early. Treat as success,
// not failure.
func Closed(err error) bool {
	if err == nil {
		return false
	}
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return false
	}
	return errno == errorNoData || errno == errorBrokenPipe
}
