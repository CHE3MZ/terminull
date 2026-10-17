package termimage

import (
	"image"
	"image/draw"
	"math"
)

// toRGBA normalizes any decoded image into a tightly packed RGBA copy.
func toRGBA(src image.Image) *image.RGBA {
	b := src.Bounds()
	if r, ok := src.(*image.RGBA); ok && b.Min == (image.Point{}) && r.Stride == b.Dx()*4 {
		return r
	}
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

// resample scales src to w by h pixels. Shrinking averages source
// footprints (box filter); growing interpolates bilinearly. Channel
// math stays premultiplied, matching the RGBA representation.
func resample(src *image.RGBA, w, h int) *image.RGBA {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	switch {
	case sw == w && sh == h:
		copy(dst.Pix, src.Pix)
	case w <= sw && h <= sh:
		boxResample(src, dst)
	default:
		bilinearResample(src, dst)
	}
	return dst
}

// boxResample averages the source footprint of every destination pixel.
func boxResample(src, dst *image.RGBA) {
	sw := src.Bounds().Dx()
	sh := src.Bounds().Dy()
	dw := dst.Bounds().Dx()
	dh := dst.Bounds().Dy()
	xs := float64(sw) / float64(dw)
	ys := float64(sh) / float64(dh)

	for dy := 0; dy < dh; dy++ {
		y0 := int(float64(dy) * ys)
		y1 := int(float64(dy+1) * ys)
		if y1 <= y0 {
			y1 = y0 + 1
		}
		if y1 > sh {
			y1 = sh
		}
		for dx := 0; dx < dw; dx++ {
			x0 := int(float64(dx) * xs)
			x1 := int(float64(dx+1) * xs)
			if x1 <= x0 {
				x1 = x0 + 1
			}
			if x1 > sw {
				x1 = sw
			}

			var r, g, b, a, n uint64
			for y := y0; y < y1; y++ {
				off := y*src.Stride + x0*4
				for x := x0; x < x1; x++ {
					r += uint64(src.Pix[off])
					g += uint64(src.Pix[off+1])
					b += uint64(src.Pix[off+2])
					a += uint64(src.Pix[off+3])
					off += 4
					n++
				}
			}
			if n == 0 {
				n = 1
			}
			o := dy*dst.Stride + dx*4
			dst.Pix[o] = uint8(r / n)
			dst.Pix[o+1] = uint8(g / n)
			dst.Pix[o+2] = uint8(b / n)
			dst.Pix[o+3] = uint8(a / n)
		}
	}
}

// bilinearResample interpolates the four source neighbors around every
// destination pixel center.
func bilinearResample(src, dst *image.RGBA) {
	sw := src.Bounds().Dx()
	sh := src.Bounds().Dy()
	dw := dst.Bounds().Dx()
	dh := dst.Bounds().Dy()

	for dy := 0; dy < dh; dy++ {
		fy := (float64(dy)+0.5)*float64(sh)/float64(dh) - 0.5
		y0 := int(math.Floor(fy))
		ty := fy - float64(y0)
		y1 := y0 + 1
		if y0 < 0 {
			y0, ty = 0, 0
		}
		if y1 >= sh {
			y1 = sh - 1
		}
		if y0 >= sh {
			y0 = sh - 1
		}

		for dx := 0; dx < dw; dx++ {
			fx := (float64(dx)+0.5)*float64(sw)/float64(dw) - 0.5
			x0 := int(math.Floor(fx))
			tx := fx - float64(x0)
			x1 := x0 + 1
			if x0 < 0 {
				x0, tx = 0, 0
			}
			if x1 >= sw {
				x1 = sw - 1
			}
			if x0 >= sw {
				x0 = sw - 1
			}

			o := dy*dst.Stride + dx*4
			p00 := y0*src.Stride + x0*4
			p01 := y0*src.Stride + x1*4
			p10 := y1*src.Stride + x0*4
			p11 := y1*src.Stride + x1*4
			for c := 0; c < 4; c++ {
				top := float64(src.Pix[p00+c])*(1-tx) + float64(src.Pix[p01+c])*tx
				bot := float64(src.Pix[p10+c])*(1-tx) + float64(src.Pix[p11+c])*tx
				dst.Pix[o+c] = uint8(top*(1-ty) + bot*ty)
			}
		}
	}
}
