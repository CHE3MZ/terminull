package statuscard

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/CHE3MZ/terminull/v2/termapp"
	"github.com/CHE3MZ/terminull/v2/termstyle"
)

func TestCardRenderPlain(t *testing.T) {
	var out bytes.Buffer
	card := Card{Title: "deployment", Message: "all systems normal"}
	if err := card.Render(&out, termstyle.NewScheme(false)); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{"custom component", "deployment", "all systems normal"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Render() = %q, want %q", got, want)
		}
	}
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("Render() = %q, want no ANSI when color is disabled", got)
	}
}

func TestCommandJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("APPDATA", tmp)
	t.Setenv("TERMINULL_ENV", "")

	app := termapp.New("component-test")
	app.Add(Command())

	var out, errOut bytes.Buffer
	app.SetOut(&out)
	app.SetErr(&errOut)
	if err := app.Execute(context.Background(), []string{
		"status", "--message", "degraded", "--output", "json",
	}); err != nil {
		t.Fatalf("Execute() error = %v\n%s", err, errOut.String())
	}

	got := out.String()
	for _, want := range []string{`"title": "deployment"`, `"message": "degraded"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("Execute() = %q, want %q", got, want)
		}
	}
}
