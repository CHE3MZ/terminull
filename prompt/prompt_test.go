package prompt

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/creack/pty"
)

// pipeIO builds non-terminal stdio. Pipes are never TTYs, so prompt calls
// without presets must fail with flag-required errors instead of hanging.
func pipeIO(t *testing.T) IO {
	t.Helper()

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
	return IO{In: inR, Out: outW, Err: &bytes.Buffer{}}
}

func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }

// stubAskOne replaces the survey call site and restores it afterwards.
// Nothing in this package may use t.Parallel: the shared var would race
// under -race (parallel tests interleave with sequential ones).
func stubAskOne(t *testing.T, stub func(p survey.Prompt, response any, opts ...survey.AskOpt) error) {
	t.Helper()

	previous := askOne
	askOne = stub
	t.Cleanup(func() { askOne = previous })
}

func failOnPrompt(t *testing.T) {
	t.Helper()
	stubAskOne(t, func(_ survey.Prompt, _ any, _ ...survey.AskOpt) error {
		t.Fatal("survey must not prompt on the preset path")
		return nil
	})
}

func TestSelectPreset(t *testing.T) {
	failOnPrompt(t)

	got, err := Select(pipeIO(t), SelectOptions{
		Message: "Env?",
		Options: []string{"dev", "prod"},
		Preset:  "prod",
		Flag:    "--env",
	})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if got != "prod" {
		t.Fatalf("Select() = %q, want preset prod", got)
	}
}

func TestSelectBadPreset(t *testing.T) {
	failOnPrompt(t)

	_, err := Select(pipeIO(t), SelectOptions{
		Options: []string{"dev", "prod"},
		Preset:  "qa",
		Flag:    "--env",
	})
	if err == nil || !strings.Contains(err.Error(), `"qa"`) || !strings.Contains(err.Error(), "--env") {
		t.Fatalf("Select() error = %v, want value and flag named", err)
	}
}

func TestSelectEmptyOptions(t *testing.T) {
	_, err := Select(pipeIO(t), SelectOptions{Flag: "--env"})
	if err == nil {
		t.Fatal("Select() = nil, want error for empty options")
	}
}

func TestSelectNonInteractive(t *testing.T) {
	failOnPrompt(t)

	_, err := Select(pipeIO(t), SelectOptions{
		Message: "Env?",
		Options: []string{"dev", "prod"},
		Flag:    "--env",
	})
	if err == nil || !strings.Contains(err.Error(), "--env") {
		t.Fatalf("Select() error = %v, want required flag named", err)
	}
}

func TestSelectBadDefault(t *testing.T) {
	_, err := Select(pipeIO(t), SelectOptions{
		Options: []string{"dev"},
		Default: "qa",
		Flag:    "--env",
	})
	if err == nil || !strings.Contains(err.Error(), `"qa"`) {
		t.Fatalf("Select() error = %v, want bad default named", err)
	}
}

func TestSelectSurveyWiring(t *testing.T) {
	stubAskOne(t, func(p survey.Prompt, response any, opts ...survey.AskOpt) error {
		sel, ok := p.(*survey.Select)
		if !ok {
			t.Fatalf("prompt type = %T, want *survey.Select", p)
		}
		if sel.Message != "Env?" || len(sel.Options) != 2 || sel.Options[1] != "prod" || sel.Default != "dev" {
			t.Fatalf("select = %+v, want message/options/default wired", sel)
		}
		if len(opts) != 2 {
			t.Fatalf("ask opts = %d, want stdio + icons wired", len(opts))
		}
		*(response.(*string)) = "prod"
		return nil
	})

	got, err := selectSurvey(pipeIO(t), SelectOptions{
		Message: "Env?",
		Options: []string{"dev", "prod"},
		Default: "dev",
	})
	if err != nil {
		t.Fatalf("selectSurvey() error = %v", err)
	}
	if got != "prod" {
		t.Fatalf("selectSurvey() = %q, want stubbed answer", got)
	}
}

