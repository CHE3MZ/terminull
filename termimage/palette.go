package termimage

import (
	"image"
	"image/color"
	"sort"
	"sync"
)

// xtermPalette is the standard 256-color xterm palette: 16 base colors,
// a 6x6x6 RGB cube, and 24 grays. Sixel and 256-color ANSI output both
// snap to it, keeping renders deterministic.
var xtermPalette = buildXtermPalette()

func buildXtermPalette() [256][3]uint8 {
	var p [256][3]uint8
	base := [16][3]uint8{
		{0, 0, 0}, {128, 0, 0}, {0, 128, 0}, {128, 128, 0},
		{0, 0, 128}, {128, 0, 128}, {0, 128, 128}, {192, 192, 192},
		{128, 128, 128}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
		{0, 0, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
	}
	copy(p[:16], base[:])

	levels := [6]uint8{0, 95, 135, 175, 215, 255}
	i := 16
	for _, r := range levels {
		for _, g := range levels {
			for _, b := range levels {
				p[i] = [3]uint8{r, g, b}
				i++
			}
		}
	}
	for j := 0; j < 24; j++ {
		v := uint8(8 + 10*j)
		p[232+j] = [3]uint8{v, v, v}
	}
	return p
}

var (
	// paletteExact maps exact palette colors to their first index, so
	// solid fills and accents never drift through the approximate LUT.
	paletteExact = buildExactIndex()
	paletteLUT   [32768]int16
	lutOnce      sync.Once
)

// xtermRGBA returns the xterm palette as color values for quantization
// paths that need a palette slice instead of the fixed table.
func xtermRGBA() []color.RGBA {
	xtermRGBAOnce.Do(func() {
		xtermRGBASlice = make([]color.RGBA, len(xtermPalette))
		for i, c := range xtermPalette {
			xtermRGBASlice[i] = color.RGBA{R: c[0], G: c[1], B: c[2], A: 255}
		}
	})
	return xtermRGBASlice
}

var (
	xtermRGBASlice []color.RGBA
	xtermRGBAOnce  sync.Once
)

func buildExactIndex() map[[3]uint8]int16 {
	m := make(map[[3]uint8]int16, 256)
	for i, c := range xtermPalette {
		if _, ok := m[c]; !ok {
			m[c] = int16(i)
		}
	}
	return m
}

// paletteIndex returns the xterm palette index closest to rgb in squared
// distance: exact hits first, then a memoized 5-bit lookup table.
func paletteIndex(r, g, b uint8) int {
	if i, ok := paletteExact[[3]uint8{r, g, b}]; ok {
		return int(i)
	}
	lutOnce.Do(func() {
		for i := range paletteLUT {
			ri := uint8((i >> 10) & 31)
			gi := uint8((i >> 5) & 31)
			bi := uint8(i & 31)
			paletteLUT[i] = int16(bruteIndex(ri<<3|ri>>2, gi<<3|gi>>2, bi<<3|bi>>2)) // #nosec G115 -- bruteIndex returns 0..255.
		}
	})
	key := int(r>>3)<<10 | int(g>>3)<<5 | int(b>>3)
	return int(paletteLUT[key])
}

// bruteIndex scans all 256 palette entries.
func bruteIndex(r, g, b uint8) int {
	best, bestDist := 0, 1<<30
	for i, c := range xtermPalette {
		dr := int(r) - int(c[0])
		dg := int(g) - int(c[1])
		db := int(b) - int(c[2])
		d := dr*dr + dg*dg + db*db
		if d < bestDist {
			best, bestDist = i, d
		}
	}
	return best
}

// paletteEntry is one histogram cell of the 15-bit color space.
type paletteEntry struct {
	r, g, b uint8
	n       uint64
}

// alphaFloor is the alpha cutoff below which a pixel counts as
// transparent: it is left unpainted by sixel instead of snapping to a
// color, and it never votes in the adaptive palette. Terminals composite
// alpha binarily, so a threshold beats pretending to blend.
const alphaFloor = 128

// adaptivePalette builds up to max colors from the image itself with a
// weighted median cut over a 15-bit histogram, so gradients follow the
// actual picture instead of the coarse fixed xterm cube.
func adaptivePalette(img *image.RGBA, max int) []color.RGBA {
	type bucket struct {
		n          uint64
		rs, gs, bs uint64
	}
	hist := make([]bucket, 32768)
	pix := img.Pix
	for i := 0; i < len(pix); i += 4 {
		if pix[i+3] < alphaFloor {
			continue
		}
		key := int(pix[i]>>3)<<10 | int(pix[i+1]>>3)<<5 | int(pix[i+2]>>3)
		h := &hist[key]
		h.n++
		h.rs += uint64(pix[i])
		h.gs += uint64(pix[i+1])
		h.bs += uint64(pix[i+2])
	}

	var entries []paletteEntry
	for _, h := range hist {
		if h.n == 0 {
			continue
		}
		entries = append(entries, paletteEntry{
			r: uint8(h.rs / h.n), // #nosec G115 -- average of uint8 samples is bounded by 255.
			g: uint8(h.gs / h.n), // #nosec G115 -- same bound as above.
			b: uint8(h.bs / h.n), // #nosec G115 -- same bound as above.
			n: h.n,
		})
	}
	if len(entries) == 0 {
		// Fully transparent image: no color is ever painted, but the
		// nearest-color math needs one entry to stay in bounds.
		return []color.RGBA{{A: 255}}
	}
	if len(entries) <= max {
		pal := make([]color.RGBA, len(entries))
		for i, e := range entries {
			pal[i] = color.RGBA{R: e.r, G: e.g, B: e.b, A: 255}
		}
		return pal
	}

	type box struct{ lo, hi int }
	boxes := []box{{0, len(entries)}}
	channel := func(e paletteEntry, c int) int {
		switch c {
		case 0:
			return int(e.r)
		case 1:
			return int(e.g)
		default:
			return int(e.b)
		}
	}
	// widest reports the channel with the largest spread and its value.
	widest := func(b box) (int, int) {
		ch, spread := 0, -1
		for c := 0; c < 3; c++ {
			lo, hi := 255, 0
			for i := b.lo; i < b.hi; i++ {
				v := channel(entries[i], c)
				if v < lo {
					lo = v
				}
				if v > hi {
					hi = v
				}
			}
			if hi-lo > spread {
				ch, spread = c, hi-lo
			}
		}
		return ch, spread
	}

	for len(boxes) < max {
		bi, bestSpread := -1, 0
		for i, b := range boxes {
			if b.hi-b.lo < 2 {
				continue
			}
			if _, spread := widest(b); spread > bestSpread {
				bi, bestSpread = i, spread
			}
		}
		if bi < 0 {
			break
		}
		b := boxes[bi]
		ch, _ := widest(b)
		seg := entries[b.lo:b.hi]
		sort.Slice(seg, func(i, j int) bool {
			for c := 0; c < 3; c++ {
				ci := channel(seg[i], (ch+c)%3)
				cj := channel(seg[j], (ch+c)%3)
				if ci != cj {
					return ci < cj
				}
			}
			return false
		})
		var total uint64
		for _, e := range seg {
			total += e.n
		}
		var cum uint64
		split := b.lo + 1
		for i := b.lo; i < b.hi-1; i++ {
			cum += entries[i].n
			split = i + 1
			if cum*2 >= total {
				break
			}
		}
		boxes[bi] = box{b.lo, split}
		boxes = append(boxes, box{split, b.hi})
	}

	pal := make([]color.RGBA, len(boxes))
	for i, b := range boxes {
		var n, rs, gs, bs uint64
		for j := b.lo; j < b.hi; j++ {
			e := entries[j]
			n += e.n
			rs += uint64(e.r) * e.n
			gs += uint64(e.g) * e.n
			bs += uint64(e.b) * e.n
		}
		pal[i] = color.RGBA{
			R: uint8(rs / n), // #nosec G115 -- weighted average of uint8 samples is bounded by 255.
			G: uint8(gs / n), // #nosec G115 -- same bound as above.
			B: uint8(bs / n), // #nosec G115 -- same bound as above.
			A: 255,
		}
	}
	return pal
}

// nearestIn scans the palette for the closest color in squared distance.
func nearestIn(pal []color.RGBA, r, g, b uint8) int {
	best, bestDist := 0, 1<<30
	for i, c := range pal {
		dr := int(r) - int(c.R)
		dg := int(g) - int(c.G)
		db := int(b) - int(c.B)
		d := dr*dr + dg*dg + db*db
		if d < bestDist {
			best, bestDist = i, d
		}
	}
	return best
}
