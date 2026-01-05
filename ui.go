package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type UISystem struct {
	game *Game
}

func NewUISystem(g *Game) *UISystem {
	return &UISystem{
		game: g,
	}
}

func (ui *UISystem) IsMouseOver(mx, my int) bool {
	// Simple hardcoded check for now, matches previous logic
	return mx >= ui.game.screenWidth-80 && mx <= ui.game.screenWidth-10 && my >= 10 && my <= 40
}

func (ui *UISystem) Update() {
	g := ui.game
	mx, my := ebiten.CursorPosition()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Plus Button
		if mx >= g.screenWidth-40 && mx <= g.screenWidth-10 && my >= 10 && my <= 40 {
			newZoom := g.camera.Zoom * 1.1
			if newZoom < 10.0 {
				g.camera.Zoom = newZoom
			}
		}
		// Minus Button
		if mx >= g.screenWidth-80 && mx <= g.screenWidth-50 && my >= 10 && my <= 40 {
			newZoom := g.camera.Zoom / 1.1
			if newZoom > 0.1 {
				g.camera.Zoom = newZoom
			}
		}
	}
}

func (ui *UISystem) Draw(screen *ebiten.Image) {
	g := ui.game

	// --- UI Buttons (Top Right) ---
	buttonColor := color.RGBA{60, 60, 70, 200}

	// Plus Button
	vector.DrawFilledRect(screen, float32(g.screenWidth-40), 10, 30, 30, buttonColor, false)
	ebitenutil.DebugPrintAt(screen, "+", g.screenWidth-30, 18)

	// Minus Button
	vector.DrawFilledRect(screen, float32(g.screenWidth-80), 10, 30, 30, buttonColor, false)
	ebitenutil.DebugPrintAt(screen, "-", g.screenWidth-70, 18)
}
