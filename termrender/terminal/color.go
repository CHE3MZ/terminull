package terminal

import "fmt"

// ColorSchema maps outcomes and flat log prefixes to ANSI color codes
// (see color constants).
type ColorSchema struct {
	SuccessColor int
	FailureColor int
	NeutralColor int
	// FlatStyle colors level, timestamp, and elapsed tokens in flat log
	// prefixes. Nil leaves those tokens uncolored.
	FlatStyle *FlatStyleColors
}

// FlatStyleColors maps flat log prefix tokens to ANSI color codes.
type FlatStyleColors struct {
	ErrorColor     int
	WarnColor      int
	InfoColor      int
	DebugColor     int
	TraceColor     int
	TimestampColor int
	ElapsedColor   int
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
	// GrayColor is ANSI bright black, a readable muted accent on both
	// dark and light terminals.
	GrayColor
)

// DefaultColorSchema paints success green, failure red, neutral yellow,
// and flat log tokens by severity and role.
func DefaultColorSchema() *ColorSchema {
	return &ColorSchema{
		SuccessColor: GreenColor,
		FailureColor: RedColor,
		NeutralColor: YellowColor,
		FlatStyle:    DefaultFlatStyleColors(),
	}
}

// DefaultFlatStyleColors colors errors red, warnings yellow, info blue,
// and debug and trace gray. Timestamps and elapsed time keep the
// terminal's default foreground so they stay readable on light themes.
func DefaultFlatStyleColors() *FlatStyleColors {
	return &FlatStyleColors{
		ErrorColor:     RedColor,
		WarnColor:      YellowColor,
		InfoColor:      BlueColor,
		DebugColor:     GrayColor,
		TraceColor:     GrayColor,
		TimestampColor: NoColor,
		ElapsedColor:   NoColor,
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
	if code > WhiteColor {
		return fmt.Sprintf("\033[9%dm", code-WhiteColor-1)
	}
	return fmt.Sprintf("\033[3%dm", code)
}
