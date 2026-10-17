package termimage

import (
	"fmt"
	"image"
	"strings"
)

// halfBlockString renders img as upper-half blocks: each cell paints the
// foreground over the background, so two image rows share one terminal
// row. Odd heights duplicate the final row. Depth picks truecolor or
// 256-color codes; SGR runs are emitted only when a color changes.
func halfBlockString(img *image.RGBA, depth int) string {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	rows := (h + 1) / 2

	var b strings.Builder
	for row := 0; row < rows; row++ {
		top := row * 2
		bot := top + 1
		if bot >= h {
			bot = h - 1
		}
		prevFg, prevBg := -1, -1
		for x := 0; x < w; x++ {
			fr, fg, fb := pixel(img, x, top)
			br, bg, bb := pixel(img, x, bot)
			fgKey := pack(fr, fg, fb)
			bgKey := pack(br, bg, bb)
			if fgKey != prevFg {
				b.WriteString(colorSequence(depth, true, fr, fg, fb))
				prevFg = fgKey
			}
			if bgKey != prevBg {
				b.WriteString(colorSequence(depth, false, br, bg, bb))
				prevBg = bgKey
			}
			b.WriteString("▀")
		}
		b.WriteString("\x1b[0m")
		if row < rows-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func pixel(img *image.RGBA, x, y int) (uint8, uint8, uint8) {
	p := y*img.Stride + x*4
	return img.Pix[p], img.Pix[p+1], img.Pix[p+2]
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
	if depth == depthTrue {
		return fmt.Sprintf("\x1b[%s;2;%d;%d;%dm", lead, r, g, b)
	}
	return fmt.Sprintf("\x1b[%s;5;%dm", lead, paletteIndex(r, g, b))
}

// asciiRamp maps luminance to characters, lightest ink last.
const asciiRamp = " .:-=+*#%@"

// asciiString renders img as a plain text ramp without any escapes.
func asciiString(img *image.RGBA) string {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	var b strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := y*img.Stride + x*4
			lum := (299*int(img.Pix[p]) + 587*int(img.Pix[p+1]) + 114*int(img.Pix[p+2])) / 1000
			b.WriteByte(asciiRamp[lum*(len(asciiRamp)-1)/255])
		}
		if y < h-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
