// Package app wraps cobra so terminull apps never touch it directly.
//
// An App owns the root command, five global flags, and the boot order:
// debug switch, output validation, renderer selection, config load.
// Handlers receive a *Context (config, logger, IO, typed flag getters)
// instead of cobra types. Only two intentional leaks exist: Flags takes
// a *pflag.FlagSet and Args is a cobra.PositionalArgs.
package app
