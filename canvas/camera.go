package canvas

// Camera controls the viewport of the infinite canvas
type Camera struct {
	X, Y float64 // World position of the center of the screen
	Zoom float64
}

func (c *Camera) WorldToScreen(wx, wy, cw, ch float64) (float64, float64) {
	sx := (wx-c.X)*c.Zoom + cw
	sy := (wy-c.Y)*c.Zoom + ch
	return sx, sy
}

func (c *Camera) ScreenToWorld(sx, sy, cw, ch float64) (float64, float64) {
	wx := (sx-cw)/c.Zoom + c.X
	wy := (sy-ch)/c.Zoom + c.Y
	return wx, wy
}
