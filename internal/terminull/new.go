package terminull

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/CHE3MZ/terminull/v2/internal/generator"
	"github.com/CHE3MZ/terminull/v2/termapp"
	"github.com/CHE3MZ/terminull/v2/termcore"
	"github.com/CHE3MZ/terminull/v2/termoutput"
	"github.com/CHE3MZ/terminull/v2/termprogress"
	"github.com/CHE3MZ/terminull/v2/termprompt"
	"github.com/CHE3MZ/terminull/v2/termretry"
	"github.com/CHE3MZ/terminull/v2/termstatic"
	"github.com/CHE3MZ/terminull/v2/termstyle"
)

// NewApp builds the terminull CLI in its house style: plain ASCII
// spinners and tree glyphs on every terminal. Framework default stays
// capability-based (emoji where supported).
func NewApp() *termapp.App {
	a := termapp.New("terminull",
		termapp.WithShort("Scaffold batteries-included Go CLI apps."),
		termapp.WithLong("Scaffold batteries-included Go CLI apps with prompting, logging, and multi-format output built in."),
		termapp.WithSymbolsOnly(),
	)
	a.Add(newCommand())
	return a
}

func newCommand() termapp.Command {
	return termapp.Command{
		Use:   "new NAME",
		Short: "scaffold a new terminull project",
		Long:  "Scaffold a new terminull project: go.mod, a main package, sample commands, and a README.",
		Args:  termapp.MaximumNArgs(1),
		Example: `terminull new myapp --module example.com/myapp
  terminull new myapp --dir ./myapp --skip-tidy`,
		Flags: func(fs *termapp.FlagSet) {
			fs.String("module", "", "go module path (default: NAME)")
			fs.String("dir", "", "target directory (default: ./NAME)")
			fs.String("replace", "", "replace framework with local path (dogfood)")
			fs.String("terminull-version", generator.DefaultTerminullVersion, "framework version to require")
			fs.Bool("skip-tidy", false, "skip go mod tidy (no network needed)")
		},
		Run: runNew,
	}
}

func runNew(ctx *termapp.Context) error {
	module, err := ctx.String("module")
	if err != nil {
		return err
	}
	dir, err := ctx.String("dir")
	if err != nil {
		return err
	}
	replace, err := ctx.String("replace")
	if err != nil {
		return err
	}
	version, err := ctx.String("terminull-version")
	if err != nil {
		return err
	}
	skipTidy, err := ctx.Bool("skip-tidy")
	if err != nil {
		return err
	}

	if err := termstatic.Banner(ctx.Out, "terminull", version, termcore.ColorFor(ctx.Out, ctx.NoColor)); err != nil {
		return err
	}

	name, err := appName(ctx)
	if err != nil {
		return err
	}
	if err := generator.ValidateName(name); err != nil {
		return termcore.NewExitCodeError(2, err)
	}

	if module == "" {
		module, err = defaultModule(ctx, name)
		if err != nil {
			return err
		}
	}
	if dir == "" {
		dir = name
	}

	files, err := generator.Generate(generator.Options{
		Name:             name,
		Module:           module,
		Dir:              dir,
		TerminullVersion: version,
		Replace:          replace,
	})
	if err != nil {
		return err
	}

	if !skipTidy {
		if err := tidyWithSpinner(ctx, dir); err != nil {
			return err
		}
	}

	// The tree opens only after all foreign output (prompts, spinners)
	// is done, and completes before the result table prints: a single
	// writer owns the terminal in every phase, so frames never stomp
	// tables and tables never split frames.
	log := ctx.Logger.Scoped("scaffold")
	log.Infof("creating %s in %s", name, dir)
	for _, file := range files {
		log.Infof("wrote %s", file)
	}
	log.Finish(true)
	ctx.StopDrawing()

	rows := make([][]string, 0, len(files))
	for _, file := range files {
		rows = append(rows, []string{file})
	}
	scheme := termstyle.NewScheme(termcore.ColorFor(ctx.Out, ctx.NoColor))
	if err := termoutput.TableWith(ctx.Out, termoutput.TableValue{
		Header: []string{"created"},
		Rows:   rows,
	}, scheme); err != nil {
		return err
	}
	_, err = fmt.Fprintln(ctx.Out, scheme.Faint(termstyle.FgWhite, "next: cd "+dir+" && go run ./cmd/"+name+" --help"))
	return err
}

