package termimage

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/CHE3MZ/terminull/style"
)

// SGR codes that return colors to the terminal defaults, so unpainted
// halves show whatever the terminal already draws behind the image.
const (
	defaultFgSeq = "\x1b[39m"
	defaultBgSeq = "\x1b[49m"
)

// Cell color state keys: -1 means unset, -2 means default colors, any
// other value packs an RGB triple.
const (
	keyUnset   = -1
	keyDefault = -2
)

// halfBlockString renders img as half blocks: each cell paints two image
// rows, the upper behind the foreground and the lower behind the
// background, so two image rows share one terminal row. Odd heights
// duplicate the final row. Pixels under the alpha floor stay unpainted:
// their half falls back to the terminal default, leaving the background
// visible. Depth picks truecolor or 256-color codes; dither diffuses
// error through the xterm palette at 256 colors, where snapping would
// band. SGR runs are emitted only when a color changes.
func halfBlockString(img *image.RGBA, depth int, dither bool) string {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	rows := (h + 1) / 2

	pix := func(x, y int) color.RGBA {
		p := y*img.Stride + x*4
		return color.RGBA{R: img.Pix[p], G: img.Pix[p+1], B: img.Pix[p+2], A: img.Pix[p+3]}
	}
	if depth == style.Depth256 && dither {
		pal := xtermRGBA()
		idx, _, none := mapToPalette(img, pal, true)
		pix = func(x, y int) color.RGBA {
			if none[y*w+x] {
				return color.RGBA{}
			}
			return pal[idx[y*w+x]]
		}
	}

	var b strings.Builder
	for row := 0; row < rows; row++ {
		top := row * 2
		bot := top + 1
		if bot >= h {
			bot = h - 1
		}
		prevFg, prevBg := keyUnset, keyUnset
		for x := 0; x < w; x++ {
			t := pix(x, top)
			u := pix(x, bot)
			topNone := t.A < alphaFloor
			botNone := u.A < alphaFloor

			glyph := " "
			fgKey, bgKey := keyDefault, keyDefault
			var fgSeq, bgSeq string
			switch {
			case topNone && botNone:
				// Blank cell: default colors only.
			case botNone:
				glyph = "▀"
				fgKey = pack(t.R, t.G, t.B)
				fgSeq = colorSequence(depth, true, t.R, t.G, t.B)
			case topNone:
				glyph = "▄"
				fgKey = pack(u.R, u.G, u.B)
				fgSeq = colorSequence(depth, true, u.R, u.G, u.B)
			default:
				glyph = "▀"
				fgKey = pack(t.R, t.G, t.B)
				fgSeq = colorSequence(depth, true, t.R, t.G, t.B)
				bgKey = pack(u.R, u.G, u.B)
				bgSeq = colorSequence(depth, false, u.R, u.G, u.B)
			}

			if fgKey != prevFg {
				if fgKey == keyDefault {
					b.WriteString(defaultFgSeq)
				} else {
					b.WriteString(fgSeq)
				}
				prevFg = fgKey
			}
			if bgKey != prevBg {
				if bgKey == keyDefault {
					b.WriteString(defaultBgSeq)
				} else {
					b.WriteString(bgSeq)
				}
				prevBg = bgKey
			}
			b.WriteString(glyph)
		}
		b.WriteString("\x1b[0m")
		if row < rows-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func pack(r, g, b uint8) int {
	return int(r)<<16 | int(g)<<8 | int(b)
}

// colorSequence emits an SGR foreground or background code at the
// requested depth; 256 mode snaps through the xterm palette.
func colorSequence(depth int, fg bool, r, g, b uint8) string {
	lead := "38"
	if !fg {
		lead = "48"
	}
	if depth == style.DepthTrue {
		return fmt.Sprintf("\x1b[%s;2;%d;%d;%dm", lead, r, g, b)
	}
	return fmt.Sprintf("\x1b[%s;5;%dm", lead, paletteIndex(r, g, b))
}

// asciiRamp maps luminance to characters, lightest ink last.
const asciiRamp = " .:-=+*#%@"

// luminance is the integer Rec. 601 gray value of a pixel.
func luminance(r, g, b uint8) int {
	return (299*int(r) + 587*int(g) + 114*int(b)) / 1000
}

// asciiString renders img as a plain text ramp without any escapes.
// Pixels under the alpha floor stay blank, and dither diffuses
// luminance error across the ramp so gradients stop banding at the
// few steps the characters can show.
func asciiString(img *image.RGBA, dither bool) string {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	steps := len(asciiRamp) - 1
	var b strings.Builder

	if !dither {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				p := y*img.Stride + x*4
				if img.Pix[p+3] < alphaFloor {
					b.WriteByte(' ')
					continue
				}
				lum := luminance(img.Pix[p], img.Pix[p+1], img.Pix[p+2])
				b.WriteByte(asciiRamp[lum*steps/255])
			}
			if y < h-1 {
				b.WriteByte('\n')
			}
		}
		return b.String()
	}

	buf := make([]float64, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := y*img.Stride + x*4
			buf[y*w+x] = float64(luminance(img.Pix[p], img.Pix[p+1], img.Pix[p+2]))
		}
	}
	spread := func(x, y int, err float64) {
		if x < 0 || x >= w || y < 0 || y >= h {
			return
		}
		buf[y*w+x] += err
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := y*img.Stride + x*4
			if img.Pix[p+3] < alphaFloor {
				b.WriteByte(' ')
				continue
			}
			v := buf[y*w+x]
			switch {
			case v < 0:
				v = 0
			case v > 255:
				v = 255
			}
			i := int(v*float64(steps)/255 + 0.5)
			b.WriteByte(asciiRamp[i])
			err := v - float64(i)*255/float64(steps)
			spread(x+1, y, err*7.0/16)
			spread(x-1, y+1, err*3.0/16)
			spread(x, y+1, err*5.0/16)
			spread(x+1, y+1, err*1.0/16)
		}
		if y < h-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
