package termrender

import (
	"io"
	"sync"
)

type genericLogEntry struct {
	LogStarted  *LogScopeStarted
	LogFinished *LogScopeFinished
	LogEntry    *LogEntryMessage
	// barrier, when non-nil, is closed after all prior entries process.
	barrier chan struct{}
}

// LogRendered is the decoupling point: the same *Logger drives a live
// VT100 tree locally and plain annotated lines on CI.
type LogRendered interface {
	RenderScopeStarted(entry *LogScopeStarted)
	RenderScopeFinished(entry *LogScopeFinished)
	RenderMessage(entry *LogEntryMessage)
}

type loggerState struct {
	entriesChannel chan *genericLogEntry
	renderer       LogRendered
	done           chan struct{}
	closeOnce      sync.Once
	wg             sync.WaitGroup
}

// Logger is goroutine-safe: one consumer goroutine drains the shared
// channel while any number of producers log through Scoped/Logf/AsWriter.
// Close releases the consumer goroutine; after Close returns, logging is dropped.
type Logger struct {
	maxLogLevel LogLevel
	scopes      []string
	state       *loggerState
}

type loggerAsWriter struct {
	logger *Logger
	level  LogLevel
}

func (w *loggerAsWriter) Write(p []byte) (n int, err error) {
	if w.logger.IsLogLevelEnabled(w.level) {
		logEntryMessage := NewLogEntryMessage(w.logger.scopes, w.level, "%s", string(p))
		logEntryMessage.raw = true
		w.logger.send(&genericLogEntry{LogEntry: logEntryMessage})
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
// Call Close when the logger is no longer needed.
func NewLogger(level LogLevel, renderer LogRendered) *Logger {
	state := &loggerState{
		entriesChannel: make(chan *genericLogEntry),
		renderer:       renderer,
		done:           make(chan struct{}),
	}
	state.wg.Add(1)
	go streamEntries(state)

	return &Logger{
		maxLogLevel: level,
		state:       state,
	}
}

// Renderer returns the underlying renderer (e.g. to stop drawing).
func (logger *Logger) Renderer() LogRendered {
	return logger.state.renderer
}

// Scoped emits a scope-started event and returns a child logger for it.
func (logger *Logger) Scoped(scope string) *Logger {
	scopes := make([]string, len(logger.scopes)+1)
	copy(scopes, logger.scopes)
	scopes[len(scopes)-1] = scope
	result := &Logger{
		maxLogLevel: logger.maxLogLevel,
		scopes:      scopes,
		state:       logger.state,
	}
	logger.send(&genericLogEntry{
		LogStarted: NewLogScopeStarted(result.scopes...),
	})
	return result
}

func streamEntries(state *loggerState) {
	defer state.wg.Done()

	for {
		select {
		case entry := <-state.entriesChannel:
			if entry.barrier != nil {
				close(entry.barrier)
				continue
			}
			if entry.LogStarted != nil {
				state.renderer.RenderScopeStarted(entry.LogStarted)
			}
			if entry.LogFinished != nil {
				state.renderer.RenderScopeFinished(entry.LogFinished)
			}
			if entry.LogEntry != nil {
				state.renderer.RenderMessage(entry.LogEntry)
			}
		case <-state.done:
			return
		}
	}
}

func (logger *Logger) send(entry *genericLogEntry) bool {
	if logger == nil || logger.state == nil {
		return false
	}
	select {
	case logger.state.entriesChannel <- entry:
		return true
	case <-logger.state.done:
		return false
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
		logger.send(&genericLogEntry{
			LogEntry: NewLogEntryMessage(logger.scopes, level, format, args...),
		})
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
	logger.send(&genericLogEntry{
		LogFinished: NewLogScopeFinished(finishType, logger.scopes...),
	})
}

// IsLogLevelEnabled reports whether level passes the logger's max.
func (logger *Logger) IsLogLevelEnabled(level LogLevel) bool {
	return level <= logger.maxLogLevel
}

// Sync blocks until every entry sent so far is processed by the
// consumer goroutine. Call it before reading renderer state (or before
// StopDrawing) to avoid racing the drain: without Sync, Finish returns
// at channel handoff while the finish event may still be unprocessed.
func (logger *Logger) Sync() {
	if logger == nil || logger.state == nil {
		return
	}
	done := make(chan struct{})
	if !logger.send(&genericLogEntry{barrier: done}) {
		return
	}
	select {
	case <-done:
	case <-logger.state.done:
	}
}

// Close drains pending entries, stops the consumer goroutine, and waits
// for it to exit. It is safe to call more than once and from a child
// logger. Logging after Close is dropped without blocking.
func (logger *Logger) Close() {
	if logger == nil || logger.state == nil {
		return
	}
	state := logger.state
	state.closeOnce.Do(func() {
		logger.Sync()
		close(state.done)
		state.wg.Wait()
	})
}
