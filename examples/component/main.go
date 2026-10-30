package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CHE3MZ/terminull/v2/examples/component/statuscard"
	"github.com/CHE3MZ/terminull/v2/termapp"
	"github.com/CHE3MZ/terminull/v2/termcore"
)

func main() {
	ctx, cancel := termcore.SignalContext(context.Background())
	defer cancel()

	app := termapp.New("component-demo")
	app.Add(statuscard.Command())
	if err := app.Execute(ctx, os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "component-demo:", err)
		os.Exit(termcore.ExitCodeOf(err))
	}
}
