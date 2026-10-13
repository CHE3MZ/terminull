package render

import "io"

type genericLogEntry struct {
	LogStarted  *LogScopeStarted
	LogFinished *LogScopeFinished
	LogEntry    *LogEntryMessage
}

// LogRendered is the decoupling point: the same *Logger drives a live
// VT100 tree locally and plain annotated lines on CI.
type LogRendered interface {
	RenderScopeStarted(entry *LogScopeStarted)
	RenderScopeFinished(entry *LogScopeFinished)
	RenderMessage(entry *LogEntryMessage)
}

// Logger is goroutine-safe: one consumer goroutine drains entriesChannel
// while any number of producers log through Scoped/Logf/AsWriter.
type Logger struct {
	maxLogLevel    LogLevel
	scopes         []string
	entriesChannel chan *genericLogEntry
	renderer       LogRendered
}

type loggerAsWriter struct {
	logger *Logger
	level  LogLevel
}

func (w *loggerAsWriter) Write(p []byte) (n int, err error) {
	if w.logger.IsLogLevelEnabled(w.level) {
		logEntryMessage := NewLogEntryMessage(w.logger.scopes, w.level, "%s", string(p))
		logEntryMessage.raw = true
		w.logger.entriesChannel <- &genericLogEntry{LogEntry: logEntryMessage}
	}
	return len(p), err
}

// FinishType reports how a scope ended.
type FinishType int

const (
	FinishTypeSucceeded FinishType = iota
	FinishTypeFailed
	FinishTypeSkipped
)

// NewLogger starts the consumer goroutine and returns the root logger.
func NewLogger(level LogLevel, renderer LogRendered) *Logger {
	logger := &Logger{
		maxLogLevel:    level,
		entriesChannel: make(chan *genericLogEntry),
		renderer:       renderer,
	}
	go logger.streamEntries(renderer)
	return logger
}

// Renderer returns the underlying renderer (e.g. to stop drawing).
func (logger *Logger) Renderer() LogRendered {
	return logger.renderer
}

// Scoped emits a scope-started event and returns a child logger for it.
func (logger *Logger) Scoped(scope string) *Logger {
	result := &Logger{
		maxLogLevel:    logger.maxLogLevel,
		scopes:         append(logger.scopes, scope),
		entriesChannel: logger.entriesChannel,
	}
	result.entriesChannel <- &genericLogEntry{
		LogStarted: NewLogScopeStarted(result.scopes...),
	}
	return result
}

func (logger *Logger) streamEntries(renderer LogRendered) {
	for {
		entry := <-logger.entriesChannel
		if entry.LogStarted != nil {
			renderer.RenderScopeStarted(entry.LogStarted)
		}
		if entry.LogFinished != nil {
			renderer.RenderScopeFinished(entry.LogFinished)
		}
		if entry.LogEntry != nil {
			renderer.RenderMessage(entry.LogEntry)
		}
	}
}

// Tracef logs at TraceLevel.
func (logger *Logger) Tracef(format string, args ...any) {
	logger.Logf(TraceLevel, format, args...)
}

// Debugf logs at DebugLevel.
func (logger *Logger) Debugf(format string, args ...any) {
	logger.Logf(DebugLevel, format, args...)
}

// Infof logs at InfoLevel.
func (logger *Logger) Infof(format string, args ...any) {
	logger.Logf(InfoLevel, format, args...)
}

// Warnf logs at WarnLevel.
func (logger *Logger) Warnf(format string, args ...any) {
	logger.Logf(WarnLevel, format, args...)
}

// Errorf logs at ErrorLevel.
func (logger *Logger) Errorf(format string, args ...any) {
	logger.Logf(ErrorLevel, format, args...)
}

// Logf logs at the given level when enabled, dropping silently otherwise.
func (logger *Logger) Logf(level LogLevel, format string, args ...any) {
	if logger.IsLogLevelEnabled(level) {
		logger.entriesChannel <- &genericLogEntry{
			LogEntry: NewLogEntryMessage(logger.scopes, level, format, args...),
		}
	}
}

// AsWriter exposes the logger as an io.Writer (raw mode, for io.Copy of
// subprocess output) at the given level.
func (logger *Logger) AsWriter(level LogLevel) io.Writer {
	return &loggerAsWriter{logger: logger, level: level}
}

// Finish ends the scope as succeeded (true) or failed (false).
func (logger *Logger) Finish(success bool) {
	var finishType FinishType

	if success {
		finishType = FinishTypeSucceeded
	} else {
		finishType = FinishTypeFailed
	}

	logger.FinishWithType(finishType)
}

// FinishWithType ends the scope with an explicit FinishType.
func (logger *Logger) FinishWithType(finishType FinishType) {
	logger.entriesChannel <- &genericLogEntry{
		LogFinished: NewLogScopeFinished(finishType, logger.scopes...),
	}
}

// IsLogLevelEnabled reports whether level passes the logger's max.
func (logger *Logger) IsLogLevelEnabled(level LogLevel) bool {
	return level <= logger.maxLogLevel
}
