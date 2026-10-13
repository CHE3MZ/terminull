// Package prompt asks one-shot questions: selects, confirms, inputs.
//
// Every prompt takes a Preset (the already-parsed flag value). A valid
// preset returns immediately without touching the terminal, so flags
// always win. Without a preset the terminal must be interactive;
// otherwise the prompt fails naming the required flag instead of
// hanging on a pipe.
//
// Render pausing is the caller's job: stop or pause live drawing before
// prompting and resume after, or the menu and the tree garble each other.
// Ctrl+C surfaces as terminal.InterruptErr; use IsInterrupt to map it to
// exit code 130 in main.
//
// The interactive survey path is intentionally thin: unit tests stub the
// single askOne call site and assert the exact survey structs built.
// Survey's own suite plus the demo matrix cover real TTYs.
package prompt
