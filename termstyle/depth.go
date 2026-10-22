package termstyle

import (
	"github.com/CHE3MZ/terminull/v2/termcore"
)

// Color depth tiers reported by ColorDepth.
const (
	Depth16   = termcore.Depth16
	Depth256  = termcore.Depth256
	DepthTrue = termcore.DepthTrue
)

// ColorDepth reports what ANSI color tiers the terminal can paint.
func ColorDepth() int {
	return termcore.ColorDepth()
}