// appName returns the NAME arg, asks for it on interactive terminals,
// or fails with usage when there is nowhere to ask.
func appName(ctx *termapp.Context) (string, error) {
	if args := ctx.Args(); len(args) > 0 {
		return args[0], nil
	}
	if !stdinTerminal(ctx) {
		return "", termcore.NewExitCodeError(2, errors.New("missing app NAME\n\nUsage: terminull new NAME [flags]"))
	}
	name, err := solo(ctx, func(c context.Context, io termprompt.IO) (string, error) {
		return termprompt.Input(c, io, termprompt.InputOptions{
			Message: "App name?",
			Flag:    "NAME",
		})
	})
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", termcore.NewExitCodeError(2, errors.New("missing app NAME\n\nUsage: terminull new NAME [flags]"))
	}
	return name, nil
}

// solo pauses live drawing while fn produces foreign terminal output
// (prompts, spinners, bars), then resumes. Tree state is untouched.
// Note: termapp.Context embeds context.Context but shadows Err with an
// io.Writer field, so the inner context is passed explicitly.
func solo[T any](ctx *termapp.Context, fn func(context.Context, termprompt.IO) (T, error)) (T, error) {
	ctx.PauseDrawing()
	defer ctx.ResumeDrawing()
	return fn(ctx.Context, promptIO(ctx))
}

// defaultModule returns the --module flag, asks interactively when a
// terminal is present, or falls back to the app name non-interactively.
func defaultModule(ctx *termapp.Context, name string) (string, error) {
	if !stdinTerminal(ctx) {
		return name, nil
	}
	return solo(ctx, func(c context.Context, io termprompt.IO) (string, error) {
		return termprompt.Input(c, io, termprompt.InputOptions{
			Message: "Go module path?",
			Default: name,
			Flag:    "--module",
		})
	})
}

func stdinTerminal(ctx *termapp.Context) bool {
	if f, ok := ctx.In.(*os.File); ok {
		return termcore.IsTerminal(f)
	}
	return false
}

// promptIO bridges app IO to prompt IO, starting from process stdio and
// narrowing to file streams. Nil streams become flag-required errors
// instead of hangs. Icon colors follow the same terminal detection as
// every other output.
func promptIO(ctx *termapp.Context) termprompt.IO {
	io := termprompt.StdIO()
	io.Err = ctx.ErrOut
	io.NoColor = !termcore.ColorFor(ctx.Out, ctx.NoColor)
	if f, ok := ctx.In.(termprompt.FileReader); ok {
		io.In = f
	}
	if f, ok := ctx.Out.(termprompt.FileWriter); ok {
		io.Out = f
	}
	return io
}

// tidyWithSpinner runs go mod tidy in dir with a spinner. Drawing stays
// paused throughout: the spinner owns the terminal until it stops. The
// constant "go" binary keeps gosec happy; a missing toolchain is clear.
func tidyWithSpinner(ctx *termapp.Context, dir string) error {
	ctx.PauseDrawing()
	defer ctx.ResumeDrawing()

	spin := termprogress.NewSpinner(ctx.Out, "resolving modules", false)
	spin.Start()
	var out string
	err := termretry.Do(func() error {
		var err error
		out, err = runGoModTidy(dir)
		return err
	}, termretry.Options{Attempts: 3, Delay: 2 * time.Second})
	spin.Stop()
	if err != nil {
		return fmt.Errorf("go mod tidy: %w: %s", err, out)
	}
	return nil
}

func runGoModTidy(dir string) (string, error) {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return string(out), errors.New("go toolchain not found in PATH")
		}
		return string(out), err
	}
	return string(out), nil
}
