package pager

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func okLookPath(prog string) (string, error) { return "/bin/" + prog, nil }

func failLookPath(prog string) (string, error) { return "", errMissing }

var errMissing = errorString("missing")

type errorString string

func (e errorString) Error() string { return string(e) }

func TestResolvePager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flag     string
		env      string
		term     string
		lookPath func(string) (string, error)
		wantProg string
		wantArgs []string
		wantOK   bool
	}{
		{"dumb terminal", "", "", "dumb", okLookPath, "", nil, false},
		{"dumb case-insensitive", "", "", "DUMB", okLookPath, "", nil, false},
		{"cat declines", "cat", "", "xterm", okLookPath, "", nil, false},
		{"cat exe declines", "cat.exe", "", "xterm", okLookPath, "", nil, false},
		{"unparseable", `"oops`, "", "xterm", okLookPath, "", nil, false},
		{"missing binary", "less", "", "xterm", failLookPath, "", nil, false},
		{"explicit with args", "less -R", "", "xterm", okLookPath, "/bin/less", []string{"-R"}, true},
		{"env fallback", "", "more", "xterm", okLookPath, "/bin/more", nil, true},
		{"flag beats env", "less", "more", "xterm", okLookPath, "/bin/less", nil, true},
		{"default less", "", "", "xterm", okLookPath, "/bin/less", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prog, args, ok := resolvePager(tt.flag, tt.env, tt.term, tt.lookPath)
			if ok != tt.wantOK || prog != tt.wantProg || !equalStrings(args, tt.wantArgs) {
				t.Fatalf("resolvePager() = (%q, %q, %v), want (%q, %q, %v)",
					prog, args, ok, tt.wantProg, tt.wantArgs, tt.wantOK)
			}
		})
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestStartPipeNoop(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	var buf bytes.Buffer
	for _, w := range []io.Writer{w, &buf} {
		got, stop, err := Start(w)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		stop()
		_ = got
	}
}

func TestSpawnCat(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("cat/env helpers are unix-only; same code path runs on windows")
	}

	cat, err := exec.LookPath("cat")
	if err != nil {
		t.Skipf("no cat: %v", err)
	}

	outPath, out := createOut(t)
	feed, stop, err := spawn(cat, nil, out, []string{"PAGER_TEST=1"})
	if err != nil {
		t.Fatalf("spawn() error = %v", err)
	}
	if _, err := feed.Write([]byte("hello\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	stop()

	if got := readOut(t, outPath); !strings.Contains(got, "hello") {
		t.Fatalf("spawned output %q misses feed", got)
	}
}

func TestSpawnEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("env helper is unix-only; same code path runs on windows")
	}

	envBin, err := exec.LookPath("env")
	if err != nil {
		t.Skipf("no env: %v", err)
	}

	t.Run("less default", func(t *testing.T) {
		outPath, out := createOut(t)
		_, stop, err := spawn(envBin, nil, out, []string{"PAGER_TEST=1"})
		if err != nil {
			t.Fatalf("spawn() error = %v", err)
		}
		stop()

		if got := readOut(t, outPath); !strings.Contains(got, "LESS=FRX") {
			t.Fatalf("spawned env %q misses default LESS", got)
		}
	})

	t.Run("less respected", func(t *testing.T) {
		outPath, out := createOut(t)
		_, stop, err := spawn(envBin, nil, out, []string{"LESS=Rx"})
		if err != nil {
			t.Fatalf("spawn() error = %v", err)
		}
		stop()

		got := readOut(t, outPath)
		if !strings.Contains(got, "LESS=Rx") {
			t.Fatalf("spawned env %q misses preset LESS", got)
		}
		if strings.Contains(got, "LESS=FRX") {
			t.Fatalf("spawned env %q overrides preset LESS", got)
		}
	})
}

// createOut makes a temp output file for spawned processes.
func createOut(t *testing.T) (string, *os.File) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "out.txt")
	// #nosec G304 -- test-owned temp path.
	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Cleanup(func() { _ = out.Close() })
	return path, out
}

func readOut(t *testing.T, path string) string {
	t.Helper()

	// #nosec G304 -- test-owned temp path, written by createOut.
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	return string(content)
}
