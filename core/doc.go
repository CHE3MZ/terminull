// Package core holds process-wide primitives shared by every terminull app:
//
//   - signals: a context cancelled on Ctrl+C (SIGINT) and SIGTERM
//   - exit codes: ExitCodeError maps errors to process exit codes
//   - version: binary version and commit, with `go install` fallback
//   - log level: one shared slog.LevelVar behind --verbose/--debug
//
// Core has no third-party dependencies and behaves identically on
// Linux, macOS, and Windows.
package core
