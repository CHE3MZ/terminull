package termapp

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestDirectivesMatchEngine pins our completion directives bit-identical
// to the engine's. The adaptCompletion cast depends on it; if upstream
// ever renumbers, this fails loudly instead of misbehaving quietly.
func TestDirectivesMatchEngine(t *testing.T) {
	t.Parallel()

	pairs := map[ShellCompDirective]cobra.ShellCompDirective{
		ShellCompDirectiveError:         cobra.ShellCompDirectiveError,
		ShellCompDirectiveNoSpace:       cobra.ShellCompDirectiveNoSpace,
		ShellCompDirectiveNoFileComp:    cobra.ShellCompDirectiveNoFileComp,
		ShellCompDirectiveFilterFileExt: cobra.ShellCompDirectiveFilterFileExt,
		ShellCompDirectiveFilterDirs:    cobra.ShellCompDirectiveFilterDirs,
		ShellCompDirectiveKeepOrder:     cobra.ShellCompDirectiveKeepOrder,
		ShellCompDirectiveDefault:       cobra.ShellCompDirectiveDefault,
	}
	for ours, engine := range pairs {
		if cobra.ShellCompDirective(ours) != engine {
			t.Fatalf("directive %d != engine %d", int(ours), int(engine))
		}
	}
}

func TestValidators(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		check   PositionalArgs
		args    []string
		wantErr string
	}{
		{"noargs ok", NoArgs, nil, ""},
		{"noargs rejects", NoArgs, []string{"x"}, `unknown command "x"`},
		{"arbitrary ok", ArbitraryArgs, []string{"a", "b", "c"}, ""},
		{"minimum ok", MinimumNArgs(2), []string{"a", "b"}, ""},
		{"minimum short", MinimumNArgs(2), []string{"a"}, "requires at least 2 arg(s), only received 1"},
		{"maximum ok", MaximumNArgs(1), []string{"a"}, ""},
		{"maximum over", MaximumNArgs(1), []string{"a", "b"}, "accepts at most 1 arg(s), received 2"},
		{"exact ok", ExactArgs(2), []string{"a", "b"}, ""},
		{"exact off", ExactArgs(2), []string{"a"}, "accepts 2 arg(s), received 1"},
		{"range ok", RangeArgs(1, 2), []string{"a"}, ""},
		{"range under", RangeArgs(1, 2), nil, "accepts between 1 and 2 arg(s), received 0"},
		{"range over", RangeArgs(1, 2), []string{"a", "b", "c"}, "accepts between 1 and 2 arg(s), received 3"},
		{"valid ok", OnlyValidArgs([]string{"prod", "staging"}), []string{"prod"}, ""},
		{"valid rejects", OnlyValidArgs([]string{"prod", "staging"}), []string{"dev"}, `invalid argument "dev"`},
		{"valid empty allows all", OnlyValidArgs(nil), []string{"anything"}, ""},
		{"valid strips descriptions", OnlyValidArgs([]string{"prod\tlive"}), []string{"prod"}, ""},
		{"matchall ok", MatchAll(ExactArgs(1), OnlyValidArgs([]string{"prod"})), []string{"prod"}, ""},
		{"matchall first fails", MatchAll(ExactArgs(1), OnlyValidArgs([]string{"prod"})), []string{"a", "b"}, "accepts 1 arg(s), received 2"},
		{"matchall second fails", MatchAll(ExactArgs(1), OnlyValidArgs([]string{"prod"})), []string{"dev"}, `invalid argument "dev"`},
	}
	for _, tc := range cases {
		err := tc.check(tc.args)
		if tc.wantErr == "" {
			if err != nil {
				t.Fatalf("%s: error = %v, want nil", tc.name, err)
			}
			continue
		}
		if err == nil {
			t.Fatalf("%s: error = nil, want %q", tc.name, tc.wantErr)
		}
		if !strings.Contains(err.Error(), tc.wantErr) {
			t.Fatalf("%s: error %q does not contain %q", tc.name, err.Error(), tc.wantErr)
		}
	}
}
