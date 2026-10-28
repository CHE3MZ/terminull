// Package termcore holds process-wide primitives shared by every terminull app:
//
//   - signals: a context cancelled on Ctrl+C (SIGINT) and SIGTERM
//   - exit codes: ExitCodeError maps errors to process exit codes
//   - version: binary version and commit, with `go install` fallback
//   - log level: one shared slog.LevelVar behind --verbose/--debug
//   - terminal facts: width, height, tty/color/console detection
//   - text: ANSI-aware width, truncation, and extended grapheme clusters
//   - terminal I/O: raw-mode capability queries, cell pixels, no-echo
//     password reads
//
// Core is the only library that imports x/sys and x/term, and it behaves
// the same way across Linux, macOS, and Windows.
package termcore
