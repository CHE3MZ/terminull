//go:build !windows

package termimage

// sixelUnits returns the per-cell pixel units for sixel sizing. POSIX
// terminals place one sixel pixel per device pixel, so the real cell
// size applies directly.
func sixelUnits(cw, ch int) (int, int) { return cw, ch }
