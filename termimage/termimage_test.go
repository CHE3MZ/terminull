package termimage

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/CHE3MZ/terminull/style"
)

func solid(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

func setPix(img *image.RGBA, x, y int, c color.RGBA) {
	img.SetRGBA(x, y, c)
}

var (
	red   = color.RGBA{R: 255, A: 255}
	green = color.RGBA{G: 255, A: 255}
	blue  = color.RGBA{B: 255, A: 255}
	black = color.RGBA{A: 255}
	white = color.RGBA{R: 255, G: 255, B: 255, A: 255}
)

func TestResampleBoxAverage(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	setPix(img, 0, 0, black)
	setPix(img, 1, 0, white)
	setPix(img, 0, 1, white)
	setPix(img, 1, 1, black)

	got := resample(img, 1, 1)
	p := got.Pix
	if p[0] != 127 || p[1] != 127 || p[2] != 127 || p[3] != 255 {
		t.Fatalf("box average = %v, want 127,127,127,255", p[:4])
	}
}

func TestResampleBilinearRamp(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	setPix(img, 0, 0, black)
	setPix(img, 1, 0, white)

	got := resample(img, 4, 1)
	want := []uint8{0, 63, 191, 255}
	for x, w := range want {
		if got.Pix[x*4] != w {
			t.Fatalf("bilinear x=%d = %d, want %d", x, got.Pix[x*4], w)
		}
	}
}

func TestPaletteIndexKnownColors(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		r, g, b uint8
		want    int
	}{
		{0, 0, 0, 0},
		{255, 0, 0, 9},
		{0, 255, 0, 10},
		{0, 0, 255, 12},
		{255, 255, 255, 15},
		{128, 128, 128, 8},
		{8, 8, 8, 232},
		{238, 238, 238, 255},
	} {
		if got := paletteIndex(tc.r, tc.g, tc.b); got != tc.want {
			t.Fatalf("paletteIndex(%d,%d,%d) = %d, want %d", tc.r, tc.g, tc.b, got, tc.want)
		}
	}
}

func TestSixelSolidRed(t *testing.T) {
	t.Parallel()

	got := string(sixelEncode(solid(3, 2, red), false))
	want := "\x1bPq\"1;1;3;2#0;2;100;0;0#0BBB\x1b\\"
	if got != want {
		t.Fatalf("sixel = %q, want %q", got, want)
	}
}

func TestSixelRunLength(t *testing.T) {
	t.Parallel()

	got := string(sixelEncode(solid(10, 1, red), false))
	if !strings.Contains(got, "!10@") {
		t.Fatalf("sixel %q misses run-length compressed row", got)
	}
}

func TestSixelBands(t *testing.T) {
	t.Parallel()

	got := string(sixelEncode(solid(1, 7, red), false))
	if !strings.Contains(got, "#0~-#0@") {
		t.Fatalf("sixel %q misses band separator and second band", got)
	}
	if !strings.HasPrefix(got, "\x1bPq\"1;1;1;7") || !strings.HasSuffix(got, "\x1b\\") {
		t.Fatalf("sixel %q misses header or terminator", got)
	}

	if again := string(sixelEncode(solid(1, 7, red), false)); again != got {
		t.Fatal("sixel encode is not deterministic")
	}
}

func TestSixelAdaptivePalette(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 3, 1))
	setPix(img, 0, 0, red)
	setPix(img, 1, 0, green)
	setPix(img, 2, 0, blue)

	got := string(sixelEncode(img, false))
	for _, want := range []string{";2;100;0;0", ";2;0;100;0", ";2;0;0;100"} {
		if !strings.Contains(got, want) {
			t.Fatalf("sixel %q misses exact palette definition %q", got, want)
		}
	}
	if !strings.HasPrefix(got, "\x1bPq\"1;1;3;1") {
		t.Fatalf("sixel raster wrong: %q", got)
	}
	if !strings.Contains(got, "#0??@$#1?@$#2@") {
		t.Fatalf("sixel layer alignment wrong: %q", got)
	}
}

func TestSixelDitherDeterministic(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			setPix(img, x, y, color.RGBA{R: uint8(x * 31), G: uint8(y * 31), B: uint8((x * y) % 256), A: 255})
		}
	}
	opts := Options{Backend: Sixel, Width: 8, Height: 4, Dither: true}
	first, err := String(img, opts, style.NewScheme(true))
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}
	second, err := String(img, opts, style.NewScheme(true))
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}
	if first != second {
		t.Fatal("dithered render is not deterministic")
	}
	if !strings.HasPrefix(first, "\x1bPq") {
		t.Fatalf("dithered render is not sixel: %q", first)
	}
}

