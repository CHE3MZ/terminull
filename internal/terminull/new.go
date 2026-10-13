package terminull

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/CHE3MZ/terminull/app"
	"github.com/CHE3MZ/terminull/internal/generator"
	"github.com/CHE3MZ/terminull/output"
	"github.com/CHE3MZ/terminull/progress"
	"github.com/CHE3MZ/terminull/prompt"
	"github.com/CHE3MZ/terminull/static"
	"github.com/CHE3MZ/terminull/style"
)

// NewApp builds the terminull CLI.
func NewApp() *app.App {
	a := app.New("terminull")
	a.Add(newCommand())
	return a
}

func newCommand() app.Command {
	return app.Command{
		Use:   "new NAME",
		Short: "scaffold a new terminull project",
		Args:  cobra.ExactArgs(1),
		Flags: func(fs *pflag.FlagSet) {
			fs.String("module", "", "go module path (default: NAME)")
			fs.String("dir", "", "target directory (default: ./NAME)")
			fs.String("replace", "", "replace framework with local path (dogfood)")
			fs.String("terminull-version", generator.DefaultTerminullVersion, "framework version to require")
			fs.Bool("skip-tidy", false, "skip go mod tidy (no network needed)")
		},
		Run: runNew,
	}
}

func runNew(ctx *app.Context) error {
	name := ctx.Args()[0]

	module, err := ctx.String("module")
	if err != nil {
		return err
	}
	if module == "" {
		module, err = defaultModule(ctx, name)
		if err != nil {
			return err
		}
	}

	dir, err := ctx.String("dir")
	if err != nil {
		return err
	}
	if dir == "" {
		dir = name
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

	if err := static.Banner(ctx.Out, "terminull", version, !ctx.NoColor); err != nil {
		return err
	}

	log := ctx.Logger.Scoped("scaffold")
	log.Infof("creating %s in %s", name, dir)

	files, err := generator.Generate(generator.Options{
		Name:             name,
		Module:           module,
		Dir:              dir,
		TerminullVersion: version,
		Replace:          replace,
	})
	if err != nil {
		log.Finish(false)
		return err
	}
	for _, file := range files {
		log.Infof("wrote %s", file)
	}

	if !skipTidy {
		if err := tidyWithSpinner(ctx, dir); err != nil {
			log.Finish(false)
			return err
		}
	}
	log.Finish(true)

	rows := make([][]string, 0, len(files))
	for _, file := range files {
		rows = append(rows, []string{file})
	}
	if err := output.Table(ctx.Out, []string{"CREATED"}, rows); err != nil {
		return err
	}
	log.Infof("next: cd %s && go run ./cmd/%s --help", dir, name)
	return nil
}

// defaultModule returns the --module flag, asks interactively when a
// terminal is present, or falls back to the app name non-interactively.
func defaultModule(ctx *app.Context, name string) (string, error) {
	if !stdinTerminal(ctx) {
		return name, nil
	}
	return prompt.Input(promptIO(ctx), prompt.InputOptions{
		Message: "Go module path?",
		Default: name,
		Flag:    "--module",
	})
}

func stdinTerminal(ctx *app.Context) bool {
	if f, ok := ctx.In.(*os.File); ok {
		return style.IsTerminal(f)
	}
	return false
}

// promptIO bridges app IO to prompt IO; nil streams become
// flag-required errors instead of hangs.
func promptIO(ctx *app.Context) prompt.IO {
	var io prompt.IO
	io.Err = ctx.Err
	if f, ok := ctx.In.(terminal.FileReader); ok {
		io.In = f
	}
	if f, ok := ctx.Out.(terminal.FileWriter); ok {
		io.Out = f
	}
	return io
}

// tidyWithSpinner runs go mod tidy in dir with a spinner. The constant
// "go" binary keeps gosec happy; a missing toolchain is a clear error.
func tidyWithSpinner(ctx *app.Context, dir string) error {
	spin := progress.NewSpinner(ctx.Out, "resolving modules", style.Emoji())
	spin.Start()
	out, err := runGoModTidy(dir)
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
