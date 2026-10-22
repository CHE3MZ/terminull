package termstyle

import (
	"github.com/CHE3MZ/terminull/termstyle/internal/color"
)

// Paint wraps s in the given ANSI color when enabled is true,
// and returns s unchanged when false.
//
// The per-instance switch (not the global color.NoColor) keeps this
// hermetic: parallel code paths can paint differently without racing.
func Paint(enabled bool, attr Attribute, s string) string {
	c := color.New(attr)
	if enabled {
		c.EnableColor()
	} else {
		c.DisableColor()
	}

	return c.Sprint(s)
}