func TestFlattenOverBackground(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	setPix(img, 0, 0, color.RGBA{})
	setPix(img, 1, 0, color.RGBA{R: 128, A: 128})

	out := flatten(img, opaque(color.RGBA{R: 255, G: 255, B: 255, A: 255}))
	if p := out.Pix; p[0] != 255 || p[1] != 255 || p[2] != 255 || p[3] != 255 {
		t.Fatalf("transparent pixel = %v, want opaque white", p[:4])
	}
	if p := out.Pix[4:8]; p[0] != 255 || p[1] != 127 || p[2] != 127 || p[3] != 255 {
		t.Fatalf("semi-transparent pixel = %v, want 255,127,127,255", p)
	}
}

func TestSixelTransparentPixelsStayUnpainted(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	setPix(img, 0, 0, color.RGBA{})
	setPix(img, 1, 0, red)

	for _, dither := range []bool{false, true} {
		got := string(sixelEncode(img, dither))
		if !strings.HasPrefix(got, "\x1bP0;1;0q\"1;1;2;1") {
			t.Fatalf("dither=%v transparent preamble wrong: %q", dither, got)
		}
		if !strings.Contains(got, "#0?@") {
			t.Fatalf("dither=%v transparent column not left unpainted: %q", dither, got)
		}
	}

	if opaque := string(sixelEncode(solid(2, 1, red), false)); !strings.HasPrefix(opaque, "\x1bPq\"1;1;2;1") {
		t.Fatalf("opaque sixel preamble changed: %q", opaque)
	}
}

func TestSixelBackgroundFlattensOverColor(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	setPix(img, 0, 0, color.RGBA{})
	setPix(img, 1, 0, red)

	out, err := String(img, Options{Backend: Sixel, Width: 2, Height: 1, Background: white}, style.NewScheme(true))
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}
	if strings.Contains(out, "P0;1;0q") {
		t.Fatalf("flattened render kept transparent preamble: %q", out)
	}
	if !strings.Contains(out, ";2;100;100;100") {
		t.Fatalf("flattened white background missing: %q", out)
	}
}

func TestRenderCellOverrideSizesSixel(t *testing.T) {
	t.Parallel()

	out, err := String(solid(2, 2, red), Options{
		Backend: Sixel, Width: 2, Height: 1, CellWidth: 10, CellHeight: 20,
	}, style.NewScheme(true))
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}
	if !strings.Contains(out, "\"1;1;20;20") {
		t.Fatalf("override raster missing: %q", out)
	}
}

func kittyChunks(t *testing.T, seq string) []string {
	t.Helper()
	var chunks []string
	rest := seq
	for {
		i := strings.Index(rest, "\x1b_G")
		if i < 0 {
			break
		}
		rest = rest[i+3:]
		semi := strings.IndexByte(rest, ';')
		if semi < 0 {
			t.Fatalf("kitty chunk misses payload: %q", rest)
		}
		end := strings.Index(rest, "\x1b\\")
		if end < 0 {
			t.Fatalf("kitty chunk misses terminator: %q", rest)
		}
		chunks = append(chunks, rest[semi+1:end])
		rest = rest[end+2:]
	}
	return chunks
}

func TestKittySingleChunk(t *testing.T) {
	t.Parallel()

	img := solid(2, 2, red)
	seq := string(kittyEscape(img, 10, 5, 0))
	if !strings.HasPrefix(seq, "\x1b_Ga=T,f=32,s=2,v=2,q=2,c=10,r=5,m=0;") {
		t.Fatalf("kitty params wrong: %q", seq)
	}
	withID := string(kittyEscape(img, 10, 5, 7))
	if !strings.HasPrefix(withID, "\x1b_Ga=T,f=32,s=2,v=2,i=7,q=2,c=10,r=5,m=0;") {
		t.Fatalf("kitty id param wrong: %q", withID)
	}
	chunks := kittyChunks(t, seq)
	if len(chunks) != 1 {
		t.Fatalf("chunks = %d, want 1", len(chunks))
	}
	raw, err := base64.StdEncoding.DecodeString(chunks[0])
	if err != nil {
		t.Fatalf("payload decode error = %v", err)
	}
	if !bytes.Equal(raw, img.Pix) {
		t.Fatal("kitty payload is not the raw RGBA pixels")
	}
}

