// Package termimage renders images to terminals.
//
// Backends cover the modern image protocols — kitty graphics (kitty,
// Ghostty, Konsole), iTerm2 inline images (iTerm2, WezTerm),
// and sixel (Windows Terminal, foot, mlterm, xterm) — plus ANSI tiers
// that work everywhere: truecolor half blocks and a plain ASCII ramp.
//
// Everything is written in-repo on the standard library alone: no cgo,
// no new module dependencies, no copied protocol code. Decoding uses
// image/png, image/jpeg, and image/gif; base64 and PNG encoding are
// stdlib; scaling, adaptive palette quantization, and optional dithering
// are ours and deterministic, so protocol output is golden-testable.
//
// Backend selection is automatic by default: environment signals pick
// kitty and iTerm2 terminals, a DA1 probe asks about sixel on real
// consoles, and the ANSI tiers catch everything else. An explicit
// Backend always wins. A disabled Scheme degrades Auto to ASCII, so
// piped output stays plain text.
package termimage

import (
	"errors"
	"image"
	"image/color"
	"io"
	"math"
	"os"
	"strings"

	"github.com/CHE3MZ/terminull/termcore"
	"github.com/CHE3MZ/terminull/termstyle"
)

// Backend selects an image protocol or ANSI tier.
type Backend int

// Backends from richest to plainest.
const (
	// Auto detects the best backend for the current terminal.
	Auto Backend = iota
	// Kitty is the kitty graphics protocol.
	Kitty
	// ITerm2 is the iTerm2 OSC 1337 inline image protocol.
	ITerm2
	// Sixel is DEC sixel (DCS).
	Sixel
	// HalfBlock paints two pixel rows per cell with upper half blocks.
	HalfBlock
	// ASCII paints luminance through a character ramp.
	ASCII
)

// String names the backend.
func (b Backend) String() string {
	switch b {
	case Auto:
		return "auto"
	case Kitty:
		return "kitty"
	case ITerm2:
		return "iterm2"
	case Sixel:
		return "sixel"
	case HalfBlock:
		return "half"
	case ASCII:
		return "ascii"
	default:
		return "unknown"
	}
}

// Options tunes a render.
type Options struct {
	// Backend forces a protocol; zero Auto detects from the terminal.
	Backend Backend
	// Width caps the render in cells; zero uses the terminal width.
	Width int
	// Height caps the render in cells; zero keeps the image aspect.
	Height int
	// Dither enables Floyd-Steinberg error diffusion wherever a backend
	// quantizes: the sixel adaptive palette, 256-color half blocks, and
	// the ascii ramp. Exact-pixel paths (kitty, iterm2, truecolor half
	// blocks) send full color and are unaffected.
	Dither bool
	// Background flattens transparency when it has alpha: every backend
	// blends the image over this color. The zero value keeps
	// transparency instead — kitty and iterm2 send native alpha, sixel
	// leaves those pixels unpainted, and half blocks and ascii render
	// them as blanks, all letting the terminal background show through.
	Background color.RGBA
	// CellWidth and CellHeight override the per-cell pixel estimate
	// used for sizing; zero queries the terminal.
	CellWidth  int
	CellHeight int
	// ImageID is the kitty graphics image id. Zero sends every render
	// as a fresh anonymous image, so sequential renders coexist; a
	// non-zero id makes a retransmit replace earlier renders of that
	// id in place, which suits redrawing a widget or an animation.
	ImageID uint32
}

// String renders img and returns the escape stream or text.
func String(img image.Image, opts Options, scheme termstyle.Scheme) (string, error) {
	var b strings.Builder
	if err := Render(&b, img, opts, scheme); err != nil {
		return "", err
	}
	return b.String(), nil
}

