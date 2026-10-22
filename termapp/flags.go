package termapp

import (
	"github.com/CHE3MZ/terminull/termapp/internal/cli/pflag"
)

// FlagSet is the terminull flag surface. It embeds the internal flag
// set, so every registration and query method works with zero
// re-implementation and zero drift: user code imports only termapp.
//
// Exotic cases (custom Value implementations, AddFlagSet merges) use
// Unwrap, which hands back the underlying set.
type FlagSet struct {
	*pflag.FlagSet
}

// Unwrap returns the underlying flag set for cases the facade does not
// name, such as custom Value implementations and set merges.
func (f *FlagSet) Unwrap() *pflag.FlagSet {
	return f.FlagSet
}
