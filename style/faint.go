package style

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Color depths for dimming decisions.
const (
	depth16       = 16
	depth256      = 256
	depthTrue     = 1 << 24
	dimFactor     = 0.5
	resetSequence = "\x1b[0m"
)

// rgbCubeLevels are the 6 channel levels of the 256-color cube.
var rgbCubeLevels = []int{0, 95, 135, 175, 215, 255}

// ansiRGB maps the 16 ANSI foreground colors to approximate xterm RGB.
// Values only feed the dimmer, so rounding is invisible in output.
var ansiRGB = map[color.Attribute][3]int{
	color.FgBlack:     {0, 0, 0},
	color.FgRed:       {205, 0, 0},
	color.FgGreen:     {0, 205, 0},
	color.FgYellow:    {205, 205, 0},
	color.FgBlue:      {0, 0, 205},
	color.FgMagenta:   {205, 0, 205},
	color.FgCyan:      {0, 205, 205},
	color.FgWhite:     {229, 229, 229},
	color.FgHiBlack:   {127, 127, 127},
	color.FgHiRed:     {255, 0, 0},
	color.FgHiGreen:   {0, 255, 0},
	color.FgHiYellow:  {255, 255, 0},
	color.FgHiBlue:    {0, 0, 255},
	color.FgHiMagenta: {255, 0, 255},
	color.FgHiCyan:    {0, 255, 255},
	color.FgHiWhite:   {255, 255, 255},
}

// colorDepth detects palette support: COLORTERM truecolor/24bit wins,
// then TERM *256color*, else the 16-color base set.
func colorDepth() int {
	switch os.Getenv("COLORTERM") {
	case "truecolor", "24bit":
		return depthTrue
	}
	if strings.Contains(os.Getenv("TERM"), "256color") {
		return depth256
	}
	return depth16
}

// dimCode returns the escape body for a half-luminance variant of attr,
// or "" when attr is not a foreground color. The 16-color branch keeps
// its legacy mapping (bright downshifts, normal stays); wider palettes
// share dimRGB.
func dimCode(attr color.Attribute, depth int) string {
	rgb, ok := ansiRGB[attr]
	if !ok {
		return ""
	}
	if depth == depth16 {
		if attr >= color.FgHiBlack && attr <= color.FgHiWhite {
			return fmt.Sprintf("%d", int(attr)-60)
		}
		return fmt.Sprintf("%d", int(attr))
	}
	return dimRGB(rgb[0], rgb[1], rgb[2], depth)
}

// dimRGB clamps out-of-range channels, halves each channel, and emits
// the closest representable color: truecolor triplet, 256-cube index,
// or nearest base color for 16-color terminals.
func dimRGB(r, g, b int, depth int) string {
	r = int(float64(clampChannel(r)) * dimFactor)
	g = int(float64(clampChannel(g)) * dimFactor)
	b = int(float64(clampChannel(b)) * dimFactor)

	switch depth {
	case depthTrue:
		return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
	case depth256:
		return fmt.Sprintf("38;5;%d", cubeIndex(r, g, b))
	default:
		return nearestBase(r, g, b)
	}
}

func clampChannel(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// baseRGB holds the 8 normal colors in code order 30-37.
var baseRGB = [][3]int{
	ansiRGB[color.FgBlack],
	ansiRGB[color.FgRed],
	ansiRGB[color.FgGreen],
	ansiRGB[color.FgYellow],
	ansiRGB[color.FgBlue],
	ansiRGB[color.FgMagenta],
	ansiRGB[color.FgCyan],
	ansiRGB[color.FgWhite],
}

// nearestBase returns the escape code of the closest base color.
func nearestBase(r, g, b int) string {
	best, bestDist := 0, -1
	for i, c := range baseRGB {
		d := (r-c[0])*(r-c[0]) + (g-c[1])*(g-c[1]) + (b-c[2])*(b-c[2])
		if bestDist < 0 || d < bestDist {
			best, bestDist = i, d
		}
	}
	return fmt.Sprintf("%d", 30+best)
}

// cubeIndex finds the nearest 256-color cube entry to rgb.
func cubeIndex(r, g, b int) int {
	nearest := func(v int) int {
		best, bestDist := 0, 256
		for i, level := range rgbCubeLevels {
			if d := abs(v - level); d < bestDist {
				best, bestDist = i, d
			}
		}
		return best
	}
	return 16 + 36*nearest(r) + 6*nearest(g) + nearest(b)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// Faint paints str in a half-luminance variant of attr, working on every
// terminal that renders the color at all — unlike the SGR faint flag,
// which most terminals silently ignore. Returns str unchanged when
// disabled or when attr is not a foreground color.
func (s Scheme) Faint(attr color.Attribute, str string) string {
	if !s.Enabled {
		return str
	}
	code := dimCode(attr, colorDepth())
	if code == "" {
		return str
	}
	return "\x1b[" + code + "m" + str + resetSequence
}

// FaintRGB paints str in a half-luminance variant of an arbitrary color,
// for custom palettes, hex-derived values, and status colors outside the
// 16 named attributes. Out-of-range channels clamp to 0-255. On 16-color
// terminals it falls back to the nearest base color, documented as an
// approximation where dimming is unrepresentable.
func (s Scheme) FaintRGB(r, g, b int, str string) string {
	if !s.Enabled {
		return str
	}
	return "\x1b[" + dimRGB(r, g, b, colorDepth()) + "m" + str + resetSequence
}
