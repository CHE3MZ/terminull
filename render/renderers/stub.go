package renderers

import "github.com/CHE3MZ/terminull/render"

// StubRenderer drops every event. Use it for silent runs and tests
// that exercise logic without asserting output.
type StubRenderer struct{}

// RenderScopeStarted drops the event.
func (*StubRenderer) RenderScopeStarted(entry *render.LogScopeStarted) {}

// RenderScopeFinished drops the event.
func (*StubRenderer) RenderScopeFinished(entry *render.LogScopeFinished) {}

// RenderMessage drops the event.
func (*StubRenderer) RenderMessage(entry *render.LogEntryMessage) {}
