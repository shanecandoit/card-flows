package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

type Button struct {
	Label   string
	X, Y    float32
	W, H    float32
	OnClick func()
}

func (b *Button) IsMouseOver(mx, my int) bool {
	return float32(mx) >= b.X && float32(mx) <= b.X+b.W &&
		float32(my) >= b.Y && float32(my) <= b.Y+b.H
}

// Draw renders the button. It uses the provided font.Face via getter.
func (b *Button) Draw(screen *ebiten.Image, getFace func() font.Face, drawText func(screen *ebiten.Image, face font.Face, s string, x, y int, clr color.Color)) {
	buttonColor := color.RGBA{60, 60, 70, 200}
	vector.DrawFilledRect(screen, b.X, b.Y, b.W, b.H, buttonColor, false)
	if getFace == nil || drawText == nil {
		return
	}
	face := getFace()
	if face == nil {
		return
	}
	drawText(screen, face, b.Label, int(b.X)+10, int(b.Y)+12, color.White)
}
