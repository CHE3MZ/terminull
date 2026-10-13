// Package static prints framework-owned, out-only text: the app banner
// and the version line.
//
// Out-only means: never reads input, never starts spinners, bars, or
// prompts, and always renders through style so NO_COLOR, dumb terminals,
// pipes, and CI look right. Static output finishes fast and exits
// cleanly on Ctrl+C because there is nothing to interrupt: help,
// version, and completion short-circuit in cobra before any renderer
// starts (see package app), and Banner writes synchronously.
//
// App-specific result printing (lists, status) lives in package output
// and follows the same out-only rules with --output formats.
package static
