package termapp

import (
	"context"
	"io"

	"github.com/CHE3MZ/terminull/v2/termapp/internal/cli"
	"github.com/CHE3MZ/terminull/v2/termconfig"
	"github.com/CHE3MZ/terminull/v2/termrender"
)

// Compile-time proof that handlers can pass *Context anywhere a
// context.Context is wanted.
var _ context.Context = (*Context)(nil)

// contextKey carries the *Context built by PersistentPreRunE.
type contextKey struct{}

// Context is what Run handlers receive: cancellation, layered config,
// a scoped-capable logger, IO, and typed flag access. It satisfies
// context.Context, so pass it directly to cancellable helpers.
type Context struct {
	context.Context

	cmd    *cli.Command
	args   []string
	stop   func()
	Config *termconfig.Config
	Logger *termrender.Logger
	Out    io.Writer
	// ErrOut receives error output. Named to avoid shadowing the
	// embedded context's Err method, so *Context itself satisfies
	// context.Context and passes directly to cancellable helpers.
	ErrOut io.Writer
	In     io.Reader
	// Output is the validated --output value: table, json, or yaml.
	Output string
	// NoColor mirrors --no-color.
	NoColor bool
	// Verbose mirrors --verbose or --debug.
	Verbose bool
}

// Args returns the positional args after flag parsing.
func (c *Context) Args() []string {
	return c.args
}

// String returns the --name flag value (local or inherited global).
func (c *Context) String(name string) (string, error) {
	return c.cmd.Flags().GetString(name)
}

// Bool returns the --name flag value.
func (c *Context) Bool(name string) (bool, error) {
	return c.cmd.Flags().GetBool(name)
}

// Int returns the --name flag value.
func (c *Context) Int(name string) (int, error) {
	return c.cmd.Flags().GetInt(name)
}

// StringSlice returns the --name flag value.
func (c *Context) StringSlice(name string) ([]string, error) {
	return c.cmd.Flags().GetStringSlice(name)
}

// Changed reports whether --name was passed explicitly.
func (c *Context) Changed(name string) bool {
	return c.cmd.Flags().Changed(name)
}

// PauseDrawing freezes live tree redraws for foreign terminal output
// (prompts, spinners, bars) and ResumeDrawing continues afterwards.
// Both are safe no-ops on non-interactive renderers. Always pair them:
//
//	ctx.PauseDrawing()
//	defer ctx.ResumeDrawing()
func (c *Context) PauseDrawing() {
	if p, ok := c.Logger.Renderer().(interface{ PauseDrawing() }); ok {
		p.PauseDrawing()
	}
}

// ResumeDrawing unfreezes live tree redraws after PauseDrawing.
func (c *Context) ResumeDrawing() {
	if p, ok := c.Logger.Renderer().(interface{ ResumeDrawing() }); ok {
		p.ResumeDrawing()
	}
}

// StopDrawing drains pending log entries, then completes the tree and
// finalizes terminal output. Call it before printing result tables so
// nothing redraws over them; the deferred stop in RunE remains as a
// backstop. Safe to call twice and a no-op on non-interactive
// renderers or a nil logger.
func (c *Context) StopDrawing() {
	if c.Logger == nil {
		return
	}
	c.Logger.Sync()
	if s, ok := c.Logger.Renderer().(interface{ StopDrawing() }); ok {
		s.StopDrawing()
	}
}
