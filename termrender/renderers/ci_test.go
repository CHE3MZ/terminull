package renderers

import (
	"strings"
	"testing"

	"github.com/CHE3MZ/terminull/termrender"
	"github.com/CHE3MZ/terminull/termrender/terminal"
)

func foldOutput(t *testing.T, makeRenderer func(*SimpleRenderer) termrender.LogRendered, scope string) string {
	t.Helper()

	var out safeBuffer
	simple := NewSimpleRenderer(&out, terminal.NoColorSchema())
	fold := makeRenderer(simple)
	fold.RenderScopeStarted(termrender.NewLogScopeStarted(scope))
	fold.RenderMessage(termrender.NewLogEntryMessage([]string{scope}, termrender.InfoLevel, "working"))
	fold.RenderScopeFinished(termrender.NewLogScopeFinished(termrender.FinishTypeSucceeded, scope))
	return out.String()
}

func TestTravisFoldMarkers(t *testing.T) {
	t.Parallel()

	got := foldOutput(t, NewTravisCILogsRenderer, "build")

	for _, want := range []string{"travis_fold:start:build", "travis_fold:end:build", "working"} {
		if !strings.Contains(got, want) {
			t.Fatalf("travis output %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "%!") || strings.Contains(got, "%s") {
		t.Fatalf("travis output %q has leftover format verbs", got)
	}
}

func TestGithubFoldMarkers(t *testing.T) {
	t.Parallel()

	got := foldOutput(t, NewGithubActionsLogsRenderer, "build")

	for _, want := range []string{"##[group]build", "##[endgroup]", "working"} {
		if !strings.Contains(got, want) {
			t.Fatalf("github output %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "EXTRA") {
		t.Fatalf("github output %q has surplus Sprintf args", got)
	}
}

func TestTeamCityFoldMarkers(t *testing.T) {
	t.Parallel()

	got := foldOutput(t, NewTeamCityLogsRenderer, "we'ird[name]")

	for _, want := range []string{
		"##teamcity[blockOpened name='we|'ird|[name|]']",
		"##teamcity[blockClosed name='we|'ird|[name|]']",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("teamcity output %q does not contain %q", got, want)
		}
	}
}