func TestKittyChunking(t *testing.T) {
	t.Parallel()

	img := solid(40, 40, green)
	seq := string(kittyEscape(img, 40, 20, 0))
	chunks := kittyChunks(t, seq)
	if len(chunks) != 3 {
		t.Fatalf("chunks = %d, want 3 for %d payload bytes", len(chunks), len(img.Pix))
	}
	if !strings.Contains(seq, "m=1;") || !strings.Contains(seq, "m=0;") {
		t.Fatalf("chunk continuation flags missing: %q", seq)
	}
	var joined []byte
	for _, c := range chunks {
		raw, err := base64.StdEncoding.DecodeString(c)
		if err != nil {
			t.Fatalf("chunk decode error = %v", err)
		}
		joined = append(joined, raw...)
	}
	if !bytes.Equal(joined, img.Pix) {
		t.Fatal("reassembled kitty payload differs from pixels")
	}
}

func TestKittyStraightAlpha(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 3, 1))
	copy(img.Pix, []byte{
		128, 0, 0, 128, // half-transparent red, premultiplied
		0, 64, 0, 64, // quarter-transparent green, premultiplied
		9, 9, 9, 0, // fully transparent
	})
	chunks := kittyChunks(t, string(kittyEscape(img, 1, 1, 0)))
	raw, err := base64.StdEncoding.DecodeString(chunks[0])
	if err != nil {
		t.Fatalf("payload decode error = %v", err)
	}
	want := []byte{
		255, 0, 0, 128,
		0, 255, 0, 64,
		0, 0, 0, 0,
	}
	if !bytes.Equal(raw, want) {
		t.Fatalf("straight alpha payload = %v, want %v", raw, want)
	}
}

func TestITermEscapeRoundTrip(t *testing.T) {
	t.Parallel()

	img := solid(3, 2, blue)
	data, err := encodePNG(img)
	if err != nil {
		t.Fatalf("encodePNG() error = %v", err)
	}
	seq := string(itermEscape(data, 7))
	if !strings.HasPrefix(seq, "\x1b]1337;File=inline=1;size=") {
		t.Fatalf("iterm header wrong: %q", seq)
	}
	if !strings.Contains(seq, "width=7;height=auto;preserveAspectRatio=1:") || !strings.HasSuffix(seq, "\a") {
		t.Fatalf("iterm params wrong: %q", seq)
	}
	colon := strings.LastIndexByte(seq, ':')
	payload := seq[colon+1 : len(seq)-1]
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		t.Fatalf("payload decode error = %v", err)
	}
	back, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("png.Decode() error = %v", err)
	}
	if back.Bounds().Dx() != 3 || back.Bounds().Dy() != 2 {
		t.Fatalf("decoded bounds = %v, want 3x2", back.Bounds())
	}
}

func TestHalfBlockTruecolor(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 1, 2))
	setPix(img, 0, 0, red)
	setPix(img, 0, 1, blue)

	got := halfBlockString(img, style.DepthTrue, false)
	for _, want := range []string{"\x1b[38;2;255;0;0m", "\x1b[48;2;0;0;255m", "▀", "\x1b[0m"} {
		if !strings.Contains(got, want) {
			t.Fatalf("halfblock %q misses %q", got, want)
		}
	}
}

func TestHalfBlock256(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 1, 2))
	setPix(img, 0, 0, red)
	setPix(img, 0, 1, blue)

	got := halfBlockString(img, style.Depth256, false)
	if !strings.Contains(got, "\x1b[38;5;9m") || !strings.Contains(got, "\x1b[48;5;12m") {
		t.Fatalf("halfblock 256 %q misses palette codes", got)
	}
}

func TestHalfBlockTransparency(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 1, 2))
	setPix(img, 0, 0, color.RGBA{})
	setPix(img, 0, 1, red)

	got := halfBlockString(img, style.DepthTrue, false)
	if !strings.Contains(got, "▄") || !strings.Contains(got, defaultBgSeq) {
		t.Fatalf("transparent top cell = %q, want lower block with default background", got)
	}
	if !strings.Contains(got, "\x1b[38;2;255;0;0m") {
		t.Fatalf("transparent top cell lost the opaque color: %q", got)
	}

	setPix(img, 0, 0, red)
	setPix(img, 0, 1, color.RGBA{})
	got = halfBlockString(img, style.DepthTrue, false)
	if !strings.Contains(got, "▀") || !strings.Contains(got, defaultBgSeq) {
		t.Fatalf("transparent bottom cell = %q, want upper block with default background", got)
	}

	setPix(img, 0, 0, color.RGBA{})
	got = halfBlockString(img, style.DepthTrue, false)
	if !strings.Contains(got, " ") || strings.Contains(got, "▀") || strings.Contains(got, "▄") {
		t.Fatalf("blank cell = %q, want a space", got)
	}
	if !strings.Contains(got, defaultFgSeq) || !strings.Contains(got, defaultBgSeq) {
		t.Fatalf("blank cell = %q, want default colors", got)
	}
}

