package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CHE3MZ/terminull/v2/internal/terminull"
	"github.com/CHE3MZ/terminull/v2/termcore"
)

func main() {
	ctx, cancel := termcore.SignalContext(context.Background())
	defer cancel()

	if err := terminull.NewApp().Execute(ctx, os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "terminull:", err)
		// Cancelled prompts surface as context.Canceled, mapped to 130.
		os.Exit(termcore.ExitCodeOf(err))
	}
}
