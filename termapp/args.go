package termapp

import (
	"fmt"
	"strings"
)

// PositionalArgs validates positional args. It takes only args, unlike
// engine's hook: the framework owns the command, so validators stay pure
// functions over argv. Count messages match the engine's wording.
type PositionalArgs func(args []string) error

// NoArgs returns an error if any args are included.
func NoArgs(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unknown command %q: takes no positional args", args[0])
	}
	return nil
}

// ArbitraryArgs never returns an error.
func ArbitraryArgs(args []string) error {
	return nil
}

// MinimumNArgs returns an error if there are fewer than n args.
func MinimumNArgs(n int) PositionalArgs {
	return func(args []string) error {
		if len(args) < n {
			return fmt.Errorf("requires at least %d arg(s), only received %d", n, len(args))
		}
		return nil
	}
}

// MaximumNArgs returns an error if there are more than n args.
func MaximumNArgs(n int) PositionalArgs {
	return func(args []string) error {
		if len(args) > n {
			return fmt.Errorf("accepts at most %d arg(s), received %d", n, len(args))
		}
		return nil
	}
}

// ExactArgs returns an error if there are not exactly n args.
func ExactArgs(n int) PositionalArgs {
	return func(args []string) error {
		if len(args) != n {
			return fmt.Errorf("accepts %d arg(s), received %d", n, len(args))
		}
		return nil
	}
}

// RangeArgs returns an error if the arg count is outside [min, max].
func RangeArgs(min, max int) PositionalArgs {
	return func(args []string) error {
		if len(args) < min || len(args) > max {
			return fmt.Errorf("accepts between %d and %d arg(s), received %d", min, max, len(args))
		}
		return nil
	}
}

// OnlyValidArgs returns an error for args outside valid. Unlike the
// engine's version it takes the list explicitly instead of reading
// command state, so it stays a pure function. Tab-suffixed descriptions
// are stripped before comparing, matching the engine's ValidArgs convention.
func OnlyValidArgs(valid []string) PositionalArgs {
	return func(args []string) error {
		if len(valid) == 0 {
			return nil
		}
		names := make([]string, 0, len(valid))
		for _, v := range valid {
			names = append(names, strings.SplitN(v, "\t", 2)[0])
		}
		for _, v := range args {
			if !stringInSlice(v, names) {
				return fmt.Errorf("invalid argument %q: want one of %s", v, strings.Join(names, ", "))
			}
		}
		return nil
	}
}

// MatchAll combines validators, reporting the first failure.
func MatchAll(pargs ...PositionalArgs) PositionalArgs {
	return func(args []string) error {
		for _, parg := range pargs {
			if err := parg(args); err != nil {
				return err
			}
		}
		return nil
	}
}

func stringInSlice(s string, list []string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
