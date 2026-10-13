package generator

const goModTemplate = `module [[.Module]]

go 1.26.0

require github.com/CHE3MZ/terminull [[.TerminullVersion]]
[[if .Replace]]
replace github.com/CHE3MZ/terminull => "[[.Replace]]"
[[end]]`

const mainTemplate = `// Command [[.Name]] is a terminull app.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CHE3MZ/terminull/core"
	"github.com/CHE3MZ/terminull/prompt"
	"[[.Module]]/internal/commands"
)

func main() {
	ctx, cancel := core.SignalContext(context.Background())
	defer cancel()

	a := commands.NewApp()
	if err := a.Execute(ctx, os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "[[.Name]]:", err)
		if prompt.IsInterrupt(err) {
			os.Exit(130)
		}
		os.Exit(core.ExitCodeOf(err))
	}
}
`

const commandsTemplate = `// Package commands wires the [[.Name]] commands.
package commands

import (
	"github.com/spf13/pflag"

	"github.com/CHE3MZ/terminull/app"
	"github.com/CHE3MZ/terminull/output"
)

// NewApp builds the CLI with its commands and global flags.
func NewApp() *app.App {
	a := app.New("[[.Name]]")
	a.Add(helloCommand())
	a.Add(statusCommand())
	return a
}

func helloCommand() app.Command {
	return app.Command{
		Use:   "hello",
		Short: "greet someone",
		Flags: func(fs *pflag.FlagSet) {
			fs.String("name", "world", "who to greet")
		},
		Run: func(ctx *app.Context) error {
			name, err := ctx.String("name")
			if err != nil {
				return err
			}
			log := ctx.Logger.Scoped("hello")
			log.Infof("greeting %s", name)
			log.Finish(true)

			if ctx.Output == output.FormatTable {
				return output.Render(ctx.Out, ctx.Output, output.TableValue{
					Header: []string{"GREETING"},
					Rows:   [][]string{{"hello " + name}},
				})
			}
			return output.Render(ctx.Out, ctx.Output, map[string]string{
				"greeting": "hello " + name,
			})
		},
	}
}

func statusCommand() app.Command {
	return app.Command{
		Use:   "status",
		Short: "show service status",
		Run: func(ctx *app.Context) error {
			names := []string{"api", "worker"}
			statuses := []string{"up", "up"}

			if ctx.Output == output.FormatTable {
				rows := make([][]string, 0, len(names))
				for i := range names {
					rows = append(rows, []string{names[i], statuses[i]})
				}
				return output.Render(ctx.Out, ctx.Output, output.TableValue{
					Header: []string{"NAME", "STATUS"},
					Rows:   rows,
				})
			}
			services := make([]map[string]string, 0, len(names))
			for i := range names {
				services = append(services, map[string]string{
					"name":   names[i],
					"status": statuses[i],
				})
			}
			return output.Render(ctx.Out, ctx.Output, services)
		},
	}
}
`

const readmeTemplate = `# [[.Name]]

Built with [terminull](https://github.com/CHE3MZ/terminull) [[.TerminullVersion]].

## Run

    go mod tidy
    go run ./cmd/[[.Name]] --help
    go run ./cmd/[[.Name]] hello --name world
    go run ./cmd/[[.Name]] status -o json
`
