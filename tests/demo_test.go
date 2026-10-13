package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/CHE3MZ/terminull/internal/demo"
)

// isolateEnv redirects config search dirs and neutralizes DEMO_ENV so
// tests never read the developer machine or CI host config.
func isolateEnv(t *testing.T) {
	t.Helper()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("APPDATA", tmp)
	t.Setenv("DEMO_ENV", "")
}

func runDemo(t *testing.T, stdin *os.File, args ...string) (string, error) {
	t.Helper()

	var out, errOut bytes.Buffer
	app := demo.NewApp()
	app.SetOut(&out)
	app.SetErr(&errOut)
	if stdin != nil {
		app.SetIn(stdin)
	}
	err := app.Execute(context.Background(), args)
	return out.String() + errOut.String(), err
}

func TestDemoHelpGolden(t *testing.T) {
	isolateEnv(t)

	got, err := runDemo(t, nil, "--help")
	if err != nil {
		t.Fatalf("Execute(--help) error = %v", err)
	}
	for _, want := range []string{"demo", "deploy", "status", "--output"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help %q does not contain %q", got, want)
		}
	}

	got, err = runDemo(t, nil, "deploy", "--help")
	if err != nil {
		t.Fatalf("Execute(deploy --help) error = %v", err)
	}
	if !strings.Contains(got, "--env") {
		t.Fatalf("deploy help %q does not contain --env", got)
	}
}

func TestDemoVersionGolden(t *testing.T) {
	isolateEnv(t)

	got, err := runDemo(t, nil, "--version")
	if err != nil {
		t.Fatalf("Execute(--version) error = %v", err)
	}
	if !strings.Contains(got, "demo version ") {
		t.Fatalf("version %q does not contain stable prefix", got)
	}
}

func TestDeployTable(t *testing.T) {
	isolateEnv(t)

	got, err := runDemo(t, nil, "deploy", "--env", "prod", "--yes")
	if err != nil {
		t.Fatalf("Execute(deploy) error = %v", err)
	}
	for _, want := range []string{"prod", "deployed", "uploading", "Started 'deploy'"} {
		if !strings.Contains(got, want) {
			t.Fatalf("deploy output %q does not contain %q", got, want)
		}
	}
}

func TestDeployJSON(t *testing.T) {
	isolateEnv(t)

	got, err := runDemo(t, nil, "deploy", "--env", "staging", "--yes", "-o", "json")
	if err != nil {
		t.Fatalf("Execute(deploy -o json) error = %v", err)
	}

	// The stream mixes banner, bar, logs, and JSON: cut at the JSON
	// array opener ("[\n"), which ANSI escapes and bar lines never emit.
	start := strings.Index(got, "[\n")
	if start < 0 {
		t.Fatalf("deploy json output %q has no array", got)
	}
	var results []map[string]string
	if err := json.Unmarshal([]byte(got[start:]), &results); err != nil {
		t.Fatalf("Unmarshal() error = %v in %q", err, got)
	}
	if len(results) != 1 || results[0]["env"] != "staging" || results[0]["status"] != "deployed" {
		t.Fatalf("results = %v", results)
	}
}

func TestDeployNeedsEnv(t *testing.T) {
	isolateEnv(t)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	_, err = runDemo(t, r, "deploy", "--yes")
	if err == nil || !strings.Contains(err.Error(), "--env") {
		t.Fatalf("Execute(deploy) error = %v, want --env named", err)
	}
}

func TestStatusFormats(t *testing.T) {
	isolateEnv(t)

	got, err := runDemo(t, nil, "status")
	if err != nil {
		t.Fatalf("Execute(status) error = %v", err)
	}
	for _, want := range []string{"api", "worker", "up"} {
		if !strings.Contains(got, want) {
			t.Fatalf("status table %q does not contain %q", got, want)
		}
	}

	got, err = runDemo(t, nil, "status", "-o", "yaml")
	if err != nil {
		t.Fatalf("Execute(status -o yaml) error = %v", err)
	}
	if !strings.Contains(got, "name: api") {
		t.Fatalf("status yaml %q does not contain service", got)
	}
}
