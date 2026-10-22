package style

import (
	"github.com/CHE3MZ/terminull/core"
)

// Color depth tiers reported by ColorDepth.
const (
	Depth16   = core.Depth16
	Depth256  = core.Depth256
	DepthTrue = core.DepthTrue
)

// ColorDepth reports what ANSI color tiers the terminal can paint.
func ColorDepth() int {
	return core.ColorDepth()
}
