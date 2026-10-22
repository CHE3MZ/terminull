package termpager

import (
	"errors"
	"syscall"

	"golang.org/x/sys/windows"
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
	return errno == windows.ERROR_NO_DATA || errno == windows.ERROR_BROKEN_PIPE
}
