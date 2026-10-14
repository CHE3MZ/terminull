package pager

import (
	"syscall"
	"testing"
)

func TestClosedPipe(t *testing.T) {
	t.Parallel()

	if !Closed(syscall.Errno(232)) {
		t.Fatal("Closed(ERROR_NO_DATA) = false")
	}
	if !Closed(syscall.Errno(109)) {
		t.Fatal("Closed(ERROR_BROKEN_PIPE) = false")
	}
}
