package tests

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const modulePath = "github.com/CHE3MZ/terminull/v2"

var libraryDeps = map[string]map[string]bool{
	"termcore":     {},
	"termstyle":    {"termcore": true},
	"termbox":      {"termcore": true, "termstyle": true},
	"termtable":    {},
	"termimage":    {"termcore": true, "termstyle": true},
	"termoutput":   {"termcore": true, "termstyle": true, "termtable": true},
	"termprogress": {"termcore": true, "termstyle": true},
	"termprompt":   {"termcore": true, "termstyle": true},
	"termrender":   {},
	"termpager":    {"termcore": true},
	"termstatic":   {"termstyle": true},
	"termretry":    {},
	"termconfig":   {},
	"termapp": {
		"termconfig": true,
		"termcore":   true,
		"termoutput": true,
		"termrender": true,
	},
}

var directExternal = map[string][]string{
	"termcore":   {"golang.org/x/sys", "golang.org/x/term"},
	"termconfig": {"gopkg.in/yaml.v3"},
	"termoutput": {"gopkg.in/yaml.v3"},
}

var externalModules = map[string]bool{
	modulePath:          true,
	"golang.org/x/sys":  true,
	"golang.org/x/term": true,
	"gopkg.in/yaml.v3":  true,
}

func repoRoot(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("Abs(..) error = %v", err)
	}
	return root
}

func runGo(t *testing.T, dir string, args ...string) []byte {
	t.Helper()

	// #nosec G204 -- fixed go command with test-controlled package names.
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, exit.Stderr)
		}
		t.Fatalf("go %s: %v", strings.Join(args, " "), err)
	}
	return out
}

func topLevel(importPath string) string {
	trimmed := strings.TrimPrefix(importPath, modulePath+"/")
	return strings.SplitN(trimmed, "/", 2)[0]
}

func allowedExternal(importPath string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") {
			return true
		}
	}
	return false
}

func TestDependencyLaw(t *testing.T) {
	root := repoRoot(t)

	for name, allowed := range libraryDeps {
		var pkg struct {
			Imports []string
		}
		if err := json.Unmarshal(runGo(t, root, "list", "-json", "./"+name), &pkg); err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}

		for _, imp := range pkg.Imports {
			if strings.HasPrefix(imp, modulePath+"/") {
				dep := topLevel(imp)
				if dep != name && !allowed[dep] {
					t.Errorf("%s imports %s: upward or unlisted edge", name, dep)
				}
				continue
			}
			first := strings.SplitN(imp, "/", 2)[0]
			if strings.Contains(first, ".") && !allowedExternal(imp, directExternal[name]) {
				t.Errorf("%s imports unexpected external %s", name, imp)
			}
		}
	}
}

func TestExternalModuleSurface(t *testing.T) {
	root := repoRoot(t)

	for name := range libraryDeps {
		out := string(runGo(t, root, "list", "-deps", "-f", "{{with .Module}}{{.Path}}{{end}}", "./"+name))
		for _, mod := range strings.Fields(out) {
			if !externalModules[mod] {
				t.Errorf("%s depends on unexpected module %s", name, mod)
			}
		}
	}
}

func TestModuleRequires(t *testing.T) {
	root := repoRoot(t)

	var mod struct {
		Require []struct {
			Path     string
			Indirect bool
		}
	}
	if err := json.Unmarshal(runGo(t, root, "mod", "edit", "-json"), &mod); err != nil {
		t.Fatalf("decode go.mod: %v", err)
	}

	wantDirect := map[string]bool{
		"golang.org/x/sys":  true,
		"golang.org/x/term": true,
		"gopkg.in/yaml.v3":  true,
	}
	allowedIndirect := map[string]bool{
		"github.com/kr/pretty": true,
		"github.com/kr/text":   true,
		"gopkg.in/check.v1":    true,
	}

	gotDirect := make(map[string]bool)
	for _, req := range mod.Require {
		if !req.Indirect {
			gotDirect[req.Path] = true
		} else if !allowedIndirect[req.Path] {
			t.Errorf("unexpected indirect requirement %s", req.Path)
		}
	}

	for path := range wantDirect {
		if !gotDirect[path] {
			t.Errorf("missing direct requirement %s", path)
		}
		delete(gotDirect, path)
	}
	for path := range gotDirect {
		t.Errorf("unexpected direct requirement %s", path)
	}
}
