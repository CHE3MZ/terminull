package termimage

import (
	"image"
	"image/color"
	"math"
)

// mapToPalette converts img to palette indexes, optionally with
// Floyd-Steinberg error diffusion. It returns the index buffer and the
// set of palette entries actually used. The plain path memoizes nearest
// lookups per distinct color; rendered images downscale to few colors.
func mapToPalette(img *image.RGBA, pal []color.RGBA, dither bool) ([]uint8, []bool) {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	idx := make([]uint8, w*h)
	used := make([]bool, len(pal))

	if !dither {
		cache := make(map[int]int)
		for y := 0; y < h; y++ {
			row := y * img.Stride
			for x := 0; x < w; x++ {
				p := row + x*4
				key := pack(img.Pix[p], img.Pix[p+1], img.Pix[p+2])
				n, ok := cache[key]
				if !ok {
					n = nearestIn(pal, img.Pix[p], img.Pix[p+1], img.Pix[p+2])
					cache[key] = n
				}
				used[n] = true
				idx[y*w+x] = uint8(n) // #nosec G115 -- palette indexes stay under 256.
			}
		}
		return idx, used
	}

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
			n := nearestIn(pal, r, g, b)
			used[n] = true
			idx[i] = uint8(n) // #nosec G115 -- palette indexes stay under 256.

			c := pal[n]
			er := buf[i][0] - float64(c.R)
			eg := buf[i][1] - float64(c.G)
			eb := buf[i][2] - float64(c.B)
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
