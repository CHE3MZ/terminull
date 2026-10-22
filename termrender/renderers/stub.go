package renderers

import "github.com/CHE3MZ/terminull/v2/termrender"

// StubRenderer drops every event. Use it for silent runs and tests
// that exercise logic without asserting output.
type StubRenderer struct{}

// RenderScopeStarted drops the event.
func (*StubRenderer) RenderScopeStarted(entry *termrender.LogScopeStarted) {}

// RenderScopeFinished drops the event.
func (*StubRenderer) RenderScopeFinished(entry *termrender.LogScopeFinished) {}

// RenderMessage drops the event.
func (*StubRenderer) RenderMessage(entry *termrender.LogEntryMessage) {}
