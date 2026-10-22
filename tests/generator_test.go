package tests

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CHE3MZ/terminull/v2/internal/terminull"
	"github.com/CHE3MZ/terminull/v2/termcore"
)

// readGenerated reads back a file the test itself generated under dir.
func readGenerated(t *testing.T, dir string, elems ...string) string {
	t.Helper()

	// #nosec G304 -- test reads files it generated under t.TempDir().
	content, err := os.ReadFile(filepath.Join(append([]string{dir}, elems...)...))
	if err != nil {
		t.Fatalf("ReadFile(%v) error = %v", elems, err)
	}
	return string(content)
}

// isolateEnv redirects config search dirs and neutralizes framework env
// so tests never read the developer machine or CI host config.
func isolateEnv(t *testing.T) {
	t.Helper()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("APPDATA", tmp)
	t.Setenv("TERMINULL_ENV", "")
}

func runTerminull(t *testing.T, stdin *os.File, args ...string) (string, error) {
	t.Helper()

	var out, errOut bytes.Buffer
	app := terminull.NewApp()
	app.SetOut(&out)
	app.SetErr(&errOut)
	if stdin != nil {
		app.SetIn(stdin)
	}
	err := app.Execute(context.Background(), args)
	return out.String() + errOut.String(), err
}

func TestTerminullHelpGolden(t *testing.T) {
	isolateEnv(t)

	got, err := runTerminull(t, nil, "--help")
	if err != nil {
		t.Fatalf("Execute(--help) error = %v", err)
	}
	for _, want := range []string{"terminull", "scaffold", "--output", "Scaffold batteries-included"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("piped help %q contains ANSI escapes, want plain", got)
	}

	got, err = runTerminull(t, nil, "new", "--help")
	if err != nil {
		t.Fatalf("Execute(new --help) error = %v", err)
	}
	for _, want := range []string{"new NAME", "--module", "--dir", "--replace", "--skip-tidy", "Examples:", "terminull new myapp"} {
		if !strings.Contains(got, want) {
			t.Fatalf("new help %q does not contain %q", got, want)
		}
	}
}

func TestNewScaffoldFiles(t *testing.T) {
	isolateEnv(t)

	dir := filepath.Join(t.TempDir(), "hello")
	got, err := runTerminull(t, nil,
		"new", "hello",
		"--module", "example.com/hello",
		"--dir", dir,
		"--skip-tidy",
	)
	if err != nil {
		t.Fatalf("Execute(new) error = %v\n%s", err, got)
	}

	for _, path := range []string{
		"go.mod",
		filepath.Join("cmd", "hello", "main.go"),
		filepath.Join("internal", "commands", "commands.go"),
		"README.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
			t.Fatalf("Stat(%s) error = %v", path, err)
		}
	}
	if !strings.Contains(got, "go.mod") || !strings.Contains(got, "created") {
		t.Fatalf("new output %q does not list created files", got)
	}
}

func TestNewBadName(t *testing.T) {
	isolateEnv(t)

	_, err := runTerminull(t, nil, "new", "has space", "--skip-tidy")
	if err == nil {
		t.Fatal("Execute(new bad name) = nil, want error")
	}
}

func TestNewNoArgsNonInteractive(t *testing.T) {
	isolateEnv(t)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	_, err = runTerminull(t, r, "new")
	if err == nil || !strings.Contains(err.Error(), "NAME") {
		t.Fatalf("Execute(new) error = %v, want usage naming NAME", err)
	}
	if got := termcore.ExitCodeOf(err); got != 2 {
		t.Fatalf("ExitCodeOf = %d, want 2 for usage errors", got)
	}
}

func TestNewInvalidNameFailsFast(t *testing.T) {
	isolateEnv(t)

	got, err := runTerminull(t, nil, "new", ".", "--skip-tidy")
	if err == nil || !strings.Contains(err.Error(), "invalid app name") {
		t.Fatalf("Execute(new .) error = %v, want validation error", err)
	}
	// Validation runs before any work: no tree output, no files listed.
	if strings.Contains(got, "scaffold' succeeded") || strings.Contains(got, "CREATED") {
		t.Fatalf("output %q shows work done despite invalid name", got)
	}
}

func TestNewModuleDefaultsNonInteractive(t *testing.T) {
	isolateEnv(t)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	dir := filepath.Join(t.TempDir(), "svc")
	if _, err := runTerminull(t, r, "new", "svc", "--dir", dir, "--skip-tidy"); err != nil {
		t.Fatalf("Execute(new) error = %v", err)
	}

	// Piped stdin cannot prompt: the module defaults to the app name.
	content := readGenerated(t, dir, "go.mod")
	if !strings.Contains(content, "module svc\n") {
		t.Fatalf("go.mod %q defaults module to name", content)
	}
}