// Render writes img to w in the chosen backend. Inside tmux, kitty and
// iTerm2 streams are wrapped for passthrough (the allow-passthrough
// option must be on); sixel goes raw because tmux 3.4+ renders it
// natively.
func Render(w io.Writer, img image.Image, opts Options, scheme termstyle.Scheme) error {
	if img == nil {
		return errors.New("termimage: nil image")
	}
	if img.Bounds().Empty() {
		return errors.New("termimage: empty image")
	}
	backend := opts.Backend
	depth := termcore.ColorDepth()
	if backend == Auto {
		backend = Detect()
		if !scheme.Enabled && backend != ASCII {
			backend = ASCII
		}
	}
	if backend == HalfBlock && (!scheme.Enabled || depth == termcore.Depth16) {
		backend = ASCII
	}

	rgba := toRGBA(img)
	if opts.Background.A != 0 && (backend == Sixel || backend == HalfBlock || backend == ASCII) {
		rgba = flatten(rgba, opts.Background)
	}
	cw, ch := cellPixels()
	if opts.CellWidth > 0 {
		cw = opts.CellWidth
	}
	if opts.CellHeight > 0 {
		ch = opts.CellHeight
	}
	cols, rows := targetSize(rgba, opts, cw, ch)

	switch backend {
	case Kitty:
		scaled := resample(rgba, cols*cw, rows*ch)
		_, err := w.Write(tmuxWrap(kittyEscape(scaled, cols, rows, opts.ImageID)))
		return err
	case ITerm2:
		scaled := resample(rgba, cols*cw, rows*ch)
		png, err := encodePNG(scaled)
		if err != nil {
			return err
		}
		_, err = w.Write(tmuxWrap(itermEscape(png, cols)))
		return err
	case Sixel:
		uw, uh := sixelUnits(cw, ch)
		if opts.CellWidth > 0 {
			uw = opts.CellWidth
		}
		if opts.CellHeight > 0 {
			uh = opts.CellHeight
		}
		scaled := resample(rgba, cols*uw, rows*uh)
		// No passthrough wrap: tmux 3.4+ implements sixel itself, so the
		// raw DCS renders (and redraws on resize) without the
		// allow-passthrough option; older tmux discards it either way.
		_, err := w.Write(sixelEncode(scaled, opts.Dither))
		return err
	case HalfBlock:
		if !consoleReady(w) {
			return writeASCII(w, rgba, cols, rows, opts.Dither)
		}
		_, err := io.WriteString(w, halfBlockString(resample(rgba, cols, rows*2), depth, opts.Dither))
		return err
	case ASCII:
		return writeASCII(w, rgba, cols, rows, opts.Dither)
	default:
		return errors.New("termimage: unknown backend")
	}
}

func writeASCII(w io.Writer, img *image.RGBA, cols, rows int, dither bool) error {
	_, err := io.WriteString(w, asciiString(resample(img, cols, rows), dither))
	return err
}

// consoleReady prepares a real console for VT escapes and reports whether
// escapes can be shown. Pipes, files, and buffers need no preparation and
// keep their escapes; a console that refuses VT asks for the ASCII tier.
func consoleReady(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok || !termcore.IsTerminal(f) {
		return true
	}
	return termcore.EnsureConsole() == nil
}

// targetSize resolves the cell footprint; the cell pixel estimate cw by
// ch was measured once and keeps the image aspect honest per backend.
func targetSize(img *image.RGBA, opts Options, cw, ch int) (cols, rows int) {
	cols = opts.Width
	if cols <= 0 {
		cols = terminalColumns()
	}
	if cols < 1 {
		cols = 1
	}
	b := img.Bounds()
	iw, ih := b.Dx(), b.Dy()
	rows = opts.Height
	if rows <= 0 && iw > 0 && ih > 0 {
		rows = int(math.Round(float64(cols*cw) * float64(ih) / (float64(iw) * float64(ch))))
	}
	if rows < 1 {
		rows = 1
	}
	return cols, rows
}

func terminalColumns() int {
	if termcore.IsTerminal(os.Stdout) {
		if n := termcore.TerminalWidth(os.Stdout); n > 0 {
			return n
		}
	}
	return 80
}

// tmuxWrap routes an escape stream through tmux passthrough when running
// inside tmux, doubling inner escapes per the DCS protocol. Only for
// protocols tmux cannot host itself (kitty, iTerm2); sixel is never
// wrapped. Requires the allow-passthrough option, which defaults off,
// and tmux drops sequences much over a megabyte.
func tmuxWrap(seq []byte) []byte {
	if os.Getenv("TMUX") == "" {
		return seq
	}
	out := make([]byte, 0, len(seq)+32)
	out = append(out, "\x1bPtmux;"...)
	for _, b := range seq {
		if b == 0x1b {
			out = append(out, 0x1b)
		}
		out = append(out, b)
	}
	out = append(out, "\x1b\\"...)
	return out
}
