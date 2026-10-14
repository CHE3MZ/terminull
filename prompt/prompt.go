package prompt

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/mattn/go-isatty"
)

// IO carries stdio for prompts. Use StdIO in real apps; tests pass pipes.
type IO struct {
	In  terminal.FileReader
	Out terminal.FileWriter
	Err io.Writer
	// NoColor strips prompt icon colors (dumb terminals, --no-color).
	// The bullet glyph itself always renders.
	NoColor bool
}

// StdIO wires the process standard streams for prompts.
func StdIO() IO {
	return IO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr}
}

// askOne is the single survey call site, stubbed in unit tests.
var askOne = func(p survey.Prompt, response any, opts ...survey.AskOpt) error {
	return survey.AskOne(p, response, opts...)
}

// IsInterrupt reports Ctrl+C during a prompt (map to exit code 130).
func IsInterrupt(err error) bool {
	return errors.Is(err, terminal.InterruptErr)
}

func stdio(io IO) []survey.AskOpt {
	iconFormat := "green+hb"
	if io.NoColor {
		iconFormat = ""
	}
	return []survey.AskOpt{
		survey.WithStdio(io.In, io.Out, io.Err),
		survey.WithIcons(func(icons *survey.IconSet) {
			// Bullet instead of survey's default "?". An empty format
			// renders plain (verified against mgutz/ansi); U+2022
			// ships in every modern console font.
			icons.Question.Text = "•"
			icons.Question.Format = iconFormat
		}),
	}
}

func interactive(in terminal.FileReader) bool {
	if in == nil {
		return false
	}
	return isatty.IsTerminal(in.Fd())
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

// Select picks one option with arrow keys and enter.
func Select(io IO, opts SelectOptions) (string, error) {
	if len(opts.Options) == 0 {
		return "", fmt.Errorf("prompt: no options for %s", flagName(opts.Flag))
	}
	if opts.Preset != "" {
		if !slices.Contains(opts.Options, opts.Preset) {
			return "", fmt.Errorf("invalid %s %q: choose from %s", flagName(opts.Flag), opts.Preset, joinOptions(opts.Options))
		}
		return opts.Preset, nil
	}
	if opts.Default != "" && !slices.Contains(opts.Options, opts.Default) {
		return "", fmt.Errorf("invalid default %q for %s: choose from %s", opts.Default, flagName(opts.Flag), joinOptions(opts.Options))
	}
	if !interactive(io.In) {
		return "", fmt.Errorf("%s is required (non-interactive terminal, pass the flag)", flagName(opts.Flag))
	}
	return selectSurvey(io, opts)
}

// selectSurvey prompts unconditionally; Select guards first.
func selectSurvey(io IO, opts SelectOptions) (string, error) {
	var def any
	if opts.Default != "" {
		def = opts.Default
	}
	var answer string
	if err := askOne(&survey.Select{
		Message: opts.Message,
		Options: opts.Options,
		Default: def,
	}, &answer, stdio(io)...); err != nil {
		return "", err
	}
	return answer, nil
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

// MultiSelect picks zero or more options with space and enter.
func MultiSelect(io IO, opts MultiSelectOptions) ([]string, error) {
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
		return nil, fmt.Errorf("%s is required (non-interactive terminal, pass the flag)", flagName(opts.Flag))
	}
	return multiSurvey(io, opts)
}

// multiSurvey prompts unconditionally; MultiSelect guards first.
func multiSurvey(io IO, opts MultiSelectOptions) ([]string, error) {
	var answer []string
	prompt := &survey.MultiSelect{Message: opts.Message, Options: opts.Options}
	if len(opts.Default) > 0 {
		prompt.Default = opts.Default
	}
	if err := askOne(prompt, &answer, stdio(io)...); err != nil {
		return nil, err
	}
	return answer, nil
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
func Confirm(io IO, opts ConfirmOptions) (bool, error) {
	if opts.Preset != nil {
		return *opts.Preset, nil
	}
	if !interactive(io.In) {
		return false, fmt.Errorf("%s is required (non-interactive terminal, pass the flag)", flagName(opts.Flag))
	}
	return confirmSurvey(io, opts)
}

// confirmSurvey prompts unconditionally; Confirm guards first.
func confirmSurvey(io IO, opts ConfirmOptions) (bool, error) {
	var answer bool
	if err := askOne(&survey.Confirm{
		Message: opts.Message,
		Default: opts.Default,
	}, &answer, stdio(io)...); err != nil {
		return false, err
	}
	return answer, nil
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
func Input(io IO, opts InputOptions) (string, error) {
	if opts.Preset != nil {
		return *opts.Preset, nil
	}
	if !interactive(io.In) {
		return "", fmt.Errorf("%s is required (non-interactive terminal, pass the flag)", flagName(opts.Flag))
	}
	return inputSurvey(io, opts)
}

// inputSurvey prompts unconditionally; Input guards first.
func inputSurvey(io IO, opts InputOptions) (string, error) {
	var answer string
	if err := askOne(&survey.Input{
		Message: opts.Message,
		Default: opts.Default,
	}, &answer, stdio(io)...); err != nil {
		return "", err
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

// Password reads one line masked with stars.
func Password(io IO, opts PasswordOptions) (string, error) {
	if opts.Preset != nil {
		return *opts.Preset, nil
	}
	if !interactive(io.In) {
		return "", fmt.Errorf("%s is required (non-interactive terminal, pass the flag)", flagName(opts.Flag))
	}
	return passwordSurvey(io, opts)
}

// passwordSurvey prompts unconditionally; Password guards first.
func passwordSurvey(io IO, opts PasswordOptions) (string, error) {
	var answer string
	if err := askOne(&survey.Password{Message: opts.Message}, &answer, stdio(io)...); err != nil {
		return "", err
	}
	return answer, nil
}
