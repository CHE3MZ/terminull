package termprompt

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/CHE3MZ/terminull/termcore"
	"github.com/CHE3MZ/terminull/termstyle"
)

// FileReader is a readable file descriptor: consoles, pipes, files.
type FileReader interface {
	io.Reader
	Fd() uintptr
}

// FileWriter is a writable file descriptor.
type FileWriter interface {
	io.Writer
	Fd() uintptr
}

// IO carries stdio for prompts. Use StdIO in real apps; tests pass pipes.
type IO struct {
	In  FileReader
	Out FileWriter
	Err io.Writer
	// NoColor strips prompt styling (dumb terminals, --no-color).
	// The bullet glyph itself always renders.
	NoColor bool
}

// StdIO wires the process standard streams for prompts.
func StdIO() IO {
	return IO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr}
}

// paint styles the bullet green when colors are allowed. The gate is
// termcore.Color on the actual destination: NoColor flag, NO_COLOR, dumb
// terminals, CI, and non-terminal writers all stay plain.
func paint(w FileWriter, noColor bool, s string) string {
	f, ok := w.(*os.File)
	if !ok {
		return s
	}
	return termstyle.NewScheme(termcore.Color(f, noColor)).Green(s)
}

func interactive(in io.Reader) bool {
	f, ok := in.(*os.File)
	if !ok {
		return false
	}
	return termcore.IsTerminal(f)
}

func flagName(flag string) string {
	if flag == "" {
		return "flag"
	}
	return flag
}

func joinOptions(options []string) string {
	return strings.Join(options, ", ")
}

func requiredError(flag string) error {
	return fmt.Errorf("%s is required (non-interactive terminal, pass the flag)", flagName(flag))
}

// readLine reads one trimmed line, abandoning a blocked read on cancel.
// A goroutine may linger on a cancelled read until input arrives; the
// process exits right after, so nothing accumulates.
func readLine(ctx context.Context, r io.Reader) (string, error) {
	type result struct {
		s   string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		s, err := bufio.NewReader(r).ReadString('\n')
		ch <- result{s, err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case v := <-ch:
		if v.err != nil && v.err != io.EOF {
			return "", v.err
		}
		return strings.TrimSpace(v.s), nil
	}
}

// ask prints the bullet question and returns the raw answer line.
func ask(ctx context.Context, io IO, message, hint string) (string, error) {
	line := paint(io.Out, io.NoColor, "•") + " " + message
	if hint != "" {
		line += " " + hint
	}
	if _, err := fmt.Fprintln(io.Out, line); err != nil {
		return "", err
	}
	return readLine(ctx, io.In)
}

// SelectOptions tunes a Select call.
type SelectOptions struct {
	// Message is the question, e.g. "Target environment?".
	Message string
	// Options are the choices in order.
	Options []string
	// Default preselects when prompting (must be one of Options).
	Default string
	// Preset is the parsed flag value: validated and returned unprompted.
	Preset string
	// Flag names the flag for errors, e.g. "--env".
	Flag string
}

// Select picks one option from a numbered list.
func Select(ctx context.Context, io IO, opts SelectOptions) (string, error) {
	if len(opts.Options) == 0 {
		return "", fmt.Errorf("prompt: no options for %s", flagName(opts.Flag))
	}
	if opts.Preset != "" {
		if !slices.Contains(opts.Options, opts.Preset) {
			return "", fmt.Errorf("invalid %s %q: choose from %s", flagName(opts.Flag), opts.Preset, joinOptions(opts.Options))
		}
		return opts.Preset, nil
	}
	def := opts.Default
	if def == "" {
		def = opts.Options[0]
	}
	if !slices.Contains(opts.Options, def) {
		return "", fmt.Errorf("invalid default %q for %s: choose from %s", opts.Default, flagName(opts.Flag), joinOptions(opts.Options))
	}
	if !interactive(io.In) {
		return "", requiredError(opts.Flag)
	}
	return selectPrompt(ctx, io, opts)
}

// selectPrompt asks unconditionally; Select guards first.
func selectPrompt(ctx context.Context, io IO, opts SelectOptions) (string, error) {
	def := opts.Default
	if def == "" {
		def = opts.Options[0]
	}

	var list strings.Builder
	for i, opt := range opts.Options {
		fmt.Fprintf(&list, "  %d) %s\n", i+1, opt)
	}
	if _, err := fmt.Fprint(io.Out, list.String()); err != nil {
		return "", err
	}
	answer, err := ask(ctx, io, opts.Message, fmt.Sprintf("[%d]", indexOf(opts.Options, def)+1))
	if err != nil {
		return "", err
	}
	if answer == "" {
		return def, nil
	}
	n, err := strconv.Atoi(answer)
	if err != nil || n < 1 || n > len(opts.Options) {
		return "", fmt.Errorf("invalid %s %q: enter 1-%d", flagName(opts.Flag), answer, len(opts.Options))
	}
	return opts.Options[n-1], nil
}

func indexOf(options []string, value string) int {
	for i, opt := range options {
		if opt == value {
			return i
		}
	}
	return 0
}

// MultiSelectOptions tunes a MultiSelect call.
type MultiSelectOptions struct {
	Message string
	Options []string
	// Default preselects (subset of Options).
	Default []string
	// Preset returns unprompted when non-nil (validated subset).
	Preset []string
	Flag   string
}

// MultiSelect picks options by comma-separated numbers.
func MultiSelect(ctx context.Context, io IO, opts MultiSelectOptions) ([]string, error) {
	if len(opts.Options) == 0 {
		return nil, fmt.Errorf("prompt: no options for %s", flagName(opts.Flag))
	}
	if opts.Preset != nil {
		if bad := firstOutside(opts.Preset, opts.Options); bad != "" {
			return nil, fmt.Errorf("invalid %s %q: choose from %s", flagName(opts.Flag), bad, joinOptions(opts.Options))
		}
		return opts.Preset, nil
	}
	if bad := firstOutside(opts.Default, opts.Options); bad != "" {
		return nil, fmt.Errorf("invalid default %q for %s: choose from %s", bad, flagName(opts.Flag), joinOptions(opts.Options))
	}
	if !interactive(io.In) {
		return nil, requiredError(opts.Flag)
	}
	return multiPrompt(ctx, io, opts)
}

// multiPrompt asks unconditionally; MultiSelect guards first.
func multiPrompt(ctx context.Context, io IO, opts MultiSelectOptions) ([]string, error) {
	var list strings.Builder
	for i, opt := range opts.Options {
		fmt.Fprintf(&list, "  %d) %s\n", i+1, opt)
	}
	if _, err := fmt.Fprint(io.Out, list.String()); err != nil {
		return nil, err
	}
	answer, err := ask(ctx, io, opts.Message, "[1,2]")
	if err != nil {
		return nil, err
	}
	if answer == "" {
		return opts.Default, nil
	}
	var picked []string
	seen := map[string]bool{}
	for _, part := range strings.Split(answer, ",") {
		part = strings.TrimSpace(part)
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 || n > len(opts.Options) {
			return nil, fmt.Errorf("invalid %s %q: enter numbers like 1,3", flagName(opts.Flag), answer)
		}
		if value := opts.Options[n-1]; !seen[value] {
			seen[value] = true
			picked = append(picked, value)
		}
	}
	return picked, nil
}

func firstOutside(values, options []string) string {
	for _, v := range values {
		if !slices.Contains(options, v) {
			return v
		}
	}
	return ""
}

// ConfirmOptions tunes a Confirm call.
type ConfirmOptions struct {
	Message string
	Default bool
	// Preset returns unprompted when non-nil.
	Preset *bool
	Flag   string
}

// Confirm asks yes/no.
func Confirm(ctx context.Context, io IO, opts ConfirmOptions) (bool, error) {
	if opts.Preset != nil {
		return *opts.Preset, nil
	}
	if !interactive(io.In) {
		return false, requiredError(opts.Flag)
	}
	return confirmPrompt(ctx, io, opts)
}

// confirmPrompt asks unconditionally; Confirm guards first.
func confirmPrompt(ctx context.Context, io IO, opts ConfirmOptions) (bool, error) {
	hint := "[y/N]"
	if opts.Default {
		hint = "[Y/n]"
	}
	answer, err := ask(ctx, io, opts.Message, hint)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(answer) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	case "":
		return opts.Default, nil
	default:
		return false, fmt.Errorf("invalid %s %q: answer y or n", flagName(opts.Flag), answer)
	}
}

