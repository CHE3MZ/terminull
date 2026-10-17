package termimage

import (
	"bytes"
	"fmt"
	"image"
)

// sixelEncode renders img as one sixel DCS stream against a palette
// adapted to the image. Layout: raster attributes, palette definitions
// for the used entries, then bands of six pixel rows where each color
// draws its columns with run-length compression. Deterministic.
func sixelEncode(img *image.RGBA, dither bool) []byte {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	pal := adaptivePalette(img, 256)
	idx, used := mapToPalette(img, pal, dither)

	pct := func(v uint8) int { return (int(v)*100 + 127) / 255 }
	var b bytes.Buffer
	b.WriteString("\x1bPq")
	fmt.Fprintf(&b, "\"1;1;%d;%d", w, h)
	for i, c := range pal {
		if !used[i] {
			continue
		}
		fmt.Fprintf(&b, "#%d;2;%d;%d;%d", i, pct(c.R), pct(c.G), pct(c.B))
	}

	cols := make([]byte, w)
	for y0 := 0; y0 < h; y0 += 6 {
		first := true
		for ci := range pal {
			if !used[ci] {
				continue
			}
			any := false
			for x := 0; x < w; x++ {
				var v byte
				for dy := 0; dy < 6; dy++ {
					y := y0 + dy
					if y >= h {
						break
					}
					if int(idx[y*w+x]) == ci {
						v |= 1 << dy
						any = true
					}
				}
				cols[x] = 63 + v
			}
			if !any {
				continue
			}
			if !first {
				b.WriteByte('$')
			}
			first = false
			fmt.Fprintf(&b, "#%d", ci)
			writeSixelRow(&b, cols)
		}
		if y0+6 < h {
			b.WriteByte('-')
		}
	}
	b.WriteString("\x1b\\")
	return b.Bytes()
}

// writeSixelRow emits one color's sixel row with run-length compression.
// Every column is emitted from zero because '?' still advances the
// horizontal position; trimming leading columns would shift the layer
// left over the layers beneath it. Only trailing empty columns drop.
func writeSixelRow(b *bytes.Buffer, cols []byte) {
	end := len(cols)
	for end > 0 && cols[end-1] == '?' {
		end--
	}
	for i := 0; i < end; {
		j := i
		for j < end && cols[j] == cols[i] {
			j++
		}
		run := j - i
		if run >= 4 {
			fmt.Fprintf(b, "!%d%c", run, cols[i])
		} else {
			for k := 0; k < run; k++ {
				b.WriteByte(cols[i])
			}
		}
		i = j
	}
}
