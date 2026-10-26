// Package termapp wraps its owned CLI engine so terminull apps never
// touch it directly.
//
// An App owns the root command, six global flags, and the boot order:
// debug switch, output and log-style validation, renderer selection,
// config load.
// Handlers receive a *Context (config, logger, IO, typed flag getters)
// instead of engine types. Commands declare Flags on termapp.FlagSet, Args
// with termapp validators, and completions with termapp directives. Only two
// intentional hatches exist: App.Raw escapes to the engine command and
// FlagSet.Unwrap escapes to the engine flag set.
package termapp