// InputOptions tunes an Input call.
type InputOptions struct {
	Message string
	Default string
	// Preset returns unprompted when non-nil (empty string is valid input).
	Preset *string
	Flag   string
}

// Input reads one line, echoing input.
func Input(ctx context.Context, io IO, opts InputOptions) (string, error) {
	if opts.Preset != nil {
		return *opts.Preset, nil
	}
	if !interactive(io.In) {
		return "", requiredError(opts.Flag)
	}
	return inputPrompt(ctx, io, opts)
}

// inputPrompt asks unconditionally; Input guards first.
func inputPrompt(ctx context.Context, io IO, opts InputOptions) (string, error) {
	hint := ""
	if opts.Default != "" {
		hint = "(" + opts.Default + ")"
	}
	answer, err := ask(ctx, io, opts.Message, hint)
	if err != nil {
		return "", err
	}
	if answer == "" {
		return opts.Default, nil
	}
	return answer, nil
}

// PasswordOptions tunes a Password call.
type PasswordOptions struct {
	Message string
	// Preset returns unprompted when non-nil.
	Preset *string
	Flag   string
}

// Password reads one line without echoing.
func Password(ctx context.Context, io IO, opts PasswordOptions) (string, error) {
	if opts.Preset != nil {
		return *opts.Preset, nil
	}
	f, ok := io.In.(*os.File)
	if !ok || !termcore.IsTerminal(f) {
		return "", requiredError(opts.Flag)
	}
	if _, err := fmt.Fprintln(io.Out, paint(io.Out, io.NoColor, "•")+" "+opts.Message); err != nil {
		return "", err
	}
	secret, err := readPassword(ctx, f)
	if err != nil {
		return "", err
	}
	return secret, nil
}

// readPassword reads without echo, abandoning a blocked read on cancel.
func readPassword(ctx context.Context, f *os.File) (string, error) {
	type result struct {
		s   []byte
		err error
	}
	ch := make(chan result, 1)
	go func() {
		s, err := termcore.ReadPassword(int(f.Fd()))
		ch <- result{s, err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case v := <-ch:
		if v.err != nil {
			return "", v.err
		}
		return string(v.s), nil
	}
}
