package termprompt

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

// pipeIO builds non-terminal stdio: writes feed answers, reads capture
// output. Pipes are never TTYs, so prompt calls without presets must
// fail with flag-required errors instead of hanging.
func pipeIO(t *testing.T, input string) (IO, func() string) {
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
	if _, err := inW.Write([]byte(input)); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	_ = inW.Close()
	drain := func() string {
		_ = outW.Close()
		data, _ := io.ReadAll(outR)
		_ = outR.Close()
		return string(data)
	}
	return IO{In: inR, Out: outW, Err: io.Discard}, drain
}

func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }

func TestSelectPreset(t *testing.T) {
	t.Parallel()

	io, _ := pipeIO(t, "")
	got, err := Select(context.Background(), io, SelectOptions{
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
	t.Parallel()

	io, _ := pipeIO(t, "")
	_, err := Select(context.Background(), io, SelectOptions{
		Options: []string{"dev", "prod"},
		Preset:  "qa",
		Flag:    "--env",
	})
	if err == nil || !strings.Contains(err.Error(), `"qa"`) || !strings.Contains(err.Error(), "--env") {
		t.Fatalf("Select() error = %v, want value and flag named", err)
	}
}

func TestSelectEmptyOptions(t *testing.T) {
	t.Parallel()

	io, _ := pipeIO(t, "")
	_, err := Select(context.Background(), io, SelectOptions{Flag: "--env"})
	if err == nil {
		t.Fatal("Select() = nil, want error for empty options")
	}
}

func TestSelectNonInteractive(t *testing.T) {
	t.Parallel()

	io, _ := pipeIO(t, "")
	_, err := Select(context.Background(), io, SelectOptions{
		Message: "Env?",
		Options: []string{"dev", "prod"},
		Flag:    "--env",
	})
	if err == nil || !strings.Contains(err.Error(), "--env") {
		t.Fatalf("Select() error = %v, want required flag named", err)
	}
}

func TestSelectBadDefault(t *testing.T) {
	t.Parallel()

	io, _ := pipeIO(t, "")
	_, err := Select(context.Background(), io, SelectOptions{
		Options: []string{"dev"},
		Default: "qa",
		Flag:    "--env",
	})
	if err == nil || !strings.Contains(err.Error(), `"qa"`) {
		t.Fatalf("Select() error = %v, want bad default named", err)
	}
}

func TestSelectNumbered(t *testing.T) {
	t.Parallel()

	io, drain := pipeIO(t, "2\n")
	got, err := selectPrompt(context.Background(), io, SelectOptions{
		Message: "Env?",
		Options: []string{"dev", "prod"},
		Flag:    "--env",
	})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if got != "prod" {
		t.Fatalf("Select() = %q, want second option", got)
	}
	gotOut := drain()
	for _, want := range []string{"•", "Env?", "1) dev", "2) prod", "[1]"} {
		if !strings.Contains(gotOut, want) {
			t.Fatalf("select output %q does not contain %q", gotOut, want)
		}
	}
}

func TestSelectDefaultOnEmpty(t *testing.T) {
	t.Parallel()

	io, _ := pipeIO(t, "\n")
	got, err := selectPrompt(context.Background(), io, SelectOptions{
		Options: []string{"dev", "prod"},
		Default: "prod",
		Flag:    "--env",
	})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if got != "prod" {
		t.Fatalf("Select() = %q, want default", got)
	}
}

func TestSelectBadNumber(t *testing.T) {
	t.Parallel()

	io, _ := pipeIO(t, "9\n")
	_, err := selectPrompt(context.Background(), io, SelectOptions{
		Options: []string{"dev", "prod"},
		Flag:    "--env",
	})
	if err == nil {
		t.Fatal("Select(9) = nil, want range error")
	}
}

func TestConfirmPresetAndGuard(t *testing.T) {
	t.Parallel()

	io, _ := pipeIO(t, "")
	if got, err := Confirm(context.Background(), io, ConfirmOptions{Preset: boolPtr(true)}); err != nil || !got {
		t.Fatalf("Confirm(preset) = %v, %v; want true, nil", got, err)
	}
	if _, err := Confirm(context.Background(), io, ConfirmOptions{Message: "Sure?", Flag: "--yes"}); err == nil {
		t.Fatal("Confirm() = nil, want required-flag error")
	}
}

func TestConfirmAnswers(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in   string
		def  bool
		want bool
	}{
		{"y\n", false, true},
		{"YES\n", false, true},
		{"n\n", true, false},
		{"\n", true, true},
		{"\n", false, false},
	} {
		io, drain := pipeIO(t, tc.in)
		got, err := confirmPrompt(context.Background(), io, ConfirmOptions{Message: "Sure?", Default: tc.def})
		if err != nil {
			t.Fatalf("Confirm(%q) error = %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("Confirm(%q) = %v, want %v", tc.in, got, tc.want)
		}
		if gotOut := drain(); !strings.Contains(gotOut, "Sure?") {
			t.Fatalf("confirm output %q misses message", gotOut)
		}
	}

	io, _ := pipeIO(t, "maybe\n")
	if _, err := Confirm(context.Background(), io, ConfirmOptions{Message: "Sure?"}); err == nil {
		t.Fatal("Confirm(maybe) = nil, want error")
	}
}

func TestInputPresetGuardAnswers(t *testing.T) {
	t.Parallel()

	io, _ := pipeIO(t, "")
	if got, err := Input(context.Background(), io, InputOptions{Preset: stringPtr("")}); err != nil || got != "" {
		t.Fatalf("Input(empty preset) = %q, %v; want empty, nil", got, err)
	}
	if _, err := Input(context.Background(), io, InputOptions{Message: "Name?", Flag: "--name"}); err == nil {
		t.Fatal("Input() = nil, want required-flag error")
	}

	io, drain := pipeIO(t, "alice\n")
	got, err := inputPrompt(context.Background(), io, InputOptions{Message: "Name?", Default: "bob"})
	if err != nil || got != "alice" {
		t.Fatalf("Input() = %q, %v; want alice, nil", got, err)
	}
	if gotOut := drain(); !strings.Contains(gotOut, "Name?") || !strings.Contains(gotOut, "(bob)") {
		t.Fatalf("input output %q misses message and default", gotOut)
	}

	io, _ = pipeIO(t, "\n")
	got, err = inputPrompt(context.Background(), io, InputOptions{Message: "Name?", Default: "bob"})
	if err != nil || got != "bob" {
		t.Fatalf("Input(empty) = %q, %v; want default bob", got, err)
	}
}

func TestInputPromptStaysOnOneLine(t *testing.T) {
	t.Parallel()

	io, drain := pipeIO(t, "alice\n")
	got, err := inputPrompt(context.Background(), io, InputOptions{Message: "Name?", Default: "bob"})
	if err != nil || got != "alice" {
		t.Fatalf("Input() = %q, %v; want alice, nil", got, err)
	}
	if gotOut := drain(); gotOut != "• Name? (bob)" {
		t.Fatalf("input prompt output %q, want inline prompt", gotOut)
	}
}

func TestPasswordPresetGuard(t *testing.T) {
	t.Parallel()

	io, _ := pipeIO(t, "")
	if got, err := Password(context.Background(), io, PasswordOptions{Preset: stringPtr("s3cret")}); err != nil || got != "s3cret" {
		t.Fatalf("Password(preset) = %q, %v", got, err)
	}
	if _, err := Password(context.Background(), io, PasswordOptions{Message: "Pass?", Flag: "--password"}); err == nil {
		t.Fatal("Password() = nil, want required-flag error")
	}
}

func TestMultiSelectPresetAndGuard(t *testing.T) {
	t.Parallel()

	io, _ := pipeIO(t, "")
	opts := MultiSelectOptions{Options: []string{"a", "b"}, Flag: "--tags"}

	got, err := MultiSelect(context.Background(), io, MultiSelectOptions{Options: opts.Options, Preset: []string{"a"}, Flag: "--tags"})
	if err != nil || len(got) != 1 || got[0] != "a" {
		t.Fatalf("MultiSelect(preset) = %v, %v", got, err)
	}
	if _, err := MultiSelect(context.Background(), io, MultiSelectOptions{Options: opts.Options, Preset: []string{"z"}, Flag: "--tags"}); err == nil {
		t.Fatal("MultiSelect(bad preset) = nil, want error")
	}
	if _, err := MultiSelect(context.Background(), io, opts); err == nil {
		t.Fatal("MultiSelect() = nil, want required-flag error")
	}
}

func TestMultiSelectNumbers(t *testing.T) {
	t.Parallel()

	io, drain := pipeIO(t, "2,1,2\n")
	got, err := multiPrompt(context.Background(), io, MultiSelectOptions{
		Message: "Tags?",
		Options: []string{"a", "b"},
		Flag:    "--tags",
	})
	if err != nil {
		t.Fatalf("MultiSelect() error = %v", err)
	}
	if len(got) != 2 || got[0] != "b" || got[1] != "a" {
		t.Fatalf("MultiSelect() = %v, want [b a] deduplicated in order", got)
	}
	if gotOut := drain(); !strings.Contains(gotOut, "1) a") {
		t.Fatalf("multi output %q misses numbered list", gotOut)
	}

	io, _ = pipeIO(t, "\n")
	got, err = multiPrompt(context.Background(), io, MultiSelectOptions{
		Options: []string{"a", "b"},
		Default: []string{"a"},
		Flag:    "--tags",
	})
	if err != nil || len(got) != 1 || got[0] != "a" {
		t.Fatalf("MultiSelect(empty) = %v, %v; want default", got, err)
	}

	io, _ = pipeIO(t, "7\n")
	if _, err := MultiSelect(context.Background(), io, MultiSelectOptions{Options: []string{"a"}, Flag: "--tags"}); err == nil {
		t.Fatal("MultiSelect(7) = nil, want error")
	}
}

func TestReadLineCancelled(t *testing.T) {
	t.Parallel()

	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		_ = inR.Close()
		_ = inW.Close()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := readLine(ctx, inR); err != context.Canceled {
		t.Fatalf("readLine(cancelled) error = %v, want context.Canceled", err)
	}
}
