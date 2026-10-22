// Package termprompt asks one-shot questions: selects, confirms, inputs.
//
// The one-shot engine is deliberately plain: print the bullet question,
// read one line, parse the answer. No cursor tracking, no re-renders, no
// second render pass — so output cannot duplicate on any terminal, legacy
// consoles included. Numbered lists replace arrow-key menus; masked
// passwords use the OS no-echo read. Ctrl+C cancels through ctx.
//
// Every prompt takes a Preset (the already-parsed flag value). A valid
// preset returns immediately without touching the terminal, so flags
// always win. Without a preset the terminal must be interactive;
// otherwise the prompt fails naming the required flag instead of
// hanging on a pipe.
//
// Editor is the opt-in exception for repl-style input: liner-backed
// line editing with history, ctrl+r search, and tab completion. It
// lends the console to each Line call and hands it back, so one-shot
// prompts and render output behave normally in between; on pipes and
// dumb terminals Line degrades to a plain read.
//
// Render pausing is the caller's job: pause live drawing before
// prompting and resume after, or the menu and the tree garble each
// other. See app.Context PauseDrawing and ResumeDrawing.
package termprompt
