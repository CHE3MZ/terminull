package termconfig

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func TestPrecedence(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "custom.yaml")
	writeFile(t, cfgFile, "b: 2\nc: 2\nd: 2\n")

	cfg, err := Load(Options{
		AppName:    "pp1",
		ConfigFile: cfgFile,
		Defaults:   map[string]any{"a": 1, "b": 1, "c": 1, "d": 1},
		Environ:    []string{"PP1_C=3", "PP1_D=3"},
		Overrides:  map[string]any{"d": 4},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	for path, want := range map[string]int{"a": 1, "b": 2, "c": 3, "d": 4} {
		if got := cfg.Int(path); got != want {
			t.Errorf("Int(%q) = %d, want %d", path, got, want)
		}
	}
}

func TestMissingFilesOK(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("APPDATA", tmp)

	cfg, err := Load(Options{
		AppName:  "missingapp",
		Defaults: map[string]any{"x": 9},
		Environ:  []string{},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got := cfg.Int("x"); got != 9 {
		t.Fatalf("Int(x) = %d, want 9", got)
	}
	if cfg.Exists("nope") || cfg.Get("nope") != nil {
		t.Fatal("absent keys should report Exists=false, Get=nil")
	}
}

func TestExplicitMissingFileErrors(t *testing.T) {
	t.Parallel()

	_, err := Load(Options{
		AppName:    "pp3",
		ConfigFile: filepath.Join(t.TempDir(), "nope.yaml"),
	})
	if err == nil || !strings.Contains(err.Error(), "nope.yaml") {
		t.Fatalf("Load() error = %v, want missing path named", err)
	}
}

func TestMalformedYAMLErrors(t *testing.T) {
	t.Parallel()

	bad := filepath.Join(t.TempDir(), "bad.yaml")
	writeFile(t, bad, "foo: [unclosed\n")

	_, err := Load(Options{AppName: "pp4", ConfigFile: bad})
	if err == nil {
		t.Fatal("Load() = nil, want YAML parse error")
	}
}

func TestStrictRejectsUnknownKeys(t *testing.T) {
	t.Parallel()

	cfg, err := Load(Options{
		AppName: "pp5",
		Environ: []string{"PP5_KNOWN=hi", "PP5_BOGUS=oops"},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	var target struct {
		Known string `koanf:"known"`
	}

	if err := cfg.Unmarshal(&target); err != nil {
		t.Fatalf("Unmarshal() error = %v, want nil for lenient decode", err)
	}
	if target.Known != "hi" {
		t.Fatalf("Known = %q, want hi", target.Known)
	}

	err = cfg.UnmarshalStrict(&target)
	if err == nil || !strings.Contains(err.Error(), "bogus") {
		t.Fatalf("UnmarshalStrict() error = %v, want unknown key named", err)
	}
}

func TestTypedEnvValues(t *testing.T) {
	t.Parallel()

	cfg, err := Load(Options{
		AppName: "pp6",
		Environ: []string{
			"PP6_FLAG=true",
			"PP6_COUNT=42",
			"PP6_ONE=1",
			"PP6_RATIO=1.5",
			"PP6_NAME=hi",
			"PP6_ZIP=007",
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Bool("flag") {
		t.Fatal("Bool(flag) = false, want true")
	}
	if got := cfg.Int("count"); got != 42 {
		t.Fatalf("Int(count) = %d, want 42", got)
	}
	if got := cfg.Int("one"); got != 1 {
		t.Fatalf("Int(one) = %d, want 1", got)
	}
	if got := cfg.Float64("ratio"); got != 1.5 {
		t.Fatalf("Float64(ratio) = %v, want 1.5", got)
	}
	if got := cfg.String("name"); got != "hi" {
		t.Fatalf("String(name) = %q, want hi", got)
	}
	if got := cfg.String("zip"); got != "007" {
		t.Fatalf("String(zip) = %q, want leading zeros kept", got)
	}
}

// TestSearchFindsUserConfigDir points every platform home at tmp and
// asserts the UserConfigDir-resolved file is picked up.
func TestSearchFindsUserConfigDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("APPDATA", tmp)

	dir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("UserConfigDir() error = %v", err)
	}
	writeFile(t, filepath.Join(dir, "sapp", "sapp.yaml"), "found: xdg\n")

	cfg, err := Load(Options{AppName: "sapp", Environ: []string{}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := cfg.String("found"); got != "xdg" {
		t.Fatalf("String(found) = %q, want xdg", got)
	}
}

// TestHomeDotfileLayeredUnderPlatformDir asserts both files layer with
// the platform dir winning on shared keys.
func TestHomeDotfileLayeredUnderPlatformDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("APPDATA", tmp)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("UserConfigDir() error = %v", err)
	}

	// Skip when both resolve identically (degenerate env).
	homeFile := filepath.Join(home, ".happ.yaml")
	platFile := filepath.Join(dir, "happ", "happ.yaml")
	if homeFile == platFile {
		t.Skip("home and config dir coincide here")
	}

	writeFile(t, homeFile, "found: home\nonlyhome: 1\n")
	writeFile(t, platFile, "found: plat\n")

	cfg, err := Load(Options{AppName: "happ", Environ: []string{}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := cfg.String("found"); got != "plat" {
		t.Fatalf("String(found) = %q, want plat to win", got)
	}
	if got := cfg.Int("onlyhome"); got != 1 {
		t.Fatalf("Int(onlyhome) = %d, want home file layered too", got)
	}
}

func TestWatchReloadsOnChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yaml")
	writeFile(t, path, "value: one\n")

	cfg, err := Load(Options{ConfigFile: path, Environ: []string{}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := cfg.String("value"); got != "one" {
		t.Fatalf("String(value) = %q, want one", got)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changes := make(chan *Config, 1)
	errs := make(chan error, 1)

	go func() {
		errs <- cfg.Watch(ctx, 10*time.Millisecond, func(next *Config, err error) {
			if err != nil {
				t.Errorf("Watch callback error = %v", err)
				return
			}
			changes <- next
		})
	}()

	var next *Config
	deadline := time.After(3 * time.Second)
	for i := 0; next == nil; i++ {
		value := fmt.Sprintf("value: v%d\n", i)
		writeFile(t, path, value)
		select {
		case next = <-changes:
		case <-time.After(50 * time.Millisecond):
		case <-deadline:
			t.Fatal("timed out waiting for config reload")
		}
	}
	if got := next.String("value"); !strings.HasPrefix(got, "v") {
		t.Fatalf("reloaded String(value) = %q, want latest write", got)
	}

	cancel()
	select {
	case err := <-errs:
		if err != context.Canceled {
			t.Fatalf("Watch() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Watch() did not return after cancel")
	}
}

func TestWatchRequiresCallback(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	if err := cfg.Watch(context.Background(), time.Second, nil); err == nil {
		t.Fatal("Watch() with nil callback returned nil")
	}
}

func TestTOMLFileAndPrecedence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	writeFile(t, path, "name = \"file\"\nport = 1\nnested = { enabled = true }\n")

	cfg, err := Load(Options{
		ConfigFile: path,
		Defaults:   map[string]any{"name": "default", "port": 0},
		Environ:    []string{"APP_PORT=2"},
		Overrides:  map[string]any{"port": 3},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := cfg.String("name"); got != "file" {
		t.Fatalf("String(name) = %q, want file", got)
	}
	if got := cfg.Int("port"); got != 3 {
		t.Fatalf("Int(port) = %d, want override 3", got)
	}
	if !cfg.Bool("nested.enabled") {
		t.Fatal("Bool(nested.enabled) = false, want true")
	}
}

func TestTOMLFormatOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.conf")
	writeFile(t, path, "name = \"demo\"\n")

	cfg, err := Load(Options{ConfigFile: path, Format: "toml", Environ: []string{}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := cfg.String("name"); got != "demo" {
		t.Fatalf("String(name) = %q, want demo", got)
	}
}
