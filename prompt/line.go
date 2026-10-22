package prompt

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/CHE3MZ/terminull/core"
	"github.com/peterh/liner"
)

// Completer returns completion candidates for the text left of the cursor.
type Completer func(line string) []string

// WordCompleter splits the line around the word under the cursor. The line
// "Hello, wo!!!" with the cursor after "wo" arrives as ("Hello, ",
// ["world", "Word"], "!!!") for candidate "Hello, world!!!".
type WordCompleter func(line string, pos int) (head string, completions []string, tail string)

// TabStyle selects how tab completions are shown.
type TabStyle int

const (
	// TabCircular cycles through candidates on the prompt line.
	TabCircular TabStyle = iota
	// TabPrints lists candidates bash-style after a second tab.
	TabPrints
)

// Editor is an opt-in line editor for repl-style input: cursor and word
// editing, kill/yank, history, ctrl+r reverse search, tab completion.
//
// It borrows the process console for the duration of each Line call and
// hands it back before returning, so one-shot prompts and render output
// keep behaving normally between calls. Interactive editing always runs
// on os.Stdin/os.Stdout; IO is used by the plain fallback that Line
// degrades to on pipes, redirected output, and dumb terminals. Line
// returns surrounding whitespace as typed, unlike the trimmed one-shot
// prompts. Editor is not safe for concurrent use.
type Editor struct {
	stdio  IO
	reader *bufio.Reader

	history       []string
	completer     Completer
	wordCompleter WordCompleter
	tabStyle      TabStyle
	multiLine     bool
	beep          bool
}

// NewEditor builds an editor over stdio. Pass StdIO in real apps.
func NewEditor(stdio IO) *Editor {
	return &Editor{
		stdio:  stdio,
		reader: bufio.NewReader(stdio.In),
		beep:   true,
	}
}

// LineOptions tunes a Line call.
type LineOptions struct {
	// Prefix is the text printed before the input, e.g. "you> ". The
	// interactive editor rejects control characters, so keep prefixes
	// plain (no ANSI colors).
	Prefix string
	// Initial is editable text the line starts with; ignored without a
	// terminal, where piped input supplies the line.
	Initial string
}

// Line reads one edited line, newline excluded. Ctrl+C returns
// context.Canceled so repls can drop the line and continue; Ctrl+D and
// end of piped input return io.EOF. Each non-empty line joins history.
func (e *Editor) Line(ctx context.Context, opts LineOptions) (string, error) {
	line, err := e.readLine(ctx, opts)
	if err != nil {
		return "", err
	}
	if line != "" {
		e.AppendHistory(line)
	}
	return line, nil
}

// readLine picks the real editor or the plain read; Line records history.
func (e *Editor) readLine(ctx context.Context, opts LineOptions) (string, error) {
	if !e.editing() {
		return e.fallbackLine(ctx, opts)
	}

	st := liner.NewLiner()
	defer func() { _ = st.Close() }()

	st.SetCtrlCAborts(true)
	st.SetBeep(e.beep)
	st.SetMultiLineMode(e.multiLine)
	st.SetTabCompletionStyle(liner.TabStyle(e.tabStyle))
	if e.completer != nil {
		st.SetCompleter(liner.Completer(e.completer))
	}
	if e.wordCompleter != nil {
		st.SetWordCompleter(liner.WordCompleter(e.wordCompleter))
	}
	for _, item := range e.history {
		st.AppendHistory(item)
	}

	return awaitPrompt(ctx, func() (string, error) {
		return st.PromptWithSuggestion(opts.Prefix, opts.Initial, -1)
	})
}

// editing reports whether the console can host the editor.
func (e *Editor) editing() bool {
	return interactive(e.stdio.In) && interactiveWriter(e.stdio.Out) && liner.TerminalSupported()
}

func interactiveWriter(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return core.IsTerminal(f)
}

// fallbackLine is the pipe and dumb-terminal path: print the prefix and
// read a raw line, so transcripts stay copyable.
func (e *Editor) fallbackLine(ctx context.Context, opts LineOptions) (string, error) {
	if _, err := fmt.Fprint(e.stdio.Out, opts.Prefix); err != nil {
		return "", err
	}
	return readLineRaw(ctx, e.reader)
}

// awaitPrompt runs the blocking editor read, abandoning it on cancel. A
// goroutine may linger in the editor until input arrives; cancelled Line
// calls are shutdown paths, so nothing accumulates.
func awaitPrompt(ctx context.Context, prompt func() (string, error)) (string, error) {
	type result struct {
		s   string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		s, err := prompt()
		ch <- result{s, err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case v := <-ch:
		switch {
		case errors.Is(v.err, liner.ErrPromptAborted):
			return "", context.Canceled
		case v.err != nil:
			return "", v.err
		default:
			return v.s, nil
		}
	}
}

// readLineRaw reads one line holding reader state across calls, so
// buffered piped input is not lost. Only the terminator is stripped.
func readLineRaw(ctx context.Context, r *bufio.Reader) (string, error) {
	type result struct {
		s   string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		s, err := r.ReadString('\n')
		ch <- result{s, err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case v := <-ch:
		s := strings.TrimSuffix(strings.TrimSuffix(v.s, "\n"), "\r")
		if v.err != nil && v.err != io.EOF {
			return "", v.err
		}
		if v.err == io.EOF && s == "" {
			return "", io.EOF
		}
		return s, nil
	}
}

// SetCompleter sets tab completion candidates.
func (e *Editor) SetCompleter(f Completer) {
	e.completer = f
}

// SetWordCompleter sets word-aware tab completion.
func (e *Editor) SetWordCompleter(f WordCompleter) {
	e.wordCompleter = f
}

// SetTabCompletionStyle selects cycling or bash-style listed candidates.
func (e *Editor) SetTabCompletionStyle(style TabStyle) {
	e.tabStyle = style
}

// SetMultiLine wraps long lines across rows instead of scrolling them.
func (e *Editor) SetMultiLine(enabled bool) {
	e.multiLine = enabled
}

// SetBeep toggles the terminal bell (on by default).
func (e *Editor) SetBeep(enabled bool) {
	e.beep = enabled
}

// AppendHistory records an entry, collapsing immediate repeats and
// capping retained entries.
func (e *Editor) AppendHistory(item string) {
	if n := len(e.history); n > 0 && e.history[n-1] == item {
		return
	}
	e.history = append(e.history, item)
	if len(e.history) > liner.HistoryLimit {
		e.history = e.history[len(e.history)-liner.HistoryLimit:]
	}
}

// ClearHistory forgets all entries.
func (e *Editor) ClearHistory() {
	e.history = nil
}

// ReadHistory loads newline-separated history from r, returning the
// number of lines read. Invalid utf-8 lines are rejected.
func (e *Editor) ReadHistory(r io.Reader) (int, error) {
	in := bufio.NewReader(r)
	n := 0
	for {
		line, part, err := in.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return n, err
		}
		if part {
			return n, fmt.Errorf("prompt: history line %d is too long", n+1)
		}
		if !utf8.Valid(line) {
			return n, fmt.Errorf("prompt: history line %d is not utf-8", n+1)
		}
		n++
		e.AppendHistory(string(line))
	}
	return n, nil
}

// WriteHistory writes newline-separated history to w, returning the
// number of entries written.
func (e *Editor) WriteHistory(w io.Writer) (int, error) {
	n := 0
	for _, item := range e.history {
		if _, err := fmt.Fprintln(w, item); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
