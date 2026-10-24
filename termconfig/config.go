package termconfig

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/CHE3MZ/terminull/v2/termconfig/internal/koanf"
	"github.com/CHE3MZ/terminull/v2/termconfig/internal/koanf/confmap"
	"github.com/CHE3MZ/terminull/v2/termconfig/internal/koanf/env"
	"github.com/CHE3MZ/terminull/v2/termconfig/internal/koanf/file"
	"github.com/CHE3MZ/terminull/v2/termconfig/internal/koanf/toml"
	"github.com/CHE3MZ/terminull/v2/termconfig/internal/koanf/yaml"
	"github.com/CHE3MZ/terminull/v2/termconfig/internal/mapstructure"
)

// Options tunes a Load call. Zero value loads nothing but empty defaults.
type Options struct {
	// AppName drives file names and the env prefix: "deploy" looks for
	// deploy.yaml and reads DEPLOY_ variables.
	AppName string
	// ConfigFile is an explicit --config path. Missing file is an error.
	ConfigFile string
	// Format overrides file format detection ("yaml" or "toml"). Empty
	// detects from ConfigFile's extension, defaulting to YAML.
	Format string
	// Defaults is the lowest-precedence layer.
	Defaults map[string]any
	// Overrides is the highest-precedence layer (parsed CLI flags).
	Overrides map[string]any
	// EnvPrefix overrides the derived UPPER(AppName)+"_" prefix.
	EnvPrefix string
	// Environ injects environment lines for tests; nil uses os.Environ.
	Environ []string
	// EnvironFunc injects an environ function; takes precedence over Environ.
	EnvironFunc func() []string
}

// Config wraps the layered koanf instance.
type Config struct {
	ko   *koanf.Koanf
	opts Options
}

// Load layers defaults, files, env, and overrides in that order.
func Load(opts Options) (*Config, error) {
	ko := koanf.New(".")

	if len(opts.Defaults) > 0 {
		if err := ko.Load(confmap.Provider(opts.Defaults, "."), nil); err != nil {
			return nil, fmt.Errorf("load defaults: %w", err)
		}
	}

	files, err := opts.configFiles()
	if err != nil {
		return nil, err
	}
	for _, path := range files {
		if err := ko.Load(file.Provider(path), opts.parserFor(path)); err != nil {
			return nil, fmt.Errorf("load config file %s: %w", path, err)
		}
	}

	envOpt := env.Opt{
		Prefix:        opts.envPrefix(),
		TransformFunc: transformFunc(opts.envPrefix()),
	}
	if opts.EnvironFunc != nil {
		envOpt.EnvironFunc = opts.EnvironFunc
	} else if opts.Environ != nil {
		environ := opts.Environ
		envOpt.EnvironFunc = func() []string { return environ }
	}
	if err := ko.Load(env.Provider(".", envOpt), nil); err != nil {
		return nil, fmt.Errorf("load environment: %w", err)
	}

	if len(opts.Overrides) > 0 {
		if err := ko.Load(confmap.Provider(opts.Overrides, "."), nil); err != nil {
			return nil, fmt.Errorf("load overrides: %w", err)
		}
	}

	return &Config{ko: ko, opts: opts}, nil
}

// Watch polls the config file search path and calls onChange with a
// freshly loaded Config whenever a candidate's size or modification time
// changes. It blocks until ctx is canceled. Call onChange only with the
// returned Config; the receiver is not mutated in place.
func (c *Config) Watch(ctx context.Context, interval time.Duration, onChange func(*Config, error)) error {
	if onChange == nil {
		return fmt.Errorf("watch: onChange is required")
	}
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	paths := c.opts.watchFiles()
	if len(paths) == 0 {
		<-ctx.Done()
		return ctx.Err()
	}

	last := fileFingerprint(paths)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			next := fileFingerprint(paths)
			if next == last {
				continue
			}
			last = next
			cfg, err := Load(c.opts)
			if err != nil {
				onChange(nil, err)
				continue
			}
			onChange(cfg, nil)
		}
	}
}

