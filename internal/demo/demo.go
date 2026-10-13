package demo

import (
	"errors"
	"time"

	"github.com/spf13/pflag"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/CHE3MZ/terminull/app"
	"github.com/CHE3MZ/terminull/core"
	"github.com/CHE3MZ/terminull/output"
	"github.com/CHE3MZ/terminull/progress"
	"github.com/CHE3MZ/terminull/prompt"
	"github.com/CHE3MZ/terminull/static"
	"github.com/CHE3MZ/terminull/style"
)

// DeployResult is the machine-readable deploy outcome.
type DeployResult struct {
	Env    string `json:"env" yaml:"env"`
	Status string `json:"status" yaml:"status"`
}

// Service is one status row.
type Service struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
	Bytes  uint64 `json:"bytes" yaml:"bytes"`
}

// NewApp builds the demo CLI: demo {deploy|status} plus standard
// --help/--version/completion and the five global flags.
func NewApp() *app.App {
	a := app.New("demo")
	a.Add(deployCommand())
	a.Add(statusCommand())
	return a
}

func deployCommand() app.Command {
	return app.Command{
		Use:   "deploy",
		Short: "deploy to an environment",
		Flags: func(fs *pflag.FlagSet) {
			fs.String("env", "", "target environment (dev|staging|prod)")
			fs.Bool("yes", false, "skip confirmation")
		},
		Run: runDeploy,
	}
}

func statusCommand() app.Command {
	return app.Command{
		Use:   "status",
		Short: "show service status",
		Run:   runStatus,
	}
}

func runDeploy(ctx *app.Context) error {
	// Static banner first: out-only, synchronous, style-aware.
	if err := static.Banner(ctx.Out, "demo", core.FullVersion(), !ctx.NoColor); err != nil {
		return err
	}

	// Environment: --env flag, config file, DEMO_ENV, prompt, in that order.
	// The flag already landed in config via overrides when passed.
	env, err := prompt.Select(promptIO(ctx), prompt.SelectOptions{
		Message: "Target environment?",
		Options: []string{"dev", "staging", "prod"},
		Default: "dev",
		Preset:  ctx.Config.String("env"),
		Flag:    "--env",
	})
	if err != nil {
		return err
	}

	yes, err := ctx.Bool("yes")
	if err != nil {
		return err
	}
	if !yes {
		ok, err := prompt.Confirm(promptIO(ctx), prompt.ConfirmOptions{
			Message: "Deploy to " + env + "?",
			Default: true,
			Flag:    "--yes",
		})
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("aborted by user")
		}
	}

	// Widgets share ctx.Out but run in sequence, never overlapping live
	// tree drawing (see package progress for the rule).
	spin := progress.NewSpinner(ctx.Out, "provisioning "+env, style.Emoji())
	spin.Start()
	time.Sleep(300 * time.Millisecond)
	spin.Stop()

	bar := progress.NewBar(ctx.Out, 4, "uploading")
	for range 4 {
		if err := bar.Add(1); err != nil {
			return err
		}
		time.Sleep(25 * time.Millisecond)
	}
	if err := bar.Finish(); err != nil {
		return err
	}

	log := ctx.Logger.Scoped("deploy")
	build := log.Scoped("build")
	build.Infof("compiling for %s", env)
	build.Finish(true)
	log.Finish(true)

	if ctx.Output == output.FormatTable {
		return output.Render(ctx.Out, ctx.Output, output.TableValue{
			Header: []string{"ENV", "STATUS"},
			Rows:   [][]string{{env, "deployed"}},
		})
	}
	return output.Render(ctx.Out, ctx.Output, []DeployResult{{Env: env, Status: "deployed"}})
}

func runStatus(ctx *app.Context) error {
	services := []Service{
		{Name: "api", Status: "up", Bytes: 1572864},
		{Name: "worker", Status: "up", Bytes: 524288},
	}

	if ctx.Output == output.FormatTable {
		rows := make([][]string, 0, len(services))
		for _, s := range services {
			rows = append(rows, []string{s.Name, s.Status, output.Bytes(s.Bytes)})
		}
		return output.Render(ctx.Out, ctx.Output, output.TableValue{
			Header: []string{"NAME", "STATUS", "SIZE"},
			Rows:   rows,
		})
	}
	return output.Render(ctx.Out, ctx.Output, services)
}

// promptIO bridges app IO to prompt IO. Streams without file
// descriptors stay nil, which the prompt guards turn into flag-required
// errors instead of hangs.
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
