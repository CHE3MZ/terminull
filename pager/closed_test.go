package pager

import (
	"errors"
	"testing"
)

func TestClosedGeneric(t *testing.T) {
	t.Parallel()

	if Closed(nil) {
		t.Fatal("Closed(nil) = true")
	}
	if Closed(errors.New("boom")) {
		t.Fatal("Closed(plain error) = true")
	}
}
