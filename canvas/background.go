package canvas

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawBackgroundGrid renders the infinite coordinate grid using the provided camera and colors.
func DrawBackgroundGrid(cam *Camera, screen *ebiten.Image, cw, ch float64, screenWidth, screenHeight int, gridSizeSmall, gridSizeLarge float64, gridColor, gridBlocked, originCross color.Color) {
	left, top := cam.ScreenToWorld(0, 0, cw, ch)
	right, bottom := cam.ScreenToWorld(float64(screenWidth), float64(screenHeight), cw, ch)

	startWx := math.Floor(left/gridSizeLarge) * gridSizeLarge
	if startWx < 0 {
		startWx = 0
	}
	startWy := math.Floor(top/gridSizeLarge) * gridSizeLarge
	if startWy < 0 {
		startWy = 0
	}

	// Vertical lines (wx >= 0)
	for wx := startWx; wx < right; wx += gridSizeSmall {
		sx, _ := cam.WorldToScreen(wx, 0, cw, ch)
		vector.StrokeLine(screen, float32(sx), 0, float32(sx), float32(screenHeight), 1, gridColor, false)
	}

	// Horizontal lines (wy >= 0)
	for wy := startWy; wy < bottom; wy += gridSizeSmall {
		_, sy := cam.WorldToScreen(0, wy, cw, ch)
		vector.StrokeLine(screen, 0, float32(sy), float32(screenWidth), float32(sy), 1, gridColor, false)
	}

	originX, originY := cam.WorldToScreen(0, 0, cw, ch)

	if originX > 0 {
		vector.DrawFilledRect(screen, 0, 0, float32(originX), float32(screenHeight), gridBlocked, false)
	}
	if originY > 0 {
		vector.DrawFilledRect(screen, float32(math.Max(0, originX)), 0, float32(screenWidth), float32(originY), gridBlocked, false)
	}

	vector.StrokeLine(screen, float32(originX-15), float32(originY), float32(originX+15), float32(originY), 2, originCross, false)
	vector.StrokeLine(screen, float32(originX), float32(originY-15), float32(originX), float32(originY+15), 2, originCross, false)
}
