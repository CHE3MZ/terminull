// Package termapp wraps cobra so terminull apps never touch it directly.
//
// An App owns the root command, five global flags, and the boot order:
// debug switch, output validation, renderer selection, config load.
// Handlers receive a *Context (config, logger, IO, typed flag getters)
// instead of cobra types. Commands declare Flags on app.FlagSet, Args
// with app validators, and completions with app directives. Only two
// intentional hatches exist: App.Raw escapes to the engine command and
// FlagSet.Unwrap escapes to the engine flag set.
package termapp
