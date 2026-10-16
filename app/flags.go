package app

import (
	"github.com/spf13/pflag"
)

// FlagSet is the terminull flag surface. It embeds *pflag.FlagSet, so
// every registration and query method works with zero re-implementation
// and zero drift: user code imports only app, never pflag.
//
// Exotic cases (custom Value types referenced by name, AddFlagSet merges)
// still need the pflag import in that one file; Unwrap hands back the
// underlying set for them.
type FlagSet struct {
	*pflag.FlagSet
}

// Unwrap returns the underlying pflag set for cases the facade does not
// name, such as custom Value implementations and set merges.
func (f *FlagSet) Unwrap() *pflag.FlagSet {
	return f.FlagSet
}
