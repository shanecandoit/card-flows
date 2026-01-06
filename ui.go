package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

func (b *Button) Draw(screen *ebiten.Image) {
	buttonColor := ColorButtonBackground
	vector.DrawFilledRect(screen, b.X, b.Y, b.W, b.H, buttonColor, false)
	ebitenutil.DebugPrintAt(screen, b.Label, int(b.X)+10, int(b.Y)+8)
}

type UISystem struct {
	game    *Game
	buttons []*Button
}

func NewUISystem(g *Game) *UISystem {
	ui := &UISystem{
		game: g,
	}
	ui.initButtons()
	return ui
}

func (ui *UISystem) initButtons() {
	g := ui.game
	ui.buttons = []*Button{
		{
			Label: "+",
			W:     ButtonWidth, H: ButtonHeight,
			OnClick: func() {
				newZoom := g.camera.Zoom * (1 + ZoomSpeed)
				if newZoom < ZoomLimitMax {
					g.camera.Zoom = newZoom
				}
			},
		},
		{
			Label: "-",
			W:     ButtonWidth, H: ButtonHeight,
			OnClick: func() {
				newZoom := g.camera.Zoom / (1 + ZoomSpeed)
				if newZoom > ZoomLimitMin {
					g.camera.Zoom = newZoom
				}
			},
		},
	}
}

func (ui *UISystem) updateButtonPositions() {
	g := ui.game
	// Position relative to top-right
	ui.buttons[0].X = float32(g.screenWidth) - ButtonWidth - ButtonMargin
	ui.buttons[0].Y = ButtonMargin
	ui.buttons[1].X = float32(g.screenWidth) - 2*ButtonWidth - 2*ButtonMargin
	ui.buttons[1].Y = ButtonMargin
}

func (ui *UISystem) IsMouseOver(mx, my int) bool {
	ui.updateButtonPositions()
	for _, b := range ui.buttons {
		if b.IsMouseOver(mx, my) {
			return true
		}
	}
	return false
}

func (ui *UISystem) Update() {
	ui.updateButtonPositions()
	mx, my := ebiten.CursorPosition()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for _, b := range ui.buttons {
			if b.IsMouseOver(mx, my) {
				b.OnClick()
				break
			}
		}
	}
}

func (ui *UISystem) Draw(screen *ebiten.Image) {
	ui.updateButtonPositions()
	for _, b := range ui.buttons {
		b.Draw(screen)
	}
}
