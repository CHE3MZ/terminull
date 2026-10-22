package terminal_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/CHE3MZ/terminull/v2/termrender/terminal"
)

func TestCalculateIncrementalUpdateSameTwoLines(t *testing.T) {
	t.Parallel()

	var result bytes.Buffer
	terminal.CalculateIncrementalUpdate(
		bufio.NewWriter(&result),
		[]string{"Foo", "Bar"},
		[]string{"Foo", "Bar"},
	)
	if result.String() != "" {
		t.Fatalf("got %q, want empty (no changes, no output)", result.String())
	}
}

func TestCalculateIncrementalUpdateAddSingleLine(t *testing.T) {
	t.Parallel()

	var result bytes.Buffer
	terminal.CalculateIncrementalUpdate(
		bufio.NewWriter(&result),
		[]string{"Foo", "Bar"},
		[]string{"Foo", "Bar", "Baz"},
	)
	if result.String() != "\rBaz\n" {
		t.Fatalf("got %q, want %q", result.String(), "\rBaz\n")
	}
}

func TestCalculateIncrementalUpdateInplaceChange(t *testing.T) {
	t.Parallel()

	var result bytes.Buffer
	terminal.CalculateIncrementalUpdate(
		bufio.NewWriter(&result),
		[]string{"Foo", "Bar", "Baz"},
		[]string{"Foo", "Updated Bar", "Baz"},
	)
	if result.String() != "\r\u001B[2A\u001B[KUpdated Bar\r\u001B[2B" {
		t.Fatalf("got %q, want in-place replace sequence", result.String())
	}
}