func TestSurveyErrorPropagates(t *testing.T) {
	stubAskOne(t, func(_ survey.Prompt, _ any, _ ...survey.AskOpt) error {
		return terminal.InterruptErr
	})

	if _, err := selectSurvey(pipeIO(t), SelectOptions{Options: []string{"a"}}); !IsInterrupt(err) {
		t.Fatalf("selectSurvey() error = %v, want interrupt", err)
	}
}

func TestConfirmPresetAndGuard(t *testing.T) {
	failOnPrompt(t)

	io := pipeIO(t)
	if got, err := Confirm(io, ConfirmOptions{Preset: boolPtr(true)}); err != nil || !got {
		t.Fatalf("Confirm(preset) = %v, %v; want true, nil", got, err)
	}
	if _, err := Confirm(io, ConfirmOptions{Message: "Sure?", Flag: "--yes"}); err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("Confirm() error = %v, want required flag named", err)
	}
}

func TestConfirmSurveyWiring(t *testing.T) {
	stubAskOne(t, func(p survey.Prompt, response any, _ ...survey.AskOpt) error {
		c, ok := p.(*survey.Confirm)
		if !ok {
			t.Fatalf("prompt type = %T, want *survey.Confirm", p)
		}
		if c.Message != "Sure?" || !c.Default {
			t.Fatalf("confirm = %+v, want message/default wired", c)
		}
		*(response.(*bool)) = true
		return nil
	})

	got, err := confirmSurvey(pipeIO(t), ConfirmOptions{Message: "Sure?", Default: true})
	if err != nil || !got {
		t.Fatalf("confirmSurvey() = %v, %v; want true, nil", got, err)
	}
}

func TestInputPresetGuardWiring(t *testing.T) {
	failOnPrompt(t)

	io := pipeIO(t)
	if got, err := Input(io, InputOptions{Preset: stringPtr("")}); err != nil || got != "" {
		t.Fatalf("Input(empty preset) = %q, %v; want empty, nil", got, err)
	}
	if _, err := Input(io, InputOptions{Message: "Name?", Flag: "--name"}); err == nil {
		t.Fatal("Input() = nil, want required-flag error")
	}
}

func TestInputSurveyWiring(t *testing.T) {
	stubAskOne(t, func(p survey.Prompt, response any, _ ...survey.AskOpt) error {
		in, ok := p.(*survey.Input)
		if !ok {
			t.Fatalf("prompt type = %T, want *survey.Input", p)
		}
		if in.Message != "Name?" || in.Default != "bob" {
			t.Fatalf("input = %+v, want message/default wired", in)
		}
		*(response.(*string)) = "alice"
		return nil
	})

	got, err := inputSurvey(pipeIO(t), InputOptions{Message: "Name?", Default: "bob"})
	if err != nil || got != "alice" {
		t.Fatalf("inputSurvey() = %q, %v; want alice, nil", got, err)
	}
}

func TestPasswordPresetGuardWiring(t *testing.T) {
	failOnPrompt(t)

	io := pipeIO(t)
	if got, err := Password(io, PasswordOptions{Preset: stringPtr("s3cret")}); err != nil || got != "s3cret" {
		t.Fatalf("Password(preset) = %q, %v", got, err)
	}
	if _, err := Password(io, PasswordOptions{Message: "Pass?", Flag: "--password"}); err == nil {
		t.Fatal("Password() = nil, want required-flag error")
	}
}

func TestPasswordSurveyWiring(t *testing.T) {
	stubAskOne(t, func(p survey.Prompt, response any, _ ...survey.AskOpt) error {
		pw, ok := p.(*survey.Password)
		if !ok {
			t.Fatalf("prompt type = %T, want *survey.Password", p)
		}
		if pw.Message != "Pass?" {
			t.Fatalf("password message = %q", pw.Message)
		}
		*(response.(*string)) = "s3cret"
		return nil
	})

	got, err := passwordSurvey(pipeIO(t), PasswordOptions{Message: "Pass?"})
	if err != nil || got != "s3cret" {
		t.Fatalf("passwordSurvey() = %q, %v", got, err)
	}
}

