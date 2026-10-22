// Package termpager pages long output through $PAGER (less by default) on
// capable terminals, and passes writers through untouched everywhere
// else (pipes, dumb terminals, missing pager).
//
// Usage in a command with long output:
//
//	out, stop, err := termpager.Start(ctx.Out)
//	if err != nil {
//	    return err
//	}
//	defer stop()
//	// ... write everything to out ...
//
// Quitting the pager early breaks later writes with EPIPE; treat those
// as success with termpager.Closed(err) instead of failing the command.
package termpager
