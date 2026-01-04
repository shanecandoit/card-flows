package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Card represents a node on the canvas
type Card struct {
	X, Y          float64
	Width, Height float64
	Color         color.Color
	Title         string
}

// Camera controls the viewport of the infinite canvas
type Camera struct {
	X, Y float64 // World position of the center of the screen
	Zoom float64
}

type Game struct {
	cards        []*Card
	camera       Camera
	screenWidth  int
	screenHeight int

	// Input state for panning
	isPanning  bool
	lastMouseX int
	lastMouseY int
}

func NewGame() *Game {
	g := &Game{
		camera: Camera{X: 400, Y: 200, Zoom: 1.0},
		cards:  []*Card{},
	}

	// Add some dummy cards (Constrained to x >= 0, y >= 0)
	g.cards = append(g.cards, &Card{X: 50, Y: 50, Width: 200, Height: 120, Color: color.RGBA{100, 149, 237, 255}, Title: "Input Data"})
	g.cards = append(g.cards, &Card{X: 300, Y: 200, Width: 180, Height: 100, Color: color.RGBA{255, 105, 180, 255}, Title: "Transformation"})
	g.cards = append(g.cards, &Card{X: 100, Y: 400, Width: 220, Height: 140, Color: color.RGBA{60, 179, 113, 255}, Title: "Output Plot"})

	return g
}

func (g *Game) Update() error {
	// --- Zooming ---
	_, dy := ebiten.Wheel()
	if dy != 0 {
		zoomSpeed := 0.1
		newZoom := g.camera.Zoom * math.Pow(1+zoomSpeed, dy)
		if newZoom > 0.1 && newZoom < 5.0 {
			g.camera.Zoom = newZoom
		}
	}

	// --- Panning ---
	mx, my := ebiten.CursorPosition()
	isPanButtonHeld := ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) ||
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	if !g.isPanning {
		if isPanButtonHeld {
			shouldStartPan := false
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
				shouldStartPan = true
			} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				if ebiten.IsKeyPressed(ebiten.KeySpace) {
					shouldStartPan = true
				} else {
					wx, wy := g.screenToWorld(float64(mx), float64(my))
					if g.getCardAt(wx, wy) == nil {
						shouldStartPan = true
					}
				}
			}

			if shouldStartPan {
				g.isPanning = true
				g.lastMouseX, g.lastMouseY = mx, my
			}
		}
	} else {
		if isPanButtonHeld {
			dx := float64(mx - g.lastMouseX)
			dy := float64(my - g.lastMouseY)

			g.camera.X -= dx / g.camera.Zoom
			g.camera.Y -= dy / g.camera.Zoom

			// Apply Canvas Limits (Top and Left limit)
			// We clamp so we can't pan too far into the negative space.
			// The user wants a top/left limit. Let's say camera can't go
			// beyond showing some margin of the origin.
			if g.camera.X < -200 {
				g.camera.X = -200
			}
			if g.camera.Y < -200 {
				g.camera.Y = -200
			}

			g.lastMouseX, g.lastMouseY = mx, my
		} else {
			g.isPanning = false
		}
	}

	return nil
}

func (g *Game) getCardAt(wx, wy float64) *Card {
	for i := len(g.cards) - 1; i >= 0; i-- {
		card := g.cards[i]
		if wx >= card.X && wx < card.X+card.Width &&
			wy >= card.Y && wy < card.Y+card.Height {
			return card
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 35, 255})

	cw := float64(g.screenWidth) / 2
	ch := float64(g.screenHeight) / 2

	g.drawGrid(screen, cw, ch)

	for _, card := range g.cards {
		screenX, screenY := g.worldToScreen(card.X, card.Y, cw, ch)
		screenW := card.Width * g.camera.Zoom
		screenH := card.Height * g.camera.Zoom

		vector.DrawFilledRect(screen, float32(screenX+5*g.camera.Zoom), float32(screenY+5*g.camera.Zoom), float32(screenW), float32(screenH), color.RGBA{0, 0, 0, 100}, false)
		vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(screenW), float32(screenH), card.Color, false)

		msg := fmt.Sprintf("%s\n(%.0f, %.0f)", card.Title, card.X, card.Y)
		ebitenutil.DebugPrintAt(screen, msg, int(screenX+5), int(screenY+5))
	}

	mx, my := ebiten.CursorPosition()
	wx, wy := g.screenToWorld(float64(mx), float64(my))
	hoverStatus := "None"
	if c := g.getCardAt(wx, wy); c != nil {
		hoverStatus = c.Title
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"Camera: (%.1f, %.1f) Zoom: %.2f\n"+
			"Mouse World: (%.1f, %.1f)\n"+
			"Hovering: %s\n"+
			"Limit: x>=0, y>=0 (Grid)",
		g.camera.X, g.camera.Y, g.camera.Zoom,
		wx, wy,
		hoverStatus,
	))
}

func (g *Game) drawGrid(screen *ebiten.Image, cw, ch float64) {
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

	// Vertical lines (wx >= 0)
	for wx := startWx; wx < right; wx += 100.0 {
		sx, _ := g.worldToScreen(wx, 0, cw, ch)
		// Only draw from the top edge of the grid (y=0)
		_, syStart := g.worldToScreen(0, 0, cw, ch)
		vector.StrokeLine(screen, float32(sx), float32(math.Max(0, syStart)), float32(sx), float32(g.screenHeight), 1, color.RGBA{255, 255, 255, 20}, false)
	}

	// Horizontal lines (wy >= 0)
	for wy := startWy; wy < bottom; wy += 100.0 {
		_, sy := g.worldToScreen(0, wy, cw, ch)
		// Only draw from the left edge of the grid (x=0)
		sxStart, _ := g.worldToScreen(0, 0, cw, ch)
		vector.StrokeLine(screen, float32(math.Max(0, sxStart)), float32(sy), float32(g.screenWidth), float32(sy), 1, color.RGBA{255, 255, 255, 20}, false)
	}

	originX, originY := g.worldToScreen(0, 0, cw, ch)

	// Draw "Void" overlays for negative space
	if originX > 0 {
		vector.DrawFilledRect(screen, 0, 0, float32(originX), float32(g.screenHeight), color.RGBA{20, 20, 25, 255}, false)
	}
	if originY > 0 {
		vector.DrawFilledRect(screen, float32(math.Max(0, originX)), 0, float32(g.screenWidth), float32(originY), color.RGBA{20, 20, 25, 255}, false)
	}

	// Origin Marker
	vector.StrokeLine(screen, float32(originX-15), float32(originY), float32(originX+15), float32(originY), 2, color.RGBA{255, 100, 100, 150}, false)
	vector.StrokeLine(screen, float32(originX), float32(originY-15), float32(originX), float32(originY+15), 2, color.RGBA{255, 100, 100, 150}, false)
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

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Card Flows Infinite Canvas")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
