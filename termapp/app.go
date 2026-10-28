package termapp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/CHE3MZ/terminull/v2/termapp/internal/cli"
	"github.com/CHE3MZ/terminull/v2/termapp/internal/cli/pflag"
	"github.com/CHE3MZ/terminull/v2/termconfig"
	"github.com/CHE3MZ/terminull/v2/termcore"
	"github.com/CHE3MZ/terminull/v2/termoutput"
	"github.com/CHE3MZ/terminull/v2/termrender"
	"github.com/CHE3MZ/terminull/v2/termrender/renderers"
	renderconfig "github.com/CHE3MZ/terminull/v2/termrender/renderers/config"
	"github.com/CHE3MZ/terminull/v2/termrender/terminal"
)

// App owns the root CLI command. Build one per process with New,
// register commands with Add, run with Execute.
type App struct {
	name    string
	version string
	short   string
	long    string
	cmd     *cli.Command
	ran     bool

	configFile    string
	output        string
	noColor       bool
	verbose       bool
	debug         bool
	logLevel      termrender.LogLevel
	logLevelValue string
	logStyle      termrender.LogStyle
	logStyleValue string
	styleColors   *terminal.FlatStyleColors
	logOutput     io.Writer
	logRenderer   LogRendererFactory
	// symbolsOnly forces ASCII spinners and statuses even where emoji
	// would render: deterministic output for logs, scripts, and house
	// styles that prefer it.
	symbolsOnly bool
}

// LogRendererOptions describes the logging destination and built-in
// style choices available to a LogRendererFactory.
type LogRendererOptions struct {
	// Out receives log output. It defaults to the command's stderr.
	Out io.Writer
	// File is Out when it is an *os.File, otherwise nil.
	File *os.File
	// Style is the selected flat log style.
	Style termrender.LogStyle
	// NoColor mirrors --no-color. Custom renderers must also honor
	// NO_COLOR, TERM=dumb, CI, and non-terminal writers when needed.
	NoColor bool
	// Colors is the flat prefix palette. Nil means plain prefix tokens.
	Colors *terminal.FlatStyleColors
	// SymbolsOnly mirrors WithSymbolsOnly.
	SymbolsOnly bool
}

// LogRendererFactory builds a custom log renderer and its stop function.
// Returning a nil renderer is an error. A nil stop is treated as a no-op.
type LogRendererFactory func(LogRendererOptions) (termrender.LogRendered, func())

// Option tunes an App at construction.
type Option func(*App)

// WithVersion sets the --version string (default termcore.FullVersion()).
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

// WithSymbolsOnly forces ASCII spinners and statuses on terminals that
// could render emoji. Use it for deterministic output and house styles
// that prefer the classic look; default follows terminal capability.
func WithSymbolsOnly() Option {
	return func(a *App) {
		a.symbolsOnly = true
	}
}

// WithLogStyle sets the default flat log style. Any non-tree style
// forces flat lines on terminals; --log-style overrides it.
func WithLogStyle(style termrender.LogStyle) Option {
	return func(a *App) {
		a.logStyle = style
	}
}

// WithLogStyleColors overrides flat prefix colors. Start from
// terminal.DefaultFlatStyleColors and override fields:
//
//	colors := terminal.DefaultFlatStyleColors()
//	colors.WarnColor = terminal.BlueColor
//	app := termapp.New("app", termapp.WithLogStyleColors(colors))
//
// Pass nil to leave level, timestamp, and elapsed tokens plain.
// --no-color disables all ANSI regardless of this option.
func WithLogStyleColors(colors *terminal.FlatStyleColors) Option {
	return func(a *App) {
		a.styleColors = colors
	}
}

// WithLogLevel sets the default log level threshold. --log-level overrides
// it; --verbose and --debug still raise the level to debug unless an
// explicit --log-level is present.
func WithLogLevel(level termrender.LogLevel) Option {
	return func(a *App) {
		a.logLevel = level
	}
}

// WithLogOutput redirects logs away from the default stderr. Use it for
// a file or a test buffer; stdout stays reserved for command data.
func WithLogOutput(w io.Writer) Option {
	return func(a *App) {
		a.logOutput = w
	}
}

// WithLogRendererFactory replaces built-in renderer selection. The
// factory can build a custom style or wrap a flat renderer in CI folds:
//
//	termapp.WithLogRendererFactory(func(opts termapp.LogRendererOptions) (termrender.LogRendered, func()) {
//		simple := renderers.NewSimpleRenderer(opts.Out, terminal.NoColorSchema())
//		return renderers.NewGithubActionsLogsRenderer(simple), func() {}
//	})
//
// A nil factory keeps the built-in auto selection.
func WithLogRendererFactory(factory LogRendererFactory) Option {
	return func(a *App) {
		a.logRenderer = factory
	}
}

