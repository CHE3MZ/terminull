package generator

const goModTemplate = `module [[.Module]]

go 1.26.0

require github.com/CHE3MZ/terminull [[.TerminullVersion]]
[[if .Replace]]
replace github.com/CHE3MZ/terminull => "[[.Replace]]"
[[end]]`

const mainTemplate = `// Command [[.Name]] is a terminull CLI app.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CHE3MZ/terminull/termcore"
	"[[.Module]]/internal/commands"
)

func main() {
	ctx, cancel := termcore.SignalContext(context.Background())
	defer cancel()

	a := commands.NewApp()
	if err := a.Execute(ctx, os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "[[.Name]]:", err)
		// Cancelled prompts surface as context.Canceled, mapped to 130.
		os.Exit(termcore.ExitCodeOf(err))
	}
}
`

const commandsTemplate = `// Package commands wires the [[.Name]] commands.
package commands

import (
	"github.com/CHE3MZ/terminull/termapp"
	"github.com/CHE3MZ/terminull/termoutput"
)

// NewApp builds the CLI with its commands and global flags.
func NewApp() *termapp.App {
	a := termapp.New("[[.Name]]")
	a.Add(helloCommand())
	a.Add(statusCommand())
	return a
}

func helloCommand() termapp.Command {
	return termapp.Command{
		Use:   "hello",
		Short: "greet someone",
		Flags: func(fs *termapp.FlagSet) {
			fs.String("name", "world", "who to greet")
		},
		Run: func(ctx *termapp.Context) error {
			name, err := ctx.String("name")
			if err != nil {
				return err
			}
			log := ctx.Logger.Scoped("hello")
			log.Infof("greeting %s", name)
			log.Finish(true)

			if ctx.Output == termoutput.FormatTable {
				return termoutput.Render(ctx.Out, ctx.Output, termoutput.TableValue{
					Header: []string{"GREETING"},
					Rows:   [][]string{{"hello " + name}},
				})
			}
			return termoutput.Render(ctx.Out, ctx.Output, map[string]string{
				"greeting": "hello " + name,
			})
		},
	}
}

func statusCommand() termapp.Command {
	return termapp.Command{
		Use:   "status",
		Short: "show service status",
		Run: func(ctx *termapp.Context) error {
			names := []string{"api", "worker"}
			statuses := []string{"up", "up"}

			if ctx.Output == termoutput.FormatTable {
				rows := make([][]string, 0, len(names))
				for i := range names {
					rows = append(rows, []string{names[i], statuses[i]})
				}
				return termoutput.Render(ctx.Out, ctx.Output, termoutput.TableValue{
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
			return termoutput.Render(ctx.Out, ctx.Output, services)
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
    go run ./cmd/[[.Name]] status --output json
`
