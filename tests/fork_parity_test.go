package tests

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/CHE3MZ/terminull/termapp"
	"github.com/CHE3MZ/terminull/termconfig"
	"github.com/CHE3MZ/terminull/termcore"
	"github.com/CHE3MZ/terminull/termoutput"
	"github.com/CHE3MZ/terminull/termprompt"
	"github.com/CHE3MZ/terminull/termretry"
	"github.com/CHE3MZ/terminull/termstyle"
)

// TestForkParity pins one public-boundary snapshot for every forked engine:
// color, retry, humanize, lineedit, koanf, and cobra/pflag.
func TestForkParity(t *testing.T) {
	t.Run("color", func(t *testing.T) {
		got := termstyle.Paint(true, termstyle.FgRed, "error")
		want := "\x1b[31merror\x1b[0m"
		if got != want {
			t.Fatalf("Paint = %q, want %q", got, want)
		}
	})

	t.Run("retry", func(t *testing.T) {
		attempts := 0
		err := termretry.Do(func() error {
			attempts++
			return errors.New("down")
		}, termretry.Options{Attempts: 3, Delay: time.Millisecond})

		got := fmt.Sprintf("%d/%v", attempts, err)
		if want := "3/down"; got != want {
			t.Fatalf("retry = %q, want %q", got, want)
		}
	})

	t.Run("humanize", func(t *testing.T) {
		got := termoutput.Bytes(1500) + "|" + termoutput.IBytes(1048576)
		if want := "1.5 kB|1.0 MiB"; got != want {
			t.Fatalf("humanize = %q, want %q", got, want)
		}
	})

	t.Run("lineedit", func(t *testing.T) {
		inR, inW, err := os.Pipe()
		if err != nil {
			t.Fatalf("os.Pipe() error = %v", err)
		}
		defer func() { _ = inR.Close() }()

		out, err := os.CreateTemp(t.TempDir(), "parity-out")
		if err != nil {
			t.Fatalf("CreateTemp() error = %v", err)
		}
		defer func() { _ = out.Close() }()

		if _, err := io.WriteString(inW, "hello\nworld\n"); err != nil {
			t.Fatalf("WriteString() error = %v", err)
		}
		if err := inW.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		editor := termprompt.NewEditor(termprompt.IO{In: inR, Out: out, Err: io.Discard, NoColor: true})
		first, err := editor.Line(context.Background(), termprompt.LineOptions{Prefix: "> "})
		if err != nil {
			t.Fatalf("Line() error = %v", err)
		}
		second, err := editor.Line(context.Background(), termprompt.LineOptions{Prefix: "> "})
		if err != nil {
			t.Fatalf("Line() error = %v", err)
		}

		var history bytes.Buffer
		if _, err := editor.WriteHistory(&history); err != nil {
			t.Fatalf("WriteHistory() error = %v", err)
		}

		got := fmt.Sprintf("%s|%s|%s", first, second, history.String())
		if want := "hello|world|hello\nworld\n"; got != want {
			t.Fatalf("lineedit = %q, want %q", got, want)
		}
	})

	t.Run("koanf", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "parity.yaml")
		if err := os.WriteFile(path, []byte("name: file\nport: 80\n"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		cfg, err := termconfig.Load(termconfig.Options{
			AppName:    "parity",
			ConfigFile: path,
			Environ:    []string{"PARITY_NAME=env"},
			Overrides:  map[string]any{"name": "flag"},
		})
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		got := fmt.Sprintf("%s:%d", cfg.String("name"), cfg.Int("port"))
		if want := "flag:80"; got != want {
			t.Fatalf("config = %q, want %q", got, want)
		}
	})

	t.Run("cobra-pflag", func(t *testing.T) {
		var out, errOut bytes.Buffer
		app := termapp.New("parity", termapp.WithVersion("v2.0.0"))
		app.SetOut(&out)
		app.SetErr(&errOut)

		if err := app.Execute(context.Background(), []string{"--version"}); err != nil {
			t.Fatalf("Execute(--version) error = %v", err)
		}
		if got, want := strings.TrimSpace(out.String()), "parity version v2.0.0"; got != want {
			t.Fatalf("version = %q, want %q", got, want)
		}

		if err := app.Execute(context.Background(), []string{"--missing"}); err == nil {
			t.Fatal("Execute(--missing) = nil, want error")
		} else if got := termcore.ExitCodeOf(err); got != 2 {
			t.Fatalf("ExitCodeOf(unknown flag) = %d, want 2", got)
		}
	})
}
