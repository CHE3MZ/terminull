package termrender

import (
	"fmt"
	"strings"
	"time"
)

// LogLevel orders severity: lower is more severe, so a max level of
// InfoLevel still passes Error and Warn while dropping Debug/Trace.
type LogLevel uint32

const (
	ErrorLevel LogLevel = iota
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

const logLevelChoices = "error|warn|info|debug|trace"

// String returns the flag value for a log level.
func (level LogLevel) String() string {
	switch level {
	case ErrorLevel:
		return "error"
	case WarnLevel:
		return "warn"
	case InfoLevel:
		return "info"
	case DebugLevel:
		return "debug"
	case TraceLevel:
		return "trace"
	default:
		return "unknown"
	}
}

// ParseLogLevel parses a --log-level flag value.
func ParseLogLevel(value string) (LogLevel, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "error", "err":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	case "trace":
		return TraceLevel, nil
	default:
		return 0, fmt.Errorf("want %s", logLevelChoices)
	}
}

// LogScopeStarted marks a scope opening.
type LogScopeStarted struct {
	scopes []string
	time   time.Time
}

// NewLogScopeStarted builds a scope-started event.
func NewLogScopeStarted(scopes ...string) *LogScopeStarted {
	return &LogScopeStarted{
		scopes: scopes,
		time:   time.Now(),
	}
}

// GetScopes returns the scope path.
func (entry *LogScopeStarted) GetScopes() []string {
	return entry.scopes
}

// GetTime returns when the scope started.
func (entry *LogScopeStarted) GetTime() time.Time {
	return entry.time
}

// LogScopeFinished marks a scope closing with a FinishType.
type LogScopeFinished struct {
	scopes     []string
	finishType FinishType
	time       time.Time
}

// NewLogScopeFinished builds a scope-finished event.
func NewLogScopeFinished(finishType FinishType, scopes ...string) *LogScopeFinished {
	return &LogScopeFinished{
		scopes:     scopes,
		finishType: finishType,
		time:       time.Now(),
	}
}

// FinishType returns how the scope ended.
func (entry *LogScopeFinished) FinishType() FinishType {
	return entry.finishType
}

// GetScopes returns the scope path.
func (entry *LogScopeFinished) GetScopes() []string {
	return entry.scopes
}

// GetTime returns when the scope finished.
func (entry *LogScopeFinished) GetTime() time.Time {
	return entry.time
}

// LogEntryMessage is a single formatted log line inside zero or more scopes.
type LogEntryMessage struct {
	Level   LogLevel
	scopes  []string
	message string
	raw     bool
	time    time.Time
}

// NewLogEntryMessage formats the message eagerly.
func NewLogEntryMessage(scopes []string, level LogLevel, format string, arguments ...any) *LogEntryMessage {
	return &LogEntryMessage{
		Level:   level,
		scopes:  scopes,
		message: fmt.Sprintf(format, arguments...),
		time:    time.Now(),
	}
}

// GetMessage returns the message with exactly one trailing newline,
// unless it arrived raw (e.g. via AsWriter), when it is kept as-is.
func (entry *LogEntryMessage) GetMessage() string {
	if entry.raw {
		return entry.message
	}
	return strings.TrimRight(entry.message, "\r\n") + "\n"
}

// GetScopes returns the scope path.
func (entry *LogEntryMessage) GetScopes() []string {
	return entry.scopes
}

// IsRaw reports whether the message arrived through AsWriter and should
// be written without added newline or decoration.
func (entry *LogEntryMessage) IsRaw() bool {
	return entry.raw
}

// GetTime returns when the message was created.
func (entry *LogEntryMessage) GetTime() time.Time {
	return entry.time
}
