package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawBackgroundGrid renders the infinite coordinate grid
func (g *Game) drawBackgroundGrid(screen *ebiten.Image, cw, ch float64) {
	left, top := g.screenToWorld(0, 0)
	right, bottom := g.screenToWorld(float64(g.screenWidth), float64(g.screenHeight))

	startWx := math.Floor(left/100.0) * 100.0
	if startWx < 0 {
		startWx = 0
	}
	startWy := math.Floor(top/100.0) * 100.0
	if startWy < 0 {
		startWy = 0
	}

	gridGrey := color.RGBA{32, 32, 10, 10}

	// Vertical lines (wx >= 0)
	for wx := startWx; wx < right; wx += 50.0 {
		sx, _ := g.worldToScreen(wx, 0, cw, ch)
		_, syStart := g.worldToScreen(0, 0, cw, ch)
		vector.StrokeLine(screen, float32(sx),
			float32(math.Max(0, syStart)), float32(sx),
			float32(g.screenHeight), 1, gridGrey, false)
	}

	// Horizontal lines (wy >= 0)
	for wy := startWy; wy < bottom; wy += 50.0 {
		_, sy := g.worldToScreen(0, wy, cw, ch)
		sxStart, _ := g.worldToScreen(0, 0, cw, ch)
		vector.StrokeLine(screen, float32(math.Max(0, sxStart)),
			float32(sy), float32(g.screenWidth), float32(sy), 1, gridGrey, false)
	}

	originX, originY := g.worldToScreen(0, 0, cw, ch)

	if originX > 0 {
		vector.DrawFilledRect(screen, 0, 0, float32(originX), float32(g.screenHeight), color.RGBA{20, 20, 25, 255}, false)
	}
	if originY > 0 {
		vector.DrawFilledRect(screen, float32(math.Max(0, originX)), 0, float32(g.screenWidth), float32(originY), color.RGBA{20, 20, 25, 255}, false)
	}

	vector.StrokeLine(screen, float32(originX-15), float32(originY), float32(originX+15), float32(originY), 2, color.RGBA{255, 100, 100, 150}, false)
	vector.StrokeLine(screen, float32(originX), float32(originY-15), float32(originX), float32(originY+15), 2, color.RGBA{255, 100, 100, 150}, false)
}