func TestHalfBlockDitherDeterministic(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 16, 2))
	for x := 0; x < 16; x++ {
		v := uint8(x * 255 / 15)
		setPix(img, x, 0, color.RGBA{R: v, G: 40, B: 200, A: 255})
		setPix(img, x, 1, color.RGBA{R: 200, G: v, B: 40, A: 255})
	}
	plain := halfBlockString(img, style.Depth256, false)
	got := halfBlockString(img, style.Depth256, true)
	if got != halfBlockString(img, style.Depth256, true) {
		t.Fatal("dithered half-block is not deterministic")
	}
	if got == plain {
		t.Fatal("256-color dithering did not change the render")
	}
}

func TestHalfBlockBackgroundFlattens(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")

	img := image.NewRGBA(image.Rect(0, 0, 1, 2))
	out, err := String(img, Options{Backend: HalfBlock, Width: 1, Height: 1, Background: white}, style.NewScheme(true))
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}
	if strings.Contains(out, defaultFgSeq) || strings.Contains(out, defaultBgSeq) {
		t.Fatalf("flattened render kept default-color cells: %q", out)
	}
	if !strings.Contains(out, "\x1b[48;2;255;255;255m") {
		t.Fatalf("flattened white background missing: %q", out)
	}
}

func TestFullyTransparentImage(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")

	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	for name, backend := range map[string]Backend{
		"kitty": Kitty, "iterm2": ITerm2, "sixel": Sixel, "half": HalfBlock, "ascii": ASCII,
	} {
		out, err := String(img, Options{Backend: backend, Width: 2, Height: 1}, style.NewScheme(true))
		if err != nil {
			t.Fatalf("%s: String() error = %v", name, err)
		}
		if out == "" {
			t.Fatalf("%s: empty output for a transparent image", name)
		}
	}

	if out, _ := String(img, Options{Backend: Sixel, Width: 2, Height: 1}, style.NewScheme(true)); !strings.Contains(out, "P0;1;0q") {
		t.Fatalf("sixel transparent preamble missing: %q", out)
	}
	if out, _ := String(img, Options{Backend: ASCII, Width: 2, Height: 1}, style.NewScheme(true)); out != "  " {
		t.Fatalf("ascii = %q, want blanks", out)
	}
}

func TestASCIIRamp(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 3, 2))
	for x := 0; x < 3; x++ {
		setPix(img, x, 0, black)
		setPix(img, x, 1, white)
	}
	if got := asciiString(img, false); got != "   \n@@@" {
		t.Fatalf("ascii = %q, want dark row then bright row", got)
	}
}

func TestASCIITransparentIsBlank(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	setPix(img, 0, 0, color.RGBA{})
	setPix(img, 1, 0, white)
	if got := asciiString(img, false); got != " @" {
		t.Fatalf("ascii = %q, want transparent blank then bright", got)
	}
}

func TestASCIIDitherDeterministic(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 16, 1))
	for x := 0; x < 16; x++ {
		v := uint8(x * 255 / 15)
		setPix(img, x, 0, color.RGBA{R: v, G: v, B: v, A: 255})
	}
	plain := asciiString(img, false)
	got := asciiString(img, true)
	if got != asciiString(img, true) {
		t.Fatal("dithered ascii is not deterministic")
	}
	if got == plain {
		t.Fatal("ascii dithering did not change the render")
	}
}

func TestHalfBlockKeepsEscapesOnBuffer(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")

	img := image.NewRGBA(image.Rect(0, 0, 1, 2))
	setPix(img, 0, 0, red)
	setPix(img, 0, 1, blue)

	var buf bytes.Buffer
	if err := Render(&buf, img, Options{Backend: HalfBlock, Width: 1, Height: 1}, style.NewScheme(true)); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(buf.String(), "▀") {
		t.Fatalf("buffer render %q lost half-block escapes", buf.String())
	}
}

