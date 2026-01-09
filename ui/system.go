package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
)

type UISystem struct {
	buttons       []*Button
	getFontFace   func() font.Face
	getScreenSize func() (int, int)
	onZoomIn      func()
	onZoomOut     func()
	drawText      func(screen *ebiten.Image, face font.Face, s string, x, y int, clr color.Color)
	Debug         *DebugPanel
}

func NewUISystem(getFontFace func() font.Face, getScreenSize func() (int, int), onZoomIn func(), onZoomOut func(), drawText func(screen *ebiten.Image, face font.Face, s string, x, y int, clr color.Color)) *UISystem {
	ui := &UISystem{
		getFontFace:   getFontFace,
		getScreenSize: getScreenSize,
		onZoomIn:      onZoomIn,
		onZoomOut:     onZoomOut,
		drawText:      drawText,
		Debug:         &DebugPanel{},
	}
	ui.initButtons()
	return ui
}

func (ui *UISystem) initButtons() {
	w, _ := ui.getScreenSize()
	zoomIn := &Button{Label: "+", W: 30, H: 30, OnClick: ui.onZoomIn}
	zoomOut := &Button{Label: "-", W: 30, H: 30, OnClick: ui.onZoomOut}
	// Position relative to top-right
	zoomIn.X = float32(w) - zoomIn.W - 10
	zoomIn.Y = 10
	zoomOut.X = float32(w) - 2*zoomOut.W - 20
	zoomOut.Y = 10
	ui.buttons = []*Button{zoomIn, zoomOut}
}

func (ui *UISystem) updateButtonPositions() {
	w, _ := ui.getScreenSize()
	if len(ui.buttons) < 2 {
		return
	}
	ui.buttons[0].X = float32(w) - ui.buttons[0].W - 10
	ui.buttons[0].Y = 10
	ui.buttons[1].X = float32(w) - 2*ui.buttons[1].W - 20
	ui.buttons[1].Y = 10
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
				if b.OnClick != nil {
					b.OnClick()
				}
				break
			}
		}
	}
}

func (ui *UISystem) Draw(screen *ebiten.Image) {
	ui.updateButtonPositions()
	for _, b := range ui.buttons {
		b.Draw(screen, ui.getFontFace, ui.drawText)
	}
	if ui.Debug != nil {
		ui.Debug.Draw(screen, ui.getScreenSize, ui.getFontFace, ui.drawText)
	}
}
