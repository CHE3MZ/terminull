package renderers

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CHE3MZ/terminull/v2/termcore"
	"github.com/CHE3MZ/terminull/v2/termrender"
	"github.com/CHE3MZ/terminull/v2/termrender/terminal"
	"github.com/CHE3MZ/terminull/v2/termrender/utils"
)

// SimpleRenderer prints linear lines: "Started 'scope'" then
// "'scope' succeeded/failed/skipped in <duration>!". Use it for CI,
// pipes, dumb terminals, and anywhere a live tree cannot draw.
type SimpleRenderer struct {
	out        io.Writer
	colors     *terminal.ColorSchema
	startTimes map[string]time.Time
	style      termrender.LogStyle
	startedAt  time.Time

	StubRenderer
}

// NewSimpleRenderer builds a SimpleRenderer; nil colors select the default schema.
func NewSimpleRenderer(out io.Writer, colors *terminal.ColorSchema) *SimpleRenderer {
	return NewSimpleRendererWithStyle(out, colors, termrender.LogStylePlain)
}

// NewSimpleRendererWithStyle builds a SimpleRenderer with flat log styling.
func NewSimpleRendererWithStyle(out io.Writer, colors *terminal.ColorSchema, style termrender.LogStyle) *SimpleRenderer {
	if colors == nil {
		colors = terminal.DefaultColorSchema()
	}
	if style == termrender.LogStyleTree {
		style = termrender.LogStylePlain
	}
	_ = termcore.EnsureConsole()
	return &SimpleRenderer{
		out:        out,
		colors:     colors,
		startTimes: make(map[string]time.Time),
		style:      style,
		startedAt:  time.Now(),
	}
}

// NewFlatRenderer builds a styled SimpleRenderer for a terminal file,
// choosing color from the destination capability and --no-color.
func NewFlatRenderer(out *os.File, noColor bool, style termrender.LogStyle) *SimpleRenderer {
	return NewFlatRendererWithColors(out, noColor, style, terminal.DefaultFlatStyleColors())
}

// NewFlatRendererWithColors builds a NewFlatRenderer with a custom
// prefix palette. Nil colors leave level, timestamp, and elapsed tokens
// plain; --no-color still disables all ANSI.
func NewFlatRendererWithColors(out *os.File, noColor bool, style termrender.LogStyle, colors *terminal.FlatStyleColors) *SimpleRenderer {
	schema := terminal.DefaultColorSchema()
	schema.FlatStyle = colors
	if !termcore.Color(out, noColor) {
		schema = terminal.NoColorSchema()
	}
	return NewSimpleRendererWithStyle(colorableWriter(out), schema, style)
}

// RenderScopeStarted prints "Started 'scope'", ignoring duplicates and roots.
func (r SimpleRenderer) RenderScopeStarted(entry *termrender.LogScopeStarted) {
	scopes := entry.GetScopes()
	level := len(scopes)
	if level == 0 {
		return
	}
	timeKey := scopeKey(scopes)
	if _, ok := r.startTimes[timeKey]; ok {
		// duplicate event
		return
	}
	started := entry.GetTime()
	if started.IsZero() {
		started = time.Now()
	}
	r.startTimes[timeKey] = started
	lastScope := scopes[level-1]
	message := terminal.GetColoredText(r.colors.NeutralColor, fmt.Sprintf("Started %s", quotedIfNeeded(lastScope)))
	r.RenderRawMessage(r.prefix(started, termrender.InfoLevel) + message + "\n")
}

// RenderScopeFinished prints "'scope' <outcome> in <duration>!".
func (r SimpleRenderer) RenderScopeFinished(entry *termrender.LogScopeFinished) {
	scopes := entry.GetScopes()
	level := len(scopes)
	if level == 0 {
		return
	}
	now := entry.GetTime()
	if now.IsZero() {
		now = time.Now()
	}
	timeKey := scopeKey(scopes)
	startTime := now
	if t, ok := r.startTimes[timeKey]; ok {
		startTime = t
		delete(r.startTimes, timeKey)
	}
	duration := now.Sub(startTime)
	formatedDuration := utils.FormatDuration(duration, true)
	lastScope := scopes[level-1]

	switch entry.FinishType() {
	case termrender.FinishTypeSucceeded:
		message := fmt.Sprintf("%s succeeded in %s!", quotedIfNeeded(lastScope), formatedDuration)
		coloredMessage := terminal.GetColoredText(r.colors.SuccessColor, message)
		r.RenderRawMessage(r.prefix(now, termrender.InfoLevel) + coloredMessage + "\n")
	case termrender.FinishTypeFailed:
		message := fmt.Sprintf("%s failed in %s!", quotedIfNeeded(lastScope), formatedDuration)
		coloredMessage := terminal.GetColoredText(r.colors.FailureColor, message)
		r.RenderRawMessage(r.prefix(now, termrender.ErrorLevel) + coloredMessage + "\n")
	case termrender.FinishTypeSkipped:
		message := fmt.Sprintf("%s skipped in %s!", quotedIfNeeded(lastScope), formatedDuration)
		coloredMessage := terminal.GetColoredText(r.colors.NeutralColor, message)
		r.RenderRawMessage(r.prefix(now, termrender.WarnLevel) + coloredMessage + "\n")
	}
}

