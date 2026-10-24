package termapp

import (
	"bytes"
	"strings"
	"testing"
)

func TestManPages(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	var run testDeployRun
	app := newTestApp(&out, &errOut, &run)
	app.Add(Command{Use: "secret", Short: "hidden command", Hidden: true})

	pages, err := app.ManPages()
	if err != nil {
		t.Fatalf("ManPages() error = %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("ManPages() returned %d pages, want 2", len(pages))
	}

	root, ok := pages["demo"]
	if !ok {
		t.Fatal("ManPages() missing root page")
	}
	for _, want := range []string{".TH \"DEMO\"", ".SH NAME", ".SH SYNOPSIS", ".SH COMMANDS", ".SH OPTIONS", "\\-\\-output", "deploy"} {
		if !strings.Contains(string(root), want) {
			t.Errorf("root man page missing %q:\n%s", want, root)
		}
	}

	child, ok := pages["demo-deploy"]
	if !ok {
		t.Fatal("ManPages() missing deploy page")
	}
	for _, want := range []string{".TH \"DEMO-DEPLOY\"", ".SH NAME", "\\-\\-env", "\\-\\-output"} {
		if !strings.Contains(string(child), want) {
			t.Errorf("deploy man page missing %q:\n%s", want, child)
		}
	}
	if strings.Contains(string(root), "secret") {
		t.Fatal("hidden command leaked into man pages")
	}
}

func TestRoffTextEscapes(t *testing.T) {
	t.Parallel()

	got := roffText(".danger - dash \\ slash")
	want := "\\&.danger \\- dash \\\\ slash"
	if got != want {
		t.Fatalf("roffText() = %q, want %q", got, want)
	}
}
