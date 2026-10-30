# custom components

terminull has no plugin registry. A custom component is an ordinary Go
package that uses the public `term*` APIs. You own the package; terminull
owns the terminal primitives.

This keeps the boundary small: no loader, no global registration, no
third-party terminal mods, and no full-screen widget framework. A component
can live in your app under `internal/components/<name>` or in a separate
module when it is reusable.

See `examples/component` for a runnable widget plus command package.

## command components

A command component returns one or more `termapp.Command` values. The app
adds it explicitly:

```go
package status

func Command() termapp.Command {
	return termapp.Command{
		Use:   "status",
		Short: "show service status",
		Run: func(ctx *termapp.Context) error {
			return render(ctx)
		},
	}
}
```

```go
a := termapp.New("app")
a.Add(status.Command())
```

Command packages should:

- import public `term*` libraries only, never `internal/` or a fork
- return values instead of registering globally or in `init`
- read flags through `*termapp.Context` and write to `ctx.Out`/`ctx.ErrOut`
- call `ctx.StopDrawing` before final output, or pause/resume around
  transient widgets
- honor `ctx.Output` when the command returns data
- keep every prompt backed by a flag and safe on non-terminals

## widget components

A widget renders to a caller-owned writer. It does not create a logger,
detect the terminal, or decide color policy:

```go
func Render(w io.Writer, card Card, scheme termstyle.Scheme) error {
	content := scheme.Bold(card.Title) + "\n" + card.Message
	return termbox.Render(w, content, termbox.Options{Padding: 1}, scheme)
}
```

The command supplies the policy:

```go
scheme := termstyle.NewScheme(termstyle.ColorFor(ctx.Out, ctx.NoColor))
return statuscard.Render(ctx.Out, card, scheme)
```

Widgets must use `termcore`/`termstyle` width helpers, return write errors,
and stay correct with color disabled, narrow terminals, and pipes. Prefer
existing `termbox` and `termtable` primitives over a new renderer.

## renderer components

Custom log renderers implement `termrender.LogRendered` and are installed
with `termapp.WithLogRendererFactory`. They must honor `NoColor`,
`SymbolsOnly`, CI modes, and non-`*os.File` writers, and return a stop
function that joins the draw loop.

## styles and data

Prefer data over code. Use built-in `termtable` styles and
`termbox.Options` instead of a custom renderer. If a seam is missing, add an
option to the owning library. `termprogress` already accepts custom spinner
frames and `termupdate` already accepts a custom `Source`.

## boundaries

Custom components are versioned by their own packages. terminull promises
its public APIs, not arbitrary component behavior. Do not add a component
registry, dynamic plugin loader, or an API that accepts "whatever the user
wants"; those turn the library into a platform and make the cross-platform
fallback guarantees impossible to test.
