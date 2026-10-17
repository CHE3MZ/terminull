package termimage

import (
	"bytes"
	"fmt"
	"image"
)

// sixelEncode renders img as one sixel DCS stream against the xterm 256
// palette. Layout: raster attributes, palette definitions for the used
// entries, then bands of six pixel rows where each color draws its
// columns with run-length compression. The result is deterministic.
func sixelEncode(img *image.RGBA) []byte {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	idx, used := ditherToPalette(img)

	var b bytes.Buffer
	b.WriteString("\x1bPq")
	fmt.Fprintf(&b, "\"1;1;%d;%d", w, h)
	for i := 0; i < 256; i++ {
		if !used[i] {
			continue
		}
		c := xtermPalette[i]
		fmt.Fprintf(&b, "#%d;2;%d;%d;%d", i, int(c[0])*100/255, int(c[1])*100/255, int(c[2])*100/255)
	}

	cols := make([]byte, w)
	for y0 := 0; y0 < h; y0 += 6 {
		first := true
		for ci := 0; ci < 256; ci++ {
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

// writeSixelRow emits one color's sixel row with run-length compression,
// trimming leading and trailing empty columns.
func writeSixelRow(b *bytes.Buffer, cols []byte) {
	start, end := 0, len(cols)
	for start < end && cols[start] == '?' {
		start++
	}
	for end > start && cols[end-1] == '?' {
		end--
	}
	for i := start; i < end; {
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
