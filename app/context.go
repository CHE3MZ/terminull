package app

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/CHE3MZ/terminull/config"
	"github.com/CHE3MZ/terminull/render"
)

// contextKey carries the *Context built by PersistentPreRunE.
type contextKey struct{}

// Context is what Run handlers receive: cancellation, layered config,
// a scoped-capable logger, IO, and typed flag access.
type Context struct {
	context.Context

	cmd    *cobra.Command
	args   []string
	stop   func()
	Config *config.Config
	Logger *render.Logger
	Out    io.Writer
	Err    io.Writer
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
