//go:build !windows

package pager

import (
	"errors"
	"syscall"
)

// Closed reports a broken downstream pipe: the pager (or piped reader)
// went away, typically because the user quit early. Treat as success,
// not failure.
func Closed(err error) bool {
	return err != nil && errors.Is(err, syscall.EPIPE)
}