// New builds an App with the seven global flags (no shorthands, so
// subcommands own -f -o -e -v freely):
//
//	--config, --output table|json|yaml, --no-color, --verbose, --debug,
//	--log-level error|warn|info|debug|trace,
//	--log-style tree|plain|time|time-level|logrus|k8s
func New(name string, opts ...Option) *App {
	a := &App{
		name:        name,
		version:     termcore.FullVersion(),
		output:      termoutput.FormatTable,
		logLevel:    termrender.InfoLevel,
		logStyle:    termrender.LogStyleTree,
		styleColors: terminal.DefaultFlatStyleColors(),
	}
	for _, opt := range opts {
		opt(a)
	}

	cmd := &cli.Command{
		Use:           name,
		Short:         a.short,
		Long:          a.long,
		Version:       a.version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cli.Command, args []string) error {
			return a.preRun(cmd, args)
		},
	}

	cmd.PersistentFlags().StringVar(&a.configFile, "config", "", "config file (default search: home, user config dir, cwd)")
	cmd.PersistentFlags().StringVar(&a.output, "output", "table", "output format: table|json|yaml")
	cmd.PersistentFlags().BoolVar(&a.noColor, "no-color", false, "disable color output")
	cmd.PersistentFlags().BoolVar(&a.verbose, "verbose", false, "verbose logging")
	cmd.PersistentFlags().BoolVar(&a.debug, "debug", false, "debug logging")
	cmd.PersistentFlags().StringVar(&a.logLevelValue, "log-level", a.logLevel.String(), "log level: error|warn|info|debug|trace")
	cmd.PersistentFlags().StringVar(&a.logStyleValue, "log-style", a.logStyle.String(), "log style: tree|plain|time|time-level|logrus|k8s")

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

// Raw escapes to the underlying CLI command for advanced use.
// Prefer Add/Execute; Raw exists so the wrapper never blocks you.
func (a *App) Raw() *cli.Command {
	return a.cmd
}

// Execute runs args against a signal-scoped ctx (see termcore.SignalContext).
// Parse and setup failures return ExitCodeError(2); handler errors pass
// through for termcore.ExitCodeOf mapping by main.
func (a *App) Execute(ctx context.Context, args []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	a.cmd.SetArgs(args)
	a.ran = false

	if err := a.cmd.ExecuteContext(ctx); err != nil {
		var exitErr termcore.ExitCodeError
		if !errors.As(err, &exitErr) && !errors.Is(err, context.Canceled) && !a.ran {
			return termcore.NewExitCodeError(2, err)
		}
		return err
	}
	return nil
}

// preRun boots, in order: output, log-level, and log-style validation;
// debug switch; console setup; renderer + logger; layered config. The
// *Context travels via cmd.Context().
func (a *App) preRun(cmd *cli.Command, args []string) error {
	if a.output != termoutput.FormatTable && a.output != termoutput.FormatJSON && a.output != termoutput.FormatYAML {
		return fmt.Errorf("invalid --output %q: want table|json|yaml", a.output)
	}

	level := a.logLevel
	if cmd.Flags().Changed("log-level") {
		var err error
		level, err = termrender.ParseLogLevel(a.logLevelValue)
		if err != nil {
			return fmt.Errorf("invalid --log-level %q: %w", a.logLevelValue, err)
		}
	} else if a.verbose || a.debug {
		if level < termrender.DebugLevel {
			level = termrender.DebugLevel
		}
	} else {
		var err error
		level, err = termrender.ParseLogLevel(a.logLevelValue)
		if err != nil {
			return fmt.Errorf("invalid --log-level %q: %w", a.logLevelValue, err)
		}
	}

	style, err := termrender.ParseLogStyle(a.logStyleValue)
	if err != nil {
		return fmt.Errorf("invalid --log-style %q: %w", a.logStyleValue, err)
	}
	a.logStyle = style

	termcore.SetDebug(level >= termrender.DebugLevel)

	// Best effort: legacy Windows consoles need VT enabled before any
	// ANSI is written. Failure just means plain output there.
	_ = termcore.EnsureConsole()

	logOut := a.logOutput
	if logOut == nil {
		logOut = cmd.ErrOrStderr()
	}
	renderer, renderStop, err := a.selectRenderer(logOut)
	if err != nil {
		return err
	}
	logger := termrender.NewLogger(level, renderer)
	// Close drains logs and stops the consumer before the renderer is
	// finalized, so late producers cannot draw over stopped output.
	// Repeated stops stay safe.
	var stopOnce sync.Once
	stopAll := func() {
		stopOnce.Do(func() {
			logger.Close()
			renderStop()
		})
	}

	cfg, err := termconfig.Load(termconfig.Options{
		AppName:    a.name,
		ConfigFile: a.configFile,
		Overrides:  collectOverrides(cmd),
	})
	if err != nil {
		stopAll()
		return err
	}

	tctx := &Context{
		Context: cmd.Context(),
		cmd:     cmd,
		args:    args,
		stop:    stopAll,
		Config:  cfg,
		Logger:  logger,
		Out:     cmd.OutOrStdout(),
		ErrOut:  cmd.ErrOrStderr(),
		In:      cmd.InOrStdin(),
		Output:  a.output,
		NoColor: a.noColor,
		Verbose: a.verbose || a.debug || level >= termrender.DebugLevel,
		Emoji:   termcore.Emoji() && !a.symbolsOnly,
	}
	cmd.SetContext(context.WithValue(cmd.Context(), contextKey{}, tctx))
	return nil
}