func TestMultiSelectPresetAndGuard(t *testing.T) {
	failOnPrompt(t)

	io := pipeIO(t)
	opts := MultiSelectOptions{Options: []string{"a", "b"}, Flag: "--tags"}

	got, err := MultiSelect(io, MultiSelectOptions{Options: opts.Options, Preset: []string{"a"}, Flag: "--tags"})
	if err != nil || len(got) != 1 || got[0] != "a" {
		t.Fatalf("MultiSelect(preset) = %v, %v", got, err)
	}
	if _, err := MultiSelect(io, MultiSelectOptions{Options: opts.Options, Preset: []string{"z"}, Flag: "--tags"}); err == nil {
		t.Fatal("MultiSelect(bad preset) = nil, want error")
	}
	if _, err := MultiSelect(io, opts); err == nil {
		t.Fatal("MultiSelect() = nil, want required-flag error")
	}
}

func TestMultiSurveyWiring(t *testing.T) {
	stubAskOne(t, func(p survey.Prompt, response any, _ ...survey.AskOpt) error {
		ms, ok := p.(*survey.MultiSelect)
		if !ok {
			t.Fatalf("prompt type = %T, want *survey.MultiSelect", p)
		}
		if ms.Message != "Tags?" || len(ms.Options) != 2 {
			t.Fatalf("multi = %+v, want message/options wired", ms)
		}
		*(response.(*[]string)) = []string{"b"}
		return nil
	})

	got, err := multiSurvey(pipeIO(t), MultiSelectOptions{
		Message: "Tags?",
		Options: []string{"a", "b"},
		Default: []string{"a"},
	})
	if err != nil || len(got) != 1 || got[0] != "b" {
		t.Fatalf("multiSurvey() = %v, %v", got, err)
	}
}

func TestIsInterrupt(t *testing.T) {
	if !IsInterrupt(terminal.InterruptErr) {
		t.Fatal("IsInterrupt(interrupt) = false")
	}
	if IsInterrupt(errors.New("boom")) || IsInterrupt(nil) {
		t.Fatal("IsInterrupt(other/nil) = true")
	}
}

func TestBulletQuestionIcon(t *testing.T) {
	var opts survey.AskOptions
	for _, opt := range stdio(pipeIO(t)) {
		if err := opt(&opts); err != nil {
			t.Fatalf("AskOpt error = %v", err)
		}
	}
	if got := opts.PromptConfig.Icons.Question.Text; got != "•" {
		t.Fatalf("question icon = %q, want bullet", got)
	}
}

// TestInputPtyRenderDiagnostic drives a real survey prompt through a
// pseudo-terminal with pre-fed input, then logs the raw byte stream and
// the prompt render count. Diagnostic only: it documents real-console
// rendering (double-render reports) on each OS without asserting a
// count, which varies by terminal backend. Skip when no pty exists.
func TestInputPtyRenderDiagnostic(t *testing.T) {
	master, slave, err := pty.Open()
	if err != nil {
		t.Skipf("no pty: %v", err)
	}
	t.Cleanup(func() {
		_ = master.Close()
		_ = slave.Close()
	})

	pio := IO{In: slave, Out: slave, Err: &bytes.Buffer{}}
	done := make(chan struct{})
	var answer string
	var askErr error
	go func() {
		defer close(done)
		answer, askErr = inputSurvey(pio, InputOptions{Message: "App name?", Flag: "NAME"})
	}()

	// The pty buffers input, so feed timing is lenient.
	time.Sleep(time.Second)
	if _, err := master.Write([]byte("IDK\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("inputSurvey hung on pty input")
	}

	// Let survey flush and exit, then drain whatever arrived.
	time.Sleep(2 * time.Second)
	_ = slave.Close()
	raw, _ := io.ReadAll(master)

	if askErr != nil {
		t.Fatalf("inputSurvey() error = %v", askErr)
	}
	if answer != "IDK" {
		t.Fatalf("inputSurvey() = %q, want IDK", answer)
	}
	t.Logf("pty bytes: %q", raw)
	t.Logf("prompt renders: %d", strings.Count(string(raw), "App name?"))
}