// RenderMessage prints the message, decorated when a flat style is active.
func (r SimpleRenderer) RenderMessage(entry *termrender.LogEntryMessage) {
	if !r.styled() || entry.IsRaw() {
		r.RenderRawMessage(entry.GetMessage())
		return
	}

	message := entry.GetMessage()
	prefix := r.prefix(entry.GetTime(), entry.Level)
	if !strings.HasSuffix(message, "\n") {
		r.RenderRawMessage(prefix + message)
		return
	}
	for _, line := range strings.Split(strings.TrimSuffix(message, "\n"), "\n") {
		r.RenderRawMessage(prefix + line + "\n")
	}
}

func (r SimpleRenderer) styled() bool {
	return r.style != termrender.LogStylePlain && r.style != termrender.LogStyleTree
}

func (r SimpleRenderer) prefix(at time.Time, level termrender.LogLevel) string {
	colors := r.flatColors()
	switch r.style {
	case termrender.LogStyleTime:
		return paintFlat(colors, flatTimestampColor(colors), "["+at.Format("15:04:05")+"]") + " "
	case termrender.LogStyleTimeLevel:
		return paintFlat(colors, flatTimestampColor(colors), "["+at.Format("15:04:05")+"]") + " " +
			paintFlat(colors, flatLevelColor(colors, level), logLevelName(level)) + " "
	case termrender.LogStyleLogrus:
		start := r.startedAt
		if start.IsZero() || at.Before(start) {
			start = at
		}
		elapsed := fmt.Sprintf("[%04d]", int(at.Sub(start)/time.Second))
		return paintFlat(colors, flatLevelColor(colors, level), logLevelName(level)) +
			paintFlat(colors, flatElapsedColor(colors), elapsed) + " "
	case termrender.LogStyleKubernetes:
		return paintFlat(colors, flatLevelColor(colors, level), kubernetesLevelName(level)) +
			paintFlat(colors, flatTimestampColor(colors), at.Format("0102 15:04:05.000000")) + " "
	default:
		return ""
	}
}

func (r SimpleRenderer) flatColors() *terminal.FlatStyleColors {
	if r.colors == nil {
		return nil
	}
	return r.colors.FlatStyle
}

func paintFlat(colors *terminal.FlatStyleColors, color int, text string) string {
	if colors == nil {
		return text
	}
	return terminal.GetColoredText(color, text)
}

func flatLevelColor(colors *terminal.FlatStyleColors, level termrender.LogLevel) int {
	if colors == nil {
		return terminal.NoColor
	}
	switch level {
	case termrender.ErrorLevel:
		return colors.ErrorColor
	case termrender.WarnLevel:
		return colors.WarnColor
	case termrender.InfoLevel:
		return colors.InfoColor
	case termrender.DebugLevel:
		return colors.DebugColor
	case termrender.TraceLevel:
		return colors.TraceColor
	default:
		return colors.InfoColor
	}
}

func flatTimestampColor(colors *terminal.FlatStyleColors) int {
	if colors == nil {
		return terminal.NoColor
	}
	return colors.TimestampColor
}

func flatElapsedColor(colors *terminal.FlatStyleColors) int {
	if colors == nil {
		return terminal.NoColor
	}
	return colors.ElapsedColor
}

func logLevelName(level termrender.LogLevel) string {
	switch level {
	case termrender.ErrorLevel:
		return "ERROR"
	case termrender.WarnLevel:
		return "WARN"
	case termrender.InfoLevel:
		return "INFO"
	case termrender.DebugLevel:
		return "DEBUG"
	case termrender.TraceLevel:
		return "TRACE"
	default:
		return "INFO"
	}
}

func kubernetesLevelName(level termrender.LogLevel) string {
	switch level {
	case termrender.ErrorLevel:
		return "E"
	case termrender.WarnLevel:
		return "W"
	case termrender.InfoLevel:
		return "I"
	case termrender.DebugLevel:
		return "D"
	case termrender.TraceLevel:
		return "T"
	default:
		return "I"
	}
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
	timeKey := scopeKey(scopes)
	_, result := r.startTimes[timeKey]
	return result
}

// scopeKey encodes scope names without slash collisions. Length prefixes
// keep ["a/b"] distinct from ["a", "b"] and preserve reuse after finish.
func scopeKey(scopes []string) string {
	var b strings.Builder
	for _, scope := range scopes {
		b.WriteString(strconv.Itoa(len(scope)))
		b.WriteByte(':')
		b.WriteString(scope)
	}
	return b.String()
}

func quotedIfNeeded(s string) string {
	if strings.ContainsAny(s, "'\"") {
		return s
	}
	return "'" + s + "'"
}
