// Package style is the single place that decides how text looks.
//
// Detection (TTY? NO_COLOR? TERM=dumb? CI? old Windows console?) happens
// here once, so every app renders identically on Linux, macOS, and
// Windows. App code never prints with fmt directly and never imports
// color libraries directly; it asks style for writers and paint.
package style
