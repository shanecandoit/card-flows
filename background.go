package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawBackgroundGrid renders the infinite coordinate grid
func (g *Game) drawBackgroundGrid(screen *ebiten.Image, cw, ch float64) {
	left, top := g.screenToWorld(0, 0)
	right, bottom := g.screenToWorld(float64(g.screenWidth), float64(g.screenHeight))

	startWx := math.Floor(left/GridSizeLarge) * GridSizeLarge
	if startWx < 0 {
		startWx = 0
	}
	startWy := math.Floor(top/GridSizeLarge) * GridSizeLarge
	if startWy < 0 {
		startWy = 0
	}

	gridGrey := ColorGrid

	// Vertical lines (wx >= 0)
	for wx := startWx; wx < right; wx += GridSizeSmall {
		sx, _ := g.worldToScreen(wx, 0, cw, ch)
		_, syStart := g.worldToScreen(0, 0, cw, ch)
		vector.StrokeLine(screen, float32(sx),
			float32(math.Max(0, syStart)), float32(sx),
			float32(g.screenHeight), 1, gridGrey, false)
	}

	// Horizontal lines (wy >= 0)
	for wy := startWy; wy < bottom; wy += GridSizeSmall {
		_, sy := g.worldToScreen(0, wy, cw, ch)
		sxStart, _ := g.worldToScreen(0, 0, cw, ch)
		vector.StrokeLine(screen, float32(math.Max(0, sxStart)),
			float32(sy), float32(g.screenWidth), float32(sy), 1, gridGrey, false)
	}

	originX, originY := g.worldToScreen(0, 0, cw, ch)

	if originX > 0 {
		vector.DrawFilledRect(screen, 0, 0, float32(originX), float32(g.screenHeight), ColorGridBlocked, false)
	}
	if originY > 0 {
		vector.DrawFilledRect(screen, float32(math.Max(0, originX)), 0, float32(g.screenWidth), float32(originY), ColorGridBlocked, false)
	}

	vector.StrokeLine(screen, float32(originX-15), float32(originY), float32(originX+15), float32(originY), 2, ColorOriginCross, false)
	vector.StrokeLine(screen, float32(originX), float32(originY-15), float32(originX), float32(originY+15), 2, ColorOriginCross, false)
}
