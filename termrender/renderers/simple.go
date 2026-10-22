package renderers

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/CHE3MZ/terminull/termrender"
	"github.com/CHE3MZ/terminull/termrender/renderers/internal/console"
	"github.com/CHE3MZ/terminull/termrender/terminal"
	"github.com/CHE3MZ/terminull/termrender/utils"
)

// SimpleRenderer prints linear lines: "Started 'scope'" then
// "'scope' succeeded/failed/skipped in <duration>!". Use it for CI,
// pipes, dumb terminals, and anywhere a live tree cannot draw.
type SimpleRenderer struct {
	out        io.Writer
	colors     *terminal.ColorSchema
	startTimes map[string]time.Time

	StubRenderer
}

// NewSimpleRenderer builds a SimpleRenderer; nil colors select the default schema.
func NewSimpleRenderer(out io.Writer, colors *terminal.ColorSchema) *SimpleRenderer {
	if colors == nil {
		colors = terminal.DefaultColorSchema()
	}
	_ = console.PrepareTerminalEnvironment()
	return &SimpleRenderer{
		out:        out,
		colors:     colors,
		startTimes: make(map[string]time.Time),
	}
}

// RenderScopeStarted prints "Started 'scope'", ignoring duplicates and roots.
func (r SimpleRenderer) RenderScopeStarted(entry *termrender.LogScopeStarted) {
	scopes := entry.GetScopes()
	level := len(scopes)
	if level == 0 {
		return
	}
	timeKey := strings.Join(scopes, "/")
	if _, ok := r.startTimes[timeKey]; ok {
		// duplicate event
		return
	}
	r.startTimes[timeKey] = time.Now()
	lastScope := scopes[level-1]
	message := terminal.GetColoredText(r.colors.NeutralColor, fmt.Sprintf("Started %s", quotedIfNeeded(lastScope)))
	r.RenderRawMessage(message + "\n")
}

// RenderScopeFinished prints "'scope' <outcome> in <duration>!".
func (r SimpleRenderer) RenderScopeFinished(entry *termrender.LogScopeFinished) {
	scopes := entry.GetScopes()
	level := len(scopes)
	if level == 0 {
		return
	}
	now := time.Now()
	startTime := now
	if t, ok := r.startTimes[strings.Join(scopes, "/")]; ok {
		startTime = t
	}
	duration := now.Sub(startTime)
	formatedDuration := utils.FormatDuration(duration, true)
	lastScope := scopes[level-1]

	switch entry.FinishType() {
	case termrender.FinishTypeSucceeded:
		message := fmt.Sprintf("%s succeeded in %s!", quotedIfNeeded(lastScope), formatedDuration)
		coloredMessage := terminal.GetColoredText(r.colors.SuccessColor, message)
		r.RenderRawMessage(coloredMessage + "\n")
	case termrender.FinishTypeFailed:
		message := fmt.Sprintf("%s failed in %s!", quotedIfNeeded(lastScope), formatedDuration)
		coloredMessage := terminal.GetColoredText(r.colors.FailureColor, message)
		r.RenderRawMessage(coloredMessage + "\n")
	case termrender.FinishTypeSkipped:
		message := fmt.Sprintf("%s skipped in %s!", quotedIfNeeded(lastScope), formatedDuration)
		coloredMessage := terminal.GetColoredText(r.colors.NeutralColor, message)
		r.RenderRawMessage(coloredMessage + "\n")
	}
}

// RenderMessage prints the message as-is.
func (r SimpleRenderer) RenderMessage(entry *termrender.LogEntryMessage) {
	r.RenderRawMessage(entry.GetMessage())
}

// RenderRawMessage writes bytes without any decoration.
func (r SimpleRenderer) RenderRawMessage(message string) {
	_, _ = r.out.Write([]byte(message))
}

// ScopeHasStarted reports whether the scope was seen starting
// (used by fold decorators to open a fold exactly once).
func (r SimpleRenderer) ScopeHasStarted(scopes []string) bool {
	level := len(scopes)
	if level == 0 {
		return true
	}
	timeKey := strings.Join(scopes, "/")
	_, result := r.startTimes[timeKey]
	return result
}

func quotedIfNeeded(s string) string {
	if strings.ContainsAny(s, "'\"") {
		return s
	}
	return "'" + s + "'"
}
