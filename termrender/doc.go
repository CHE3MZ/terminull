// Package termrender is terminull's hierarchical log renderer.
//
// It is a fork of github.com/cirruslabs/echelon (MIT, Copyright
// 2017-present Cirrus Labs Inc., see LICENSE in this directory),
// modernized for current Go and x/sys, with CI fold decorators
// (Travis, GitHub Actions, TeamCity) folded in from cirrus-cli.
//
// Business logic depends only on *Logger (Scoped/Logf/Finish) and the
// LogRendered interface. main wires the concrete renderer once:
//
//	interactive with emoji on a modern terminal,
//	symbols-only on legacy Windows consoles,
//	plain simple lines on CI, pipes, and dumb terminals.
package termrender
