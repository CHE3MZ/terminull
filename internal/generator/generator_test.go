package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateHappyPath(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "app")
	files, err := Generate(Options{
		Name:             "app",
		Module:           "example.com/app",
		Dir:              dir,
		TerminullVersion: "v9.9.9",
		Replace:          `C:\fake\path`,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(files) != 4 {
		t.Fatalf("files = %v, want 4 created", files)
	}

	for _, path := range []string{
		"go.mod",
		filepath.Join("cmd", "app", "main.go"),
		filepath.Join("internal", "commands", "commands.go"),
		"README.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
			t.Fatalf("Stat(%s) error = %v", path, err)
		}
	}

	gomod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	for _, want := range []string{
		"module example.com/app",
		"require github.com/CHE3MZ/terminull v9.9.9",
		"replace github.com/CHE3MZ/terminull => C:/fake/path",
	} {
		if !strings.Contains(string(gomod), want) {
			t.Fatalf("go.mod %q does not contain %q", gomod, want)
		}
	}

	main, err := os.ReadFile(filepath.Join(dir, "cmd", "app", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	for _, want := range []string{`commands.NewApp()`, `"example.com/app/internal/commands"`} {
		if !strings.Contains(string(main), want) {
			t.Fatalf("main.go does not contain %q", want)
		}
	}
}

func TestGenerateDefaults(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "hello")
	if _, err := Generate(Options{Name: "hello", Dir: dir}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	gomod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	if !strings.Contains(string(gomod), "module hello") {
		t.Fatalf("go.mod %q defaults module to name", gomod)
	}
	if !strings.Contains(string(gomod), "require github.com/CHE3MZ/terminull "+DefaultTerminullVersion) {
		t.Fatalf("go.mod %q defaults framework version", gomod)
	}
}

func TestGenerateBadNames(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"", "has space", "-dash", ".dot", "nul", "COM1", "a/b"} {
		if _, err := Generate(Options{Name: name, Dir: filepath.Join(t.TempDir(), "x")}); err == nil {
			t.Fatalf("Generate(%q) = nil, want error", name)
		}
	}
}

func TestGenerateNonEmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := Generate(Options{Name: "app", Dir: dir}); err == nil {
		t.Fatal("Generate() into non-empty dir = nil, want error")
	}
}
