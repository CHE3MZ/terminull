// Package termprogress owns spinners and bars. Apps never construct the
// underlying widgets directly: every widget shares one caller-provided
// writer (the render output), so concurrent spinners, bars, and logs
// cannot garble each other at the file level.
//
// Live-tree rule: stop (or never start) InteractiveRenderer drawing while
// a spinner or bar runs, and resume after. The widgets draw with their
// own escape traffic, which fights the tree's incremental redraws.
// Pipes, CI, and dumb terminals need no pausing: nothing live draws there.
//
// Terminal rule: spinners animate only when their destination is a
// terminal; Start elsewhere is a safe no-op. Bars always render (they
// degrade to plain lines), so CI logs keep full progress.
//
// Frame rule: New accepts any caller-provided frame list (see CharSets).
// NewSpinner picks braille dots when termcore.Emoji() allows them and
// ASCII elsewhere. Other callers should gate emoji and wide glyphs with
// termcore.Emoji() themselves.
//
// Redraw rule: ANSI-capable terminals clear the previous frame, including
// wrapped lines, and hide the cursor while animating; legacy Windows
// consoles overwrite with spaces instead. Stop clears the region before
// writing the final message.
//
// Scheme rule: widgets paint through an optional termstyle.Scheme
// (WithScheme); the zero scheme stays plain, so pipes and CI logs keep
// raw text.
package termprogress
