package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

type DebugPanel struct {
	Error string
}

func (d *DebugPanel) SetError(msg string) {
	d.Error = msg
}

func (d *DebugPanel) Clear() {
	d.Error = ""
}

func (d *DebugPanel) Draw(screen *ebiten.Image, getScreenSize func() (int, int), getFace func() font.Face, drawText func(screen *ebiten.Image, face font.Face, s string, x, y int, clr color.Color)) {
	if d == nil || d.Error == "" {
		return
	}
	w, h := getScreenSize()
	// Panel size
	pw, ph := 300, 80
	x := w - pw - 10
	y := h - ph - 10
	bg := color.RGBA{40, 40, 40, 220}
	vectorDrawRect(screen, float32(x), float32(y), float32(pw), float32(ph), bg)
	if getFace != nil && drawText != nil {
		face := getFace()
		if face != nil {
			drawText(screen, face, d.Error, x+8, y+8, color.RGBA{255, 200, 50, 255})
		}
	}
}

// Minimal helper to draw rect without importing vector package twice (keeps implementation small)
func vectorDrawRect(screen *ebiten.Image, x, y, w, h float32, clr color.Color) {
	// Use vector stroked rectangle to draw filled rect. Re-importing vector would be fine, but keep small helper.
	// We'll cheat: create a tiny 1x1 image and scale it. Simpler and avoids extra imports.
	img := ebiten.NewImage(int(w), int(h))
	img.Fill(clr)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}
