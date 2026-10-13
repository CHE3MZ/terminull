package generator

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// DefaultTerminullVersion pins new projects to the latest released
// framework. Override per project with Options.TerminullVersion.
const DefaultTerminullVersion = "v1.0.0"

// Options tunes a Generate call. Name is required; Module defaults to
// Name (a relative module path builds fine locally); Dir defaults to
// ./Name; TerminullVersion defaults to DefaultTerminullVersion.
type Options struct {
	// Name is the binary, directory, and app name, e.g. "hello".
	Name string
	// Module is the go.mod module path, e.g. "example.com/hello".
	Module string
	// Dir is the target directory, created when missing.
	Dir string
	// TerminullVersion is the framework require version, e.g. "v1.0.0".
	TerminullVersion string
	// Replace adds a replace directive (local dogfood), slashes normalized.
	Replace string
}

var validName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// reservedNames cannot be directories or binaries on Windows.
var reservedNames = map[string]bool{
	"con": true, "prn": true, "aux": true, "nul": true,
	"com1": true, "com2": true, "com3": true, "com4": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true,
}

type resolved struct {
	Name             string
	Module           string
	Dir              string
	TerminullVersion string
	Replace          string
}

func (opts Options) resolve() (resolved, error) {
	r := resolved{
		Name:             opts.Name,
		Module:           opts.Module,
		TerminullVersion: opts.TerminullVersion,
		Replace:          strings.ReplaceAll(opts.Replace, "\\", "/"),
	}

	if !validName.MatchString(r.Name) || reservedNames[strings.ToLower(r.Name)] {
		return resolved{}, fmt.Errorf("generator: invalid app name %q: use letters, digits, _ and -", opts.Name)
	}
	if r.Module == "" {
		r.Module = r.Name
	}
	if strings.ContainsAny(r.Module, " \t\n\r") {
		return resolved{}, fmt.Errorf("generator: invalid module path %q", opts.Module)
	}
	if r.TerminullVersion == "" {
		r.TerminullVersion = DefaultTerminullVersion
	}
	if opts.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return resolved{}, fmt.Errorf("generator: working directory: %w", err)
		}
		r.Dir = filepath.Join(cwd, r.Name)
	} else {
		r.Dir = opts.Dir
	}
	return r, nil
}

// Generate writes the scaffold into Dir and returns created file paths.
// An existing Dir must be empty; generated Go files are gofmt-formatted.
func Generate(opts Options) ([]string, error) {
	r, err := opts.resolve()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(r.Dir)
	if err == nil && len(entries) > 0 {
		return nil, fmt.Errorf("generator: %s exists and is not empty", r.Dir)
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("generator: read %s: %w", r.Dir, err)
	}

	files := map[string]string{
		"go.mod":                                goModTemplate,
		filepath.Join("cmd", r.Name, "main.go"): mainTemplate,
		filepath.Join("internal", "commands", "commands.go"): commandsTemplate,
		"README.md": readmeTemplate,
	}

	var created []string
	for path, tmpl := range files {
		content, err := render(tmpl, r)
		if err != nil {
			return nil, err
		}
		raw := []byte(content)
		if strings.HasSuffix(path, ".go") {
			raw, err = format.Source(raw)
			if err != nil {
				return nil, fmt.Errorf("generator: template %s produced invalid Go: %w", path, err)
			}
		}
		full := filepath.Join(r.Dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
			return nil, fmt.Errorf("generator: create dirs: %w", err)
		}
		if err := os.WriteFile(full, raw, 0o600); err != nil {
			return nil, fmt.Errorf("generator: write %s: %w", path, err)
		}
		created = append(created, path)
	}
	return created, nil
}

func render(tmpl string, r resolved) (string, error) {
	t, err := template.New("file").Delims("[[", "]]").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("generator: parse template: %w", err)
	}
	var b strings.Builder
	if err := t.Execute(&b, r); err != nil {
		return "", fmt.Errorf("generator: render template: %w", err)
	}
	return b.String(), nil
}
