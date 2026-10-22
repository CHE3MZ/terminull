//go:build windows

package termimage

// sixelUnits returns the DEC VT240 reference cell of 10x20 pixels:
// Windows consoles (Windows Terminal and conhost) size sixel pixels
// with that fixed assumption instead of the real font cell, so images
// declare their raster in these units to keep aspect on any font.
func sixelUnits(_, _ int) (int, int) { return 10, 20 }
