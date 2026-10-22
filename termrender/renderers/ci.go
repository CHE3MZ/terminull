package renderers

import (
	"strings"

	"github.com/CHE3MZ/terminull/termrender"
)

// FoldableLogsRenderer decorates a SimpleRenderer with CI fold markers:
// a start marker when a scope first opens, an end marker when it closes.
type FoldableLogsRenderer struct {
	delegate          *SimpleRenderer
	startFoldTemplate string
	endFoldTemplate   string
	escapeFunc        func(s string) string

	StubRenderer
}

// RenderScopeStarted opens a fold on first sight, then delegates.
func (r FoldableLogsRenderer) RenderScopeStarted(entry *termrender.LogScopeStarted) {
	if !r.delegate.ScopeHasStarted(entry.GetScopes()) {
		r.printFoldMessage(entry.GetScopes(), r.startFoldTemplate, true)
	}
	r.delegate.RenderScopeStarted(entry)
}

// RenderScopeFinished delegates, then closes the fold with the scope name
// (upstream left a literal %s here; the name is required to match folds).
func (r FoldableLogsRenderer) RenderScopeFinished(entry *termrender.LogScopeFinished) {
	r.delegate.RenderScopeFinished(entry)
	r.printFoldMessage(entry.GetScopes(), r.endFoldTemplate, true)
}

// RenderMessage delegates the message unchanged.
func (r FoldableLogsRenderer) RenderMessage(entry *termrender.LogEntryMessage) {
	r.delegate.RenderMessage(entry)
}

// RenderRawMessage writes bytes without any decoration.
func (r FoldableLogsRenderer) RenderRawMessage(message string) {
	r.delegate.RenderRawMessage(message)
}

func (r FoldableLogsRenderer) printFoldMessage(scopes []string, template string, needsLastScope bool) {
	if scopesCount := len(scopes); scopesCount > 0 {
		lastScope := scopes[scopesCount-1]

		if r.escapeFunc != nil {
			lastScope = r.escapeFunc(lastScope)
		}

		var args []any

		// Templates without a verb (e.g. GitHub's ##[endgroup]) take no
		// args: surplus Sprintf args would print as %!(EXTRA ...).
		if needsLastScope && strings.Contains(template, "%s") {
			args = append(args, lastScope)
		}

		foldingMessage := termrender.NewLogEntryMessage(scopes, termrender.InfoLevel, template, args...)
		r.delegate.RenderMessage(foldingMessage)
	}
}

// NewTravisCILogsRenderer wraps simple output in Travis fold markers.
func NewTravisCILogsRenderer(renderer *SimpleRenderer) termrender.LogRendered {
	return &FoldableLogsRenderer{
		delegate:          renderer,
		startFoldTemplate: "travis_fold:start:%s",
		endFoldTemplate:   "travis_fold:end:%s",
	}
}

// GithubActionsLogsRenderer wraps simple output in GitHub Actions groups.
type GithubActionsLogsRenderer struct {
	*FoldableLogsRenderer
}

// NewGithubActionsLogsRenderer wraps simple output in GitHub Actions groups.
func NewGithubActionsLogsRenderer(renderer *SimpleRenderer) termrender.LogRendered {
	return &GithubActionsLogsRenderer{
		&FoldableLogsRenderer{
			delegate:          renderer,
			startFoldTemplate: "##[group]%s",
			endFoldTemplate:   "##[endgroup]",
		},
	}
}

// NewTeamCityLogsRenderer wraps simple output in TeamCity blocks,
// escaping service-message metacharacters in scope names.
func NewTeamCityLogsRenderer(renderer *SimpleRenderer) termrender.LogRendered {
	replacer := strings.NewReplacer(
		"'", "|'",
		"[", "|[",
		"]", "|]",
		"|", "||",
		"\n", "|n",
	)

	return &FoldableLogsRenderer{
		delegate:          renderer,
		startFoldTemplate: "##teamcity[blockOpened name='%s']",
		endFoldTemplate:   "##teamcity[blockClosed name='%s']",
		escapeFunc:        replacer.Replace,
	}
}