func TestStringSchemeDisabledIsASCII(t *testing.T) {
	t.Parallel()

	got, err := String(solid(4, 4, white), Options{Backend: HalfBlock, Width: 4, Height: 2}, style.Scheme{})
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}
	if strings.Contains(got, "\x1b") {
		t.Fatalf("scheme disabled output contains escapes: %q", got)
	}
}

func TestStringKittyAndTmuxWrap(t *testing.T) {
	t.Setenv("TMUX", "/tmp/tmux-1000/default,123,0")

	scheme := style.NewScheme(true)
	got, err := String(solid(2, 2, red), Options{Backend: Kitty, Width: 2, Height: 1}, scheme)
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}
	if !strings.HasPrefix(got, "\x1bPtmux;\x1b\x1b_G") {
		t.Fatalf("tmux passthrough missing: %q", got)
	}
	if !strings.HasSuffix(got, "\x1b\\") {
		t.Fatalf("passthrough terminator missing: %q", got)
	}
}

func TestStringSixelStaysRawInTmux(t *testing.T) {
	t.Setenv("TMUX", "/tmp/tmux-1000/default,123,0")

	got, err := String(solid(2, 2, red), Options{Backend: Sixel, Width: 1, Height: 1}, style.NewScheme(true))
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}
	if strings.Contains(got, "\x1bPtmux;") {
		t.Fatalf("sixel must bypass tmux passthrough: %q", got)
	}
	if !strings.HasPrefix(got, "\x1bP") {
		t.Fatalf("sixel stream malformed: %q", got)
	}
}

func TestRenderRejectsEmptyImage(t *testing.T) {
	t.Parallel()

	if _, err := String(image.NewRGBA(image.Rect(0, 0, 0, 0)), Options{}, style.NewScheme(true)); err == nil {
		t.Fatal("String() accepted an empty image")
	}
}

func TestTargetSize(t *testing.T) {
	t.Parallel()

	img := solid(100, 50, red)
	cols, rows := targetSize(img, Options{Width: 40}, 8, 16)
	if cols != 40 || rows != 10 {
		t.Fatalf("targetSize = %d,%d, want 40,10", cols, rows)
	}
	cols, rows = targetSize(img, Options{Width: 40, Height: 4}, 8, 16)
	if cols != 40 || rows != 4 {
		t.Fatalf("targetSize with height = %d,%d, want 40,4", cols, rows)
	}
}

func TestBackendFromEnv(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  map[string]string
		want Backend
	}{
		{"kitty id", map[string]string{"KITTY_WINDOW_ID": "1"}, Kitty},
		{"ghostty term", map[string]string{"TERM": "xterm-ghostty"}, Kitty},
		{"wezterm", map[string]string{"TERM_PROGRAM": "WezTerm"}, ITerm2},
		{"konsole 22.04", map[string]string{"KONSOLE_VERSION": "220400"}, Kitty},
		{"konsole newer", map[string]string{"KONSOLE_VERSION": "240200"}, Kitty},
		{"iterm2", map[string]string{"TERM_PROGRAM": "iTerm.app"}, ITerm2},
		{"foot", map[string]string{"TERM": "foot"}, Sixel},
		{"mlterm", map[string]string{"TERM": "mlterm"}, Sixel},
	} {
		getenv := func(k string) string { return tc.env[k] }
		if got, ok := backendFromEnv(getenv); !ok || got != tc.want {
			t.Fatalf("%s: backendFromEnv = %v, %v; want %v", tc.name, got, ok, tc.want)
		}
	}
	for _, v := range []string{"211200", "220400x"} {
		getenv := func(k string) string {
			if k == "KONSOLE_VERSION" {
				return v
			}
			return ""
		}
		if b, ok := backendFromEnv(getenv); ok {
			t.Fatalf("KONSOLE_VERSION=%s detected %v, want fallback", v, b)
		}
	}
	if b, ok := backendFromEnv(func(string) string { return "" }); ok {
		t.Fatalf("empty env detected %v, want fallback", b)
	}
}

func TestHasSixelDA(t *testing.T) {
	t.Parallel()

	if !hasSixelDA([]byte("\x1b[?62;4;6c")) {
		t.Fatal("DA1 with attribute 4 not recognized")
	}
	if hasSixelDA([]byte("\x1b[?62;1;6c")) {
		t.Fatal("DA1 without attribute 4 recognized as sixel")
	}
	if hasSixelDA([]byte("garbage")) {
		t.Fatal("garbage recognized as sixel")
	}
}
