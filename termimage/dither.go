package termimage

import (
	"image"
	"math"
)

// ditherToPalette converts img to xterm palette indexes with
// Floyd–Steinberg error diffusion, returning the index buffer and the
// set of palette entries actually used. Working values are premultiplied
// like the source, so transparent pixels darken toward black, which is
// what sixel shows for them.
func ditherToPalette(img *image.RGBA) ([]uint8, [256]bool) {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	idx := make([]uint8, w*h)
	var used [256]bool

	buf := make([][3]float64, w*h)
	for y := 0; y < h; y++ {
		row := y * img.Stride
		for x := 0; x < w; x++ {
			p := row + x*4
			buf[y*w+x] = [3]float64{
				float64(img.Pix[p]),
				float64(img.Pix[p+1]),
				float64(img.Pix[p+2]),
			}
		}
	}

	spread := func(x, y int, er, eg, eb, factor float64) {
		if x < 0 || x >= w || y < 0 || y >= h {
			return
		}
		i := y*w + x
		buf[i][0] += er * factor
		buf[i][1] += eg * factor
		buf[i][2] += eb * factor
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			r := clampByte(buf[i][0])
			g := clampByte(buf[i][1])
			b := clampByte(buf[i][2])
			n := paletteIndex(r, g, b)
			used[n] = true
			idx[i] = uint8(n) // #nosec G115 -- paletteIndex returns 0..255.

			c := xtermPalette[n]
			er := buf[i][0] - float64(c[0])
			eg := buf[i][1] - float64(c[1])
			eb := buf[i][2] - float64(c[2])
			spread(x+1, y, er, eg, eb, 7.0/16)
			spread(x-1, y+1, er, eg, eb, 3.0/16)
			spread(x, y+1, er, eg, eb, 5.0/16)
			spread(x+1, y+1, er, eg, eb, 1.0/16)
		}
	}
	return idx, used
}

func clampByte(v float64) uint8 {
	n := math.Round(v)
	switch {
	case n < 0:
		return 0
	case n > 255:
		return 255
	default:
		return uint8(n)
	}
}
