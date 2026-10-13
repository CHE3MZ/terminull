package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CHE3MZ/terminull/core"
	"github.com/CHE3MZ/terminull/internal/terminull"
	"github.com/CHE3MZ/terminull/prompt"
)

func main() {
	ctx, cancel := core.SignalContext(context.Background())
	defer cancel()

	if err := terminull.NewApp().Execute(ctx, os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "terminull:", err)
		if prompt.IsInterrupt(err) {
			os.Exit(130)
		}
		os.Exit(core.ExitCodeOf(err))
	}
}
