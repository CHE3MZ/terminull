// Package progress owns spinners and bars. Apps never construct the
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
// Frame rule: braille dots on capable terminals, plain ASCII spinners
// (|/-\ differences) on legacy Windows consoles and anywhere else
// style.Emoji() is false.
package progress
