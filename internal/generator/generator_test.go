package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	gomod := readGenerated(t, dir, "go.mod")
	for _, want := range []string{
		"module example.com/app",
		"require github.com/CHE3MZ/terminull/v2 v9.9.9",
		"replace github.com/CHE3MZ/terminull/v2 => \"C:/fake/path\"",
	} {
		if !strings.Contains(gomod, want) {
			t.Fatalf("go.mod %q does not contain %q", gomod, want)
		}
	}

	main := readGenerated(t, dir, "cmd", "app", "main.go")
	for _, want := range []string{`commands.NewApp()`, `"example.com/app/internal/commands"`} {
		if !strings.Contains(main, want) {
			t.Fatalf("main.go does not contain %q", want)
		}
	}
}

func TestDefaultTerminullVersionMatchesModuleMajor(t *testing.T) {
	t.Parallel()

	// #nosec G304 -- test reads the repository go.mod.
	gomod, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	module := strings.TrimPrefix(strings.SplitN(string(gomod), "\n", 2)[0], "module ")
	major := module[strings.LastIndex(module, "/")+1:]
	if !strings.HasPrefix(DefaultTerminullVersion, major+".") {
		t.Fatalf("DefaultTerminullVersion = %q, want %s.x for module %s", DefaultTerminullVersion, major, module)
	}
}

func TestGenerateDefaults(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "hello")
	if _, err := Generate(Options{Name: "hello", Dir: dir}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	gomod := readGenerated(t, dir, "go.mod")
	if !strings.Contains(gomod, "module hello") {
		t.Fatalf("go.mod %q defaults module to name", gomod)
	}
	if !strings.Contains(gomod, "require github.com/CHE3MZ/terminull/v2 "+DefaultTerminullVersion) {
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
