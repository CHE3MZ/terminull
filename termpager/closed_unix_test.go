//go:build !windows

package termpager

import (
	"fmt"
	"syscall"
	"testing"
)

func TestClosedPipe(t *testing.T) {
	t.Parallel()

	if !Closed(syscall.EPIPE) {
		t.Fatal("Closed(EPIPE) = false")
	}
	if !Closed(fmt.Errorf("write: %w", syscall.EPIPE)) {
		t.Fatal("Closed(wrapped EPIPE) = false")
	}
}
