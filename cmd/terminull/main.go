package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CHE3MZ/terminull/core"
	"github.com/CHE3MZ/terminull/internal/terminull"
)

func main() {
	ctx, cancel := core.SignalContext(context.Background())
	defer cancel()

	if err := terminull.NewApp().Execute(ctx, os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "terminull:", err)
		// Cancelled prompts surface as context.Canceled, mapped to 130.
		os.Exit(core.ExitCodeOf(err))
	}
}
