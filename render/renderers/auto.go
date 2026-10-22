package renderers

import (
	"io"
	"os"
	"strings"

	"github.com/CHE3MZ/terminull/core"
	"github.com/CHE3MZ/terminull/render"
	"github.com/CHE3MZ/terminull/render/renderers/config"
	"github.com/CHE3MZ/terminull/render/terminal"
)

// AutoRenderer picks the right renderer for out, mirroring cirrus-cli's
// auto output: a live tree on capable terminals, plain lines on CI,
// pipes, and dumb terminals. It returns the renderer and a stop func
// (deferred after the command runs; a no-op for simple output).
//
// noColor disables ANSI even where supported (--no-color / NO_COLOR).
// Pass os.Stdout for normal command output.
func AutoRenderer(out *os.File, noColor bool) (render.LogRendered, func()) {
	noop := func() {}

	if !core.IsTerminal(out) || os.Getenv("CI") == "true" || isDumbTerminal() {
		schema := terminal.DefaultColorSchema()
		if !core.Color(out, noColor) {
			schema = terminal.NoColorSchema()
		}
		return NewSimpleRenderer(colorableWriter(out), schema), noop
	}

	cfg := config.NewDefaultRenderingConfig()
	if noColor {
		cfg.Colors = terminal.NoColorSchema()
	}

	interactive := NewInteractiveRenderer(out, cfg)
	return interactive, interactive.StopDrawing
}

func isDumbTerminal() bool {
	return strings.EqualFold(os.Getenv("TERM"), "dumb")
}

// colorableWriter routes the standard streams through ANSI translation
// on legacy Windows consoles; any other file passes through unchanged
// (pipes and files need no translation).
func colorableWriter(out *os.File) io.Writer {
	switch out {
	case os.Stdout:
		return core.Stdout()
	case os.Stderr:
		return core.Stderr()
	default:
		return out
	}
}
