package main

// Camera controls the viewport of the infinite canvas
type Camera struct {
	X, Y float64 // World position of the center of the screen
	Zoom float64
}

func (g *Game) worldToScreen(wx, wy, cw, ch float64) (float64, float64) {
	sx := (wx-g.camera.X)*g.camera.Zoom + cw
	sy := (wy-g.camera.Y)*g.camera.Zoom + ch
	return sx, sy
}

func (g *Game) screenToWorld(sx, sy float64) (float64, float64) {
	cw := float64(g.screenWidth) / 2
	ch := float64(g.screenHeight) / 2

	wx := (sx-cw)/g.camera.Zoom + g.camera.X
	wy := (sy-ch)/g.camera.Zoom + g.camera.Y
	return wx, wy
}
