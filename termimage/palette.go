package termimage

import "sync"

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
			paletteLUT[i] = int16(bruteIndex(ri<<3|ri>>2, gi<<3|gi>>2, bi<<3|bi>>2))
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
