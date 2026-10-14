package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/CHE3MZ/terminull/config"
	"github.com/CHE3MZ/terminull/core"
	"github.com/CHE3MZ/terminull/render"
	"github.com/CHE3MZ/terminull/render/renderers"
	"github.com/CHE3MZ/terminull/render/terminal"
)

// App owns the root cobra command. Build one per process with New,
// register commands with Add, run with Execute.
type App struct {
	name    string
	version string
	short   string
	long    string
	cmd     *cobra.Command
	ran     bool

	configFile string
	output     string
	noColor    bool
	verbose    bool
	debug      bool
}

// Option tunes an App at construction.
type Option func(*App)

// WithVersion sets the --version string (default core.FullVersion()).
func WithVersion(version string) Option {
	return func(a *App) {
		a.version = version
	}
}

// WithShort sets the one-line description shown atop help.
func WithShort(short string) Option {
	return func(a *App) {
		a.short = short
	}
}

// WithLong sets the full description shown atop help.
func WithLong(long string) Option {
	return func(a *App) {
		a.long = long
	}
}

// New builds an App with the five global flags:
//
//	--config/-f, --output/-o table|json|yaml, --no-color, --verbose, --debug
func New(name string, opts ...Option) *App {
	a := &App{
		name:    name,
		version: core.FullVersion(),
		output:  "table",
	}
	for _, opt := range opts {
		opt(a)
	}

	cmd := &cobra.Command{
		Use:           name,
		Short:         a.short,
		Long:          a.long,
		Version:       a.version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return a.preRun(cmd, args)
		},
	}

	cmd.PersistentFlags().StringVarP(&a.configFile, "config", "f", "", "config file (default search: home, user config dir, cwd)")
	cmd.PersistentFlags().StringVarP(&a.output, "output", "o", "table", "output format: table|json|yaml")
	cmd.PersistentFlags().BoolVar(&a.noColor, "no-color", false, "disable color output")
	cmd.PersistentFlags().BoolVar(&a.verbose, "verbose", false, "verbose logging")
	cmd.PersistentFlags().BoolVar(&a.debug, "debug", false, "debug logging")

	a.cmd = cmd
	return a
}

// Add registers a top-level command (which may carry Children groups).
func (a *App) Add(c Command) {
	a.cmd.AddCommand(a.build(c))
}

// SetOut redirects standard output (tests capture help and results here).
func (a *App) SetOut(w io.Writer) {
	a.cmd.SetOut(w)
}

// SetErr redirects standard error.
func (a *App) SetErr(w io.Writer) {
	a.cmd.SetErr(w)
}

// SetIn redirects standard input (prompts read here).
func (a *App) SetIn(r io.Reader) {
	a.cmd.SetIn(r)
}

// Raw escapes to the underlying cobra command for advanced use.
// Prefer Add/Execute; Raw exists so the wrapper never blocks you.
func (a *App) Raw() *cobra.Command {
	return a.cmd
}

// Execute runs args against a signal-scoped ctx (see core.SignalContext).
// Parse and setup failures return ExitCodeError(2); handler errors pass
// through for core.ExitCodeOf mapping by main.
func (a *App) Execute(ctx context.Context, args []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	a.cmd.SetArgs(args)
	a.ran = false

	if err := a.cmd.ExecuteContext(ctx); err != nil {
		var exitErr core.ExitCodeError
		if !errors.As(err, &exitErr) && !errors.Is(err, context.Canceled) && !a.ran {
			return core.NewExitCodeError(2, err)
		}
		return err
	}
	return nil
}

// preRun boots, in order: debug switch, output validation, renderer +
// logger, layered config. The *Context travels via cmd.Context().
func (a *App) preRun(cmd *cobra.Command, args []string) error {
	core.SetDebug(a.debug || a.verbose)

	if a.output != "table" && a.output != "json" && a.output != "yaml" {
		return fmt.Errorf("invalid --output %q: want table|json|yaml", a.output)
	}

	level := render.InfoLevel
	if a.verbose || a.debug {
		level = render.DebugLevel
	}

	renderer, stop := a.selectRenderer(cmd)
	logger := render.NewLogger(level, renderer)

	cfg, err := config.Load(config.Options{
		AppName:    a.name,
		ConfigFile: a.configFile,
		Overrides:  collectOverrides(cmd),
	})
	if err != nil {
		stop()
		return err
	}

	tctx := &Context{
		Context: cmd.Context(),
		cmd:     cmd,
		args:    args,
		stop:    stop,
		Config:  cfg,
		Logger:  logger,
		Out:     cmd.OutOrStdout(),
		Err:     cmd.ErrOrStderr(),
		In:      cmd.InOrStdin(),
		Output:  a.output,
		NoColor: a.noColor,
		Verbose: a.verbose || a.debug,
	}
	cmd.SetContext(context.WithValue(cmd.Context(), contextKey{}, tctx))
	return nil
}

// selectRenderer uses the auto renderer for real terminal files and a
// plain simple renderer for captured writers (tests, pipes without files),
// so static output stays assertion-friendly.
func (a *App) selectRenderer(cmd *cobra.Command) (render.LogRendered, func()) {
	if outFile, ok := cmd.OutOrStdout().(*os.File); ok {
		return renderers.AutoRenderer(outFile, a.noColor)
	}
	renderer := renderers.NewSimpleRenderer(cmd.OutOrStdout(), terminal.NoColorSchema())
	return renderer, func() {}
}

// build converts a Command, recursing into Children groups. Drawing starts
// inside RunE (not preRun) so --help/--version/completion never draw.
func (a *App) build(c Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:           c.Use,
		Short:         c.Short,
		Long:          c.Long,
		Example:       c.Example,
		Args:          c.Args,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			tctx, ok := cmd.Context().Value(contextKey{}).(*Context)
			if !ok || tctx == nil {
				return fmt.Errorf("terminull: missing context (was PersistentPreRunE skipped?)")
			}
			// Terminal setup runs synchronously so prompts started by
			// Run never race console configuration; only the idle draw
			// loop runs in the background.
			if starter, ok := tctx.Logger.Renderer().(interface {
				Setup()
				StartDrawing()
			}); ok {
				starter.Setup()
				go starter.StartDrawing()
			}
			defer tctx.stop()
			a.ran = true
			if c.Run == nil {
				return cmd.Help()
			}
			return c.Run(tctx)
		},
	}

	if c.Flags != nil {
		c.Flags(cmd.Flags())
	}
	for _, child := range c.Children {
		cmd.AddCommand(a.build(child))
	}
	return cmd
}

// collectOverrides gathers explicitly-changed flags with types, so CLI
// flags win over env and files in the config layer.
func collectOverrides(cmd *cobra.Command) map[string]any {
	overrides := make(map[string]any)

	collect := func(fs *pflag.FlagSet) {
		fs.Visit(func(f *pflag.Flag) {
			switch f.Value.Type() {
			case "bool":
				if v, err := fs.GetBool(f.Name); err == nil {
					overrides[f.Name] = v
				}
			case "int":
				if v, err := fs.GetInt(f.Name); err == nil {
					overrides[f.Name] = v
				}
			case "stringSlice", "stringArray":
				if v, err := fs.GetStringSlice(f.Name); err == nil {
					overrides[f.Name] = v
				}
			default:
				overrides[f.Name] = f.Value.String()
			}
		})
	}

	collect(cmd.Flags())
	collect(cmd.PersistentFlags())
	collect(cmd.InheritedFlags())
	return overrides
}
