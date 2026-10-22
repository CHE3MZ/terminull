package terminal

import "fmt"

// ColorSchema maps outcomes to ANSI color codes (see color constants).
type ColorSchema struct {
	SuccessColor int
	FailureColor int
	NeutralColor int
}

// ResetSequence clears all ANSI attributes.
const ResetSequence = "\033[0m"

// NoColor disables coloring for a slot.
const NoColor = -1

const (
	BlackColor = iota
	RedColor
	GreenColor
	YellowColor
	BlueColor
	MagentaColor
	CyanColor
	WhiteColor
)

// DefaultColorSchema paints success green, failure red, neutral yellow.
func DefaultColorSchema() *ColorSchema {
	return &ColorSchema{
		SuccessColor: GreenColor,
		FailureColor: RedColor,
		NeutralColor: YellowColor,
	}
}

// NoColorSchema disables all coloring (NO_COLOR, dumb, piped).
func NoColorSchema() *ColorSchema {
	return &ColorSchema{
		SuccessColor: NoColor,
		FailureColor: NoColor,
		NeutralColor: NoColor,
	}
}

// GetColoredText wraps text in the color code, or returns it unchanged
// for NoColor.
func GetColoredText(color int, text string) string {
	if color == NoColor {
		return text
	}

	return fmt.Sprintf("%s%s%s", GetColorSequence(color), text, ResetSequence)
}

// GetColorSequence returns the ANSI escape for the color code.
func GetColorSequence(code int) string {
	if code < 0 {
		return ResetSequence
	}
	return fmt.Sprintf("\033[3%dm", code)
}
