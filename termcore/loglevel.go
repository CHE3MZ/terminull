package termcore

import "log/slog"

// Level is the single shared slog level for the process.
// --verbose/--debug flip it to slog.LevelDebug via SetDebug. Apps wire
// it into their handler, e.g. slog.NewTextHandler(out,
// &slog.HandlerOptions{Level: termcore.Level}).
var Level = &slog.LevelVar{}

// SetDebug enables debug logging when debug is true and resets a debug
// level to info otherwise, so repeated app runs do not leak a previous
// debug setting without overriding a custom level.
func SetDebug(debug bool) {
	if debug {
		Level.Set(slog.LevelDebug)
		return
	}
	if Level.Level() == slog.LevelDebug {
		Level.Set(slog.LevelInfo)
	}
}
