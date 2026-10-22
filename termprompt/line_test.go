package termprompt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	liner "github.com/CHE3MZ/terminull/v2/termprompt/internal/lineedit"
)

func TestLineFallbackReadsRawLines(t *testing.T) {
	t.Parallel()

	stdio, drain := pipeIO(t, "hello world \nsecond\n")
	ed := NewEditor(stdio)

	got, err := ed.Line(context.Background(), LineOptions{Prefix: "you> "})
	if err != nil || got != "hello world " {
		t.Fatalf("Line() = %q, %v; want raw first line with trailing space", got, err)
	}
	got, err = ed.Line(context.Background(), LineOptions{Prefix: "you> "})
	if err != nil || got != "second" {
		t.Fatalf("Line() = %q, %v; want second line", got, err)
	}
	if _, err := ed.Line(context.Background(), LineOptions{Prefix: "you> "}); !errors.Is(err, io.EOF) {
		t.Fatalf("Line() error = %v, want io.EOF", err)
	}
	if out := drain(); !strings.Contains(out, "you> ") {
		t.Fatalf("output %q misses prefix", out)
	}

	var hist strings.Builder
	if _, err := ed.WriteHistory(&hist); err != nil {
		t.Fatalf("WriteHistory() error = %v", err)
	}
	if hist.String() != "hello world \nsecond\n" {
		t.Fatalf("history = %q, want both read lines", hist.String())
	}
}

func TestLineFallbackLastLineWithoutNewline(t *testing.T) {
	t.Parallel()

	stdio, _ := pipeIO(t, "tail")
	ed := NewEditor(stdio)

	got, err := ed.Line(context.Background(), LineOptions{})
	if err != nil || got != "tail" {
		t.Fatalf("Line() = %q, %v; want unterminated tail line", got, err)
	}
	if _, err := ed.Line(context.Background(), LineOptions{}); !errors.Is(err, io.EOF) {
		t.Fatalf("Line() error = %v, want io.EOF", err)
	}
}

func TestLineFallbackCancelled(t *testing.T) {
	t.Parallel()

	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		_ = inR.Close()
		_ = inW.Close()
		_ = outR.Close()
		_ = outW.Close()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ed := NewEditor(IO{In: inR, Out: outW, Err: io.Discard})
	if _, err := ed.Line(ctx, LineOptions{Prefix: "> "}); err != context.Canceled {
		t.Fatalf("Line(cancelled) error = %v, want context.Canceled", err)
	}
}

func TestAwaitPromptMappings(t *testing.T) {
	t.Parallel()

	if _, err := awaitPrompt(context.Background(), func() (string, error) {
		return "", liner.ErrPromptAborted
	}); err != context.Canceled {
		t.Fatalf("abort error = %v, want context.Canceled", err)
	}
	if _, err := awaitPrompt(context.Background(), func() (string, error) {
		return "", io.EOF
	}); err != io.EOF {
		t.Fatalf("eof error = %v, want io.EOF", err)
	}
	got, err := awaitPrompt(context.Background(), func() (string, error) {
		return "hi", nil
	})
	if err != nil || got != "hi" {
		t.Fatalf("awaitPrompt() = %q, %v; want hi, nil", got, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	unblock := make(chan struct{})
	if _, err := awaitPrompt(ctx, func() (string, error) {
		<-unblock
		return "", nil
	}); err != context.Canceled {
		t.Fatalf("cancelled awaitPrompt error = %v, want context.Canceled", err)
	}
	close(unblock)
}

func TestEditorHistoryRoundTrip(t *testing.T) {
	t.Parallel()

	stdio, _ := pipeIO(t, "")
	ed := NewEditor(stdio)

	n, err := ed.ReadHistory(strings.NewReader("one\ntwo\ntwo\nthree\n"))
	if err != nil || n != 4 {
		t.Fatalf("ReadHistory() = %d, %v; want 4 lines read", n, err)
	}

	var out strings.Builder
	if w, err := ed.WriteHistory(&out); err != nil || w != 3 || out.String() != "one\ntwo\nthree\n" {
		t.Fatalf("WriteHistory() = %d, %v, %q; want 3 rows and immediate repeats collapsed", w, err, out.String())
	}

	ed.ClearHistory()
	out.Reset()
	if w, err := ed.WriteHistory(&out); err != nil || w != 0 || out.Len() != 0 {
		t.Fatalf("WriteHistory() after clear = %d, %v, %q; want empty", w, err, out.String())
	}

	if _, err := ed.ReadHistory(strings.NewReader("bad\xff\n")); err == nil {
		t.Fatal("ReadHistory(invalid utf-8) = nil, want error")
	}
}

func TestEditorHistoryCap(t *testing.T) {
	t.Parallel()

	stdio, _ := pipeIO(t, "")
	ed := NewEditor(stdio)

	ed.AppendHistory("same")
	ed.AppendHistory("same")
	for i := range liner.HistoryLimit + 50 {
		ed.AppendHistory(fmt.Sprintf("line %d", i))
	}

	var out strings.Builder
	n, err := ed.WriteHistory(&out)
	if err != nil || n != liner.HistoryLimit {
		t.Fatalf("WriteHistory() = %d, %v; want capped at %d", n, err, liner.HistoryLimit)
	}
	if !strings.HasPrefix(out.String(), "line 50\n") {
		t.Fatalf("history starts %q, want oldest trimmed to line 50", out.String()[:min(out.Len(), 32)])
	}
}
