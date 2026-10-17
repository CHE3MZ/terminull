package termimage

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"
)

// kittyChunk is the parameter payload limit kitty accepts per escape.
const kittyChunk = 4096

// kittyEscape builds the kitty graphics escape stream for img: the raw
// RGBA pixels, base64'd and chunked, transmitted and displayed at the
// given cell size under a stable image id so re-renders replace.
func kittyEscape(img *image.RGBA, cols, rows int) []byte {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	b64 := base64.StdEncoding.EncodeToString(img.Pix)

	var b bytes.Buffer
	for off, part := 0, 0; off < len(b64); part++ {
		end := off + kittyChunk
		if end > len(b64) {
			end = len(b64)
		}
		more := 0
		if end < len(b64) {
			more = 1
		}
		if part == 0 {
			fmt.Fprintf(&b, "\x1b_Ga=T,f=32,s=%d,v=%d,i=1,c=%d,r=%d,m=%d;%s\x1b\\", w, h, cols, rows, more, b64[off:end])
		} else {
			fmt.Fprintf(&b, "\x1b_Gm=%d;%s\x1b\\", more, b64[off:end])
		}
		off = end
	}
	return b.Bytes()
}

// encodePNG compresses an image to PNG bytes. RGBA pixels are
// premultiplied, PNG stores straight alpha, so they convert first;
// other image types pass through.
func encodePNG(img image.Image) ([]byte, error) {
	if src, ok := img.(*image.RGBA); ok {
		straight := image.NewNRGBA(src.Bounds())
		draw.Draw(straight, straight.Bounds(), src, src.Bounds().Min, draw.Src)
		img = straight
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// itermEscape builds the iTerm2 OSC 1337 inline image sequence: a PNG,
// base64'd, sized to a cell width with aspect preserved by the terminal.
func itermEscape(pngData []byte, cols int) []byte {
	payload := base64.StdEncoding.EncodeToString(pngData)
	return fmt.Appendf(nil, "\x1b]1337;File=inline=1;size=%d;width=%d;height=auto;preserveAspectRatio=1:%s\a", len(pngData), cols, payload)
}
