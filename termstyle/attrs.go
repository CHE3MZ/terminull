package termstyle

import (
	"github.com/CHE3MZ/terminull/v2/termstyle/internal/color"
)

// Attribute is a foreground color value. It aliases the owned color
// engine's attribute type, so style constants flow into Faint and Paint
// while user code imports only terminull.
type Attribute = color.Attribute

// Foreground colors for Faint and Paint. Values match the engine's
// (asserted in tests); the set is stable, so pinning it here carries
// no drift risk.
const (
	FgBlack   Attribute = Attribute(color.FgBlack)
	FgRed     Attribute = Attribute(color.FgRed)
	FgGreen   Attribute = Attribute(color.FgGreen)
	FgYellow  Attribute = Attribute(color.FgYellow)
	FgBlue    Attribute = Attribute(color.FgBlue)
	FgMagenta Attribute = Attribute(color.FgMagenta)
	FgCyan    Attribute = Attribute(color.FgCyan)
	FgWhite   Attribute = Attribute(color.FgWhite)

	FgHiBlack   Attribute = Attribute(color.FgHiBlack)
	FgHiRed     Attribute = Attribute(color.FgHiRed)
	FgHiGreen   Attribute = Attribute(color.FgHiGreen)
	FgHiYellow  Attribute = Attribute(color.FgHiYellow)
	FgHiBlue    Attribute = Attribute(color.FgHiBlue)
	FgHiMagenta Attribute = Attribute(color.FgHiMagenta)
	FgHiCyan    Attribute = Attribute(color.FgHiCyan)
	FgHiWhite   Attribute = Attribute(color.FgHiWhite)
)

// Text decorations for Paint and the Scheme helpers. Values match the
// engine's (asserted in tests). Every probed modern terminal renders
// all three; terminals without support show plain text, which is why
// they stay behind the same Enabled gate as colors.
const (
	Italic        Attribute = Attribute(color.Italic)
	Underline     Attribute = Attribute(color.Underline)
	Strikethrough Attribute = Attribute(color.CrossedOut)
)