func fileFingerprint(paths []string) string {
	var b strings.Builder
	for _, path := range paths {
		b.WriteString(path)
		b.WriteByte('|')
		info, err := os.Stat(path)
		if err != nil {
			b.WriteString("missing")
		} else {
			b.WriteString(strconv.FormatInt(info.ModTime().UnixNano(), 10))
			b.WriteByte(':')
			b.WriteString(strconv.FormatInt(info.Size(), 10))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func (opts Options) parserFor(path string) koanf.Parser {
	format := strings.ToLower(strings.TrimPrefix(opts.Format, "."))
	if format == "" {
		format = strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	}
	if format == "toml" {
		return toml.Parser()
	}
	return yaml.Parser()
}

func (opts Options) envPrefix() string {
	if opts.EnvPrefix != "" {
		return opts.EnvPrefix
	}
	return strings.ToUpper(opts.AppName) + "_"
}

// configFiles resolves the file layer: the explicit path alone, or every
// existing search candidate in low-to-high precedence order.
func (opts Options) configFiles() ([]string, error) {
	if opts.ConfigFile != "" {
		if _, err := os.Stat(opts.ConfigFile); err != nil {
			return nil, fmt.Errorf("config file %s: %w", opts.ConfigFile, err)
		}
		return []string{opts.ConfigFile}, nil
	}

	if opts.AppName == "" {
		return nil, nil
	}

	var found []string
	for _, path := range opts.candidateFiles() {
		if _, err := os.Stat(path); err == nil {
			found = append(found, path)
		}
	}
	return found, nil
}

// candidateFiles is every search path, whether or not it exists, in
// low-to-high precedence order.
func (opts Options) candidateFiles() []string {
	var candidates []string

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		candidates = append(candidates,
			filepath.Join(home, "."+opts.AppName+".yaml"),
			filepath.Join(home, "."+opts.AppName+".toml"),
		)
	}
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		candidates = append(candidates,
			filepath.Join(dir, opts.AppName, opts.AppName+".yaml"),
			filepath.Join(dir, opts.AppName, opts.AppName+".toml"),
		)
	}
	if cwd, err := os.Getwd(); err == nil && cwd != "" {
		candidates = append(candidates,
			filepath.Join(cwd, opts.AppName+".yaml"),
			filepath.Join(cwd, "."+opts.AppName+".yaml"),
			filepath.Join(cwd, opts.AppName+".toml"),
			filepath.Join(cwd, "."+opts.AppName+".toml"),
		)
	}
	return candidates
}

// watchFiles includes missing search candidates so creating a new
// higher-precedence file triggers a reload.
func (opts Options) watchFiles() []string {
	if opts.ConfigFile != "" {
		return []string{opts.ConfigFile}
	}
	if opts.AppName == "" {
		return nil
	}
	return opts.candidateFiles()
}

// transformFunc maps APPNAME_DB_HOST to db.host and types scalar values
// so env strings merge cleanly with typed YAML values.
func transformFunc(prefix string) func(string, string) (string, any) {
	return func(k, v string) (string, any) {
		key := strings.ToLower(strings.TrimPrefix(k, prefix))
		key = strings.ReplaceAll(key, "_", ".")

		if key == "" {
			return "", nil
		}
		return key, typedValue(v)
	}
}

// typedValue converts numeric-looking and boolean env values, leaving
// everything else a string. Leading-zero numbers stay strings ("007").
func typedValue(v string) any {
	if i, err := strconv.ParseInt(v, 10, 64); err == nil {
		if len(v) > 1 && strings.HasPrefix(v, "0") && !strings.HasPrefix(v, "-0") {
			return v
		}
		return i
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}
	return v
}

// Get returns the value at the dotted path, or nil when absent.
func (c *Config) Get(path string) any {
	return c.ko.Get(path)
}

// Exists reports whether the dotted path is set.
func (c *Config) Exists(path string) bool {
	return c.ko.Exists(path)
}

// String returns the value at path as a string, "" when absent.
func (c *Config) String(path string) string {
	return c.ko.String(path)
}

// Bool returns the value at path as a bool, false when absent.
func (c *Config) Bool(path string) bool {
	return c.ko.Bool(path)
}

// Int returns the value at path as an int, 0 when absent.
func (c *Config) Int(path string) int {
	return c.ko.Int(path)
}

// Float64 returns the value at path as a float64, 0 when absent.
func (c *Config) Float64(path string) float64 {
	return c.ko.Float64(path)
}

// Strings returns the value at path as a string slice.
func (c *Config) Strings(path string) []string {
	return c.ko.Strings(path)
}

// All returns every layered key for dumps and debugging.
func (c *Config) All() map[string]any {
	return c.ko.All()
}

// Unmarshal decodes into target with weak typing (env strings convert).
// Unknown keys are ignored; use UnmarshalStrict to reject them.
func (c *Config) Unmarshal(target any) error {
	return c.ko.UnmarshalWithConf("", target, koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			TagName:          "koanf",
			WeaklyTypedInput: true,
		},
	})
}

// UnmarshalStrict decodes like Unmarshal but fails naming any key that
// has no matching struct field, catching typos in files and env.
func (c *Config) UnmarshalStrict(target any) error {
	return c.ko.UnmarshalWithConf("", target, koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			TagName:          "koanf",
			WeaklyTypedInput: true,
			ErrorUnused:      true,
		},
	})
}
