package renderer

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"
	"time"
)

func RenderPreview(snapshot Snapshot, src image.Image, now time.Time) *image.RGBA {
	width := snapshot.Display.Width
	height := snapshot.Display.Height
	if width <= 0 {
		width = 1920
	}
	if height <= 0 {
		height = 1080
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.Black), image.Point{}, draw.Src)
	drawFitted(dst, src, nonEmpty(snapshot.Playback.Fit, "contain"))
	drawOverlay(dst, strings.ToUpper(nonEmpty(snapshot.Status, "ready"))+" "+now.Format("15:04"))
	return dst
}

func drawFitted(dst *image.RGBA, src image.Image, fit string) {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	dstW := dst.Bounds().Dx()
	dstH := dst.Bounds().Dy()
	if srcW <= 0 || srcH <= 0 || dstW <= 0 || dstH <= 0 {
		return
	}
	scaleX := float64(dstW) / float64(srcW)
	scaleY := float64(dstH) / float64(srcH)
	scale := math.Min(scaleX, scaleY)
	if fit == "cover" {
		scale = math.Max(scaleX, scaleY)
	}
	w := int(math.Round(float64(srcW) * scale))
	h := int(math.Round(float64(srcH) * scale))
	x0 := (dstW - w) / 2
	y0 := (dstH - h) / 2
	for y := 0; y < dstH; y++ {
		srcY := int(float64(y-y0)/scale) + srcBounds.Min.Y
		if srcY < srcBounds.Min.Y || srcY >= srcBounds.Max.Y {
			continue
		}
		for x := 0; x < dstW; x++ {
			srcX := int(float64(x-x0)/scale) + srcBounds.Min.X
			if srcX < srcBounds.Min.X || srcX >= srcBounds.Max.X {
				continue
			}
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
}

func drawOverlay(dst *image.RGBA, text string) {
	x := 16
	y := dst.Bounds().Dy() - 32
	if y < 8 {
		y = 8
	}
	padding := 8
	textW := len(text) * 6
	bg := image.Rect(x-padding, y-padding, x+textW+padding, y+7+padding)
	draw.Draw(dst, bg, image.NewUniform(color.RGBA{A: 150}), image.Point{}, draw.Over)
	for _, r := range text {
		drawGlyph(dst, x, y, r, color.RGBA{R: 245, G: 245, B: 245, A: 255})
		x += 6
	}
}

func drawGlyph(dst *image.RGBA, x, y int, r rune, c color.Color) {
	rows, ok := glyphs[r]
	if !ok {
		rows = glyphs[' ']
	}
	for row, bits := range rows {
		for col := 0; col < 5; col++ {
			if bits&(1<<(4-col)) != 0 {
				dst.Set(x+col, y+row, c)
			}
		}
	}
}

var glyphs = map[rune][7]byte{
	' ': {0, 0, 0, 0, 0, 0, 0},
	'-': {0, 0, 0, 31, 0, 0, 0},
	':': {0, 4, 4, 0, 4, 4, 0},
	'0': {14, 17, 19, 21, 25, 17, 14},
	'1': {4, 12, 4, 4, 4, 4, 14},
	'2': {14, 17, 1, 2, 4, 8, 31},
	'3': {30, 1, 1, 14, 1, 1, 30},
	'4': {2, 6, 10, 18, 31, 2, 2},
	'5': {31, 16, 30, 1, 1, 17, 14},
	'6': {6, 8, 16, 30, 17, 17, 14},
	'7': {31, 1, 2, 4, 8, 8, 8},
	'8': {14, 17, 17, 14, 17, 17, 14},
	'9': {14, 17, 17, 15, 1, 2, 12},
	'A': {14, 17, 17, 31, 17, 17, 17},
	'D': {30, 17, 17, 17, 17, 17, 30},
	'E': {31, 16, 16, 30, 16, 16, 31},
	'G': {14, 17, 16, 23, 17, 17, 14},
	'M': {17, 27, 21, 21, 17, 17, 17},
	'O': {14, 17, 17, 17, 17, 17, 14},
	'P': {30, 17, 17, 30, 16, 16, 16},
	'R': {30, 17, 17, 30, 20, 18, 17},
	'T': {31, 4, 4, 4, 4, 4, 4},
	'Y': {17, 17, 10, 4, 4, 4, 4},
}