// selectRenderer lets a custom factory take over, then falls back to
// the auto renderer for real terminal files and a plain simple renderer
// for captured writers (tests, pipes without files), so static output
// stays assertion-friendly. WithSymbolsOnly pins the interactive tree to
// ASCII glyphs on terminals. A non-tree log style always uses flat lines.
func (a *App) selectRenderer(out io.Writer) (termrender.LogRendered, func(), error) {
	file, _ := out.(*os.File)

	if a.logRenderer != nil {
		renderer, stop := a.logRenderer(LogRendererOptions{
			Out:         out,
			File:        file,
			Style:       a.logStyle,
			NoColor:     a.noColor,
			Colors:      a.styleColors,
			SymbolsOnly: a.symbolsOnly,
		})
		if renderer == nil {
			return nil, nil, fmt.Errorf("log renderer factory returned nil")
		}
		if stop == nil {
			stop = func() {}
		}
		return renderer, stop, nil
	}

	if file != nil {
		if a.symbolsOnly && a.logStyle == termrender.LogStyleTree {
			renderer, stop := a.symbolsRenderer(file)
			return renderer, stop, nil
		}
		renderer, stop := renderers.AutoRendererWithStyleColors(file, a.noColor, a.logStyle, a.styleColors)
		return renderer, stop, nil
	}
	if a.logStyle == termrender.LogStyleTree {
		renderer := renderers.NewSimpleRenderer(out, terminal.NoColorSchema())
		return renderer, func() {}, nil
	}
	renderer := renderers.NewSimpleRendererWithStyle(out, terminal.NoColorSchema(), a.logStyle)
	return renderer, func() {}, nil
}

// symbolsRenderer mirrors AutoRenderer's interactive branch with a
// symbols-only config, so house styles stay ASCII on any terminal.
func (a *App) symbolsRenderer(outFile *os.File) (termrender.LogRendered, func()) {
	cfg := renderconfig.NewDefaultSymbolsOnlyRenderingConfig()
	if a.noColor {
		cfg.Colors = terminal.NoColorSchema()
	}
	interactive := renderers.NewInteractiveRenderer(outFile, cfg)
	return interactive, interactive.StopDrawing
}

// build converts a Command, recursing into Children groups. Drawing starts
// inside RunE (not preRun) so --help/--version/completion never draw.
func (a *App) build(c Command) *cli.Command {
	cmd := &cli.Command{
		Use:               c.Use,
		Short:             c.Short,
		Long:              c.Long,
		Example:           c.Example,
		Args:              adaptArgs(c.Args),
		ValidArgs:         c.ValidArgs,
		ValidArgsFunction: adaptCompletion(c.ValidArgsFunction),
		Hidden:            c.Hidden,
		SilenceUsage:      true,
		SilenceErrors:     true,
		RunE: func(cmd *cli.Command, _ []string) error {
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
		c.Flags(&FlagSet{FlagSet: cmd.Flags()})
	}
	for _, child := range c.Children {
		cmd.AddCommand(a.build(child))
	}
	return cmd
}

// adaptArgs maps a terminull validator onto the engine hook. Nil keeps
// the engine default, so group parents and bare commands behave exactly
// as before.
func adaptArgs(pa PositionalArgs) cli.PositionalArgs {
	if pa == nil {
		return nil
	}
	return func(_ *cli.Command, args []string) error {
		return pa(args)
	}
}

// adaptCompletion maps a terminull completion onto the engine hook. The
// directive cast is safe: our values are bit-identical to the engine's
// (see TestDirectivesMatchEngine).
func adaptCompletion(fn CompletionFunc) func(*cli.Command, []string, string) ([]string, cli.ShellCompDirective) {
	if fn == nil {
		return nil
	}
	return func(_ *cli.Command, args []string, toComplete string) ([]string, cli.ShellCompDirective) {
		completions, directive := fn(args, toComplete)
		return completions, cli.ShellCompDirective(directive)
	}
}

// collectOverrides gathers explicitly-changed flags with types, so CLI
// flags win over env and files in the config layer.
func collectOverrides(cmd *cli.Command) map[string]any {
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
