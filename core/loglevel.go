package core

import "log/slog"

// Level is the single shared slog level for the process.
// --verbose/--debug flip it to slog.LevelDebug via SetDebug.
var Level = &slog.LevelVar{}

// SetDebug enables debug logging when debug is true.
func SetDebug(debug bool) {
	if debug {
		Level.Set(slog.LevelDebug)
	}
}
