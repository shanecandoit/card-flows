package main

import (
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	cards        []*Card
	camera       Camera
	screenWidth  int
	screenHeight int

	// Input state for panning
	isPanning  bool
	lastMouseX int
	lastMouseY int

	// Input state for card dragging
	activeCard  *Card
	dragOffsetX float64
	dragOffsetY float64
	isHot       bool // True for the first frame a card is clicked

	// Input state for resizing
	resizingCard   *Card
	resizingCorner int // 0: TL, 1: TR, 2: BL, 3: BR

	screenshotRequested bool
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

	// Add String:find_replace block
	g.cards = append(g.cards, &Card{
		Title:  "String:find_replace",
		X:      500,
		Y:      50,
		Width:  250,
		Height: 150,
		Color:  color.RGBA{100, 100, 250, 255},
		Inputs: []Port{
			{Name: "input", Type: "string"},
			{Name: "find", Type: "string"},
			{Name: "replace", Type: "string"},
		},
		Outputs: []Port{
			{Name: "result", Type: "string"},
		},
	})

	return g
}

func (g *Game) Update() error {
	// --- Screenshot ---
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.screenshotRequested = true
	}

	// --- Zooming ---
	_, dy := ebiten.Wheel()

	// Keyboard Zooming
	if ebiten.IsKeyPressed(ebiten.KeyEqual) || ebiten.IsKeyPressed(ebiten.KeyKPAdd) {
		dy += 0.1
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) || ebiten.IsKeyPressed(ebiten.KeyKPSubtract) {
		dy -= 0.1
	}

	if dy != 0 {
		zoomSpeed := 0.1
		newZoom := g.camera.Zoom * math.Pow(1+zoomSpeed, dy)
		if newZoom > 0.1 && newZoom < 10.0 {
			g.camera.Zoom = newZoom
		}
	}

	mx, my := ebiten.CursorPosition()

	// UI Buttons Hit Detection (Top Right)
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

	wx, wy := g.screenToWorld(float64(mx), float64(my))

	// --- Card Resizing Logic ---
	overUI := mx >= g.screenWidth-80 && mx <= g.screenWidth-10 && my >= 10 && my <= 40
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !ebiten.IsKeyPressed(ebiten.KeySpace) && !overUI {
		// Check for corners first
		for i := len(g.cards) - 1; i >= 0; i-- {
			card := g.cards[i]
			corner := card.GetCornerAt(wx, wy, g.camera.Zoom)
			if corner != -1 {
				g.resizingCard = card
				g.resizingCorner = corner
				break
			}
		}

		if g.resizingCard == nil {
			// If not resizing, try dragging
			if card := g.getCardAt(wx, wy); card != nil {
				g.activeCard = card
				g.dragOffsetX = wx - card.X
				g.dragOffsetY = wy - card.Y
				g.isHot = true
			}
		}
	} else if g.resizingCard != nil {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			card := g.resizingCard
			minSize := 50.0
			maxSize := 500.0

			// Snap target coordinates to 50px grid
			swx := math.Round(wx/50) * 50
			swy := math.Round(wy/50) * 50

			switch g.resizingCorner {
			case 0: // TL
				diffX := card.X - swx
				diffY := card.Y - swy
				newW := card.Width + diffX
				newH := card.Height + diffY
				if newW >= minSize && newW <= maxSize {
					card.X = swx
					card.Width = newW
				}
				if newH >= minSize && newH <= maxSize {
					card.Y = swy
					card.Height = newH
				}
			case 1: // TR
				newW := swx - card.X
				diffY := card.Y - swy
				newH := card.Height + diffY
				if newW >= minSize && newW <= maxSize {
					card.Width = newW
				}
				if newH >= minSize && newH <= maxSize {
					card.Y = swy
					card.Height = newH
				}
			case 2: // BL
				diffX := card.X - swx
				newW := card.Width + diffX
				newH := swy - card.Y
				if newW >= minSize && newW <= maxSize {
					card.X = swx
					card.Width = newW
				}
				if newH >= minSize && newH <= maxSize {
					card.Height = newH
				}
			case 3: // BR
				newW := swx - card.X
				newH := swy - card.Y
				if newW >= minSize && newW <= maxSize {
					card.Width = newW
				}
				if newH >= minSize && newH <= maxSize {
					card.Height = newH
				}
			}
		} else {
			// Released
			g.resizingCard.Width = math.Round(g.resizingCard.Width/50) * 50
			g.resizingCard.Height = math.Round(g.resizingCard.Height/50) * 50
			g.resizingCard = nil
		}
	} else if g.activeCard != nil {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.isHot = false
			// Move card
			newX := wx - g.dragOffsetX
			newY := wy - g.dragOffsetY

			// Small grid snap for movement
			g.activeCard.X = math.Round(newX/10) * 10
			g.activeCard.Y = math.Round(newY/10) * 10
		} else {
			// Released - Snap to main grid
			g.activeCard.X = math.Round(g.activeCard.X/50) * 50
			g.activeCard.Y = math.Round(g.activeCard.Y/50) * 50

			if g.activeCard.X < 0 {
				g.activeCard.X = 0
			}
			if g.activeCard.Y < 0 {
				g.activeCard.Y = 0
			}

			g.activeCard = nil
			g.isHot = false
		}
	}

	// --- Panning ---
	isPanButtonHeld := ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) ||
		(ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.activeCard == nil && g.resizingCard == nil && !overUI)

	if !g.isPanning {
		if isPanButtonHeld {
			shouldStartPan := false
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
				shouldStartPan = true
			} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				if ebiten.IsKeyPressed(ebiten.KeySpace) {
					shouldStartPan = true
				} else {
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

	mx, my := ebiten.CursorPosition()
	wx, wy := g.screenToWorld(float64(mx), float64(my))
	hoveredCard := g.getCardAt(wx, wy)

	for _, card := range g.cards {
		screenX, screenY := g.worldToScreen(card.X, card.Y, cw, ch)
		screenW := card.Width * g.camera.Zoom
		screenH := card.Height * g.camera.Zoom

		// Shadow
		vector.DrawFilledRect(screen, float32(screenX+5*g.camera.Zoom), float32(screenY+5*g.camera.Zoom), float32(screenW), float32(screenH), color.RGBA{0, 0, 0, 100}, false)

		// Body
		vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(screenW), float32(screenH), card.Color, false)

		// Border Logic
		borderColor := color.RGBA{0, 0, 0, 0}
		showBorder := false

		if card == g.activeCard {
			showBorder = true
			if g.isHot {
				borderColor = color.RGBA{255, 140, 0, 255} // Orange
			} else {
				borderColor = color.RGBA{50, 205, 50, 255} // Green
			}
		} else if card == hoveredCard && g.activeCard == nil {
			showBorder = true
			borderColor = color.RGBA{0, 120, 255, 255} // Blue
		}

		if showBorder {
			borderThickness := float32(3 * g.camera.Zoom)
			borderOffset := float32(2 * g.camera.Zoom)
			vector.StrokeRect(screen,
				float32(screenX)-borderOffset-borderThickness/2,
				float32(screenY)-borderOffset-borderThickness/2,
				float32(screenW)+2*(borderOffset+borderThickness/2),
				float32(screenH)+2*(borderOffset+borderThickness/2),
				borderThickness, borderColor, false)
		}

		// Draw Corner Handle if hovering or resizing
		resizingThis := (card == g.resizingCard)
		hCorner := card.GetCornerAt(wx, wy, g.camera.Zoom)
		if (card == hoveredCard && hCorner != -1) || resizingThis {
			cIdx := hCorner
			if resizingThis {
				cIdx = g.resizingCorner
			}

			var cx, cy float64
			switch cIdx {
			case 0:
				cx, cy = card.X, card.Y
			case 1:
				cx, cy = card.X+card.Width, card.Y
			case 2:
				cx, cy = card.X, card.Y+card.Height
			case 3:
				cx, cy = card.X+card.Width, card.Y+card.Height
			}

			scx, scy := g.worldToScreen(cx, cy, cw, ch)
			radius := float32(6 * g.camera.Zoom)
			vector.DrawFilledCircle(screen, float32(scx), float32(scy), radius, color.RGBA{255, 255, 255, 200}, false)
			vector.StrokeCircle(screen, float32(scx), float32(scy), radius, 2, color.RGBA{0, 120, 255, 255}, false)
		}

		// Title
		msg := fmt.Sprintf("%s\n(%.0f, %.0f)", card.Title, card.X, card.Y)
		ebitenutil.DebugPrintAt(screen, msg, int(screenX+5), int(screenY+5))

		// --- Dividers ---
		dividerColor := color.RGBA{0, 0, 0, 50}
		// Header divider
		headerHeight := 50.0
		hy := card.Y + headerHeight
		shx1, shy := g.worldToScreen(card.X, hy, cw, ch)
		shx2, _ := g.worldToScreen(card.X+card.Width, hy, cw, ch)
		vector.StrokeLine(screen, float32(shx1), float32(shy), float32(shx2), float32(shy), 1, dividerColor, false)

		// Footer divider (only if there are outputs)
		footerHeight := 0.0
		if len(card.Outputs) > 0 {
			footerHeight = 30.0
			fy := card.Y + card.Height - footerHeight
			sfx1, sfy := g.worldToScreen(card.X, fy, cw, ch)
			sfx2, _ := g.worldToScreen(card.X+card.Width, fy, cw, ch)
			vector.StrokeLine(screen, float32(sfx1), float32(sfy), float32(sfx2), float32(sfy), 1, dividerColor, false)
		}

		// --- Ports Rendering ---
		portSize := 10.0 * g.camera.Zoom

		// Inputs (Left edge)
		if len(card.Inputs) > 0 {
			usableHeight := card.Height - headerHeight - footerHeight
			ySpacing := usableHeight / float64(len(card.Inputs)+1)
			for i, port := range card.Inputs {
				py := card.Y + headerHeight + ySpacing*float64(i+1)
				spx, spy := g.worldToScreen(card.X, py, cw, ch)

				// Grey Square
				vector.DrawFilledRect(screen, float32(spx-portSize/2), float32(spy-portSize/2), float32(portSize), float32(portSize), color.RGBA{150, 150, 150, 255}, false)
				// Black hole
				vector.DrawFilledCircle(screen, float32(spx), float32(spy), float32(3*g.camera.Zoom), color.RGBA{0, 0, 0, 255}, false)

				// Label
				label := fmt.Sprintf("%s:%s", port.Name, port.Type)
				ebitenutil.DebugPrintAt(screen, label, int(spx+portSize), int(spy-8*g.camera.Zoom))
			}
		}

		// Outputs (Bottom edge)
		if len(card.Outputs) > 0 {
			xSpacing := card.Width / float64(len(card.Outputs)+1)
			for i, port := range card.Outputs {
				px := card.X + xSpacing*float64(i+1)
				spx, spy := g.worldToScreen(px, card.Y+card.Height, cw, ch)

				// Grey Square
				vector.DrawFilledRect(screen, float32(spx-portSize/2), float32(spy-portSize/2), float32(portSize), float32(portSize), color.RGBA{150, 150, 150, 255}, false)
				// Black hole
				vector.DrawFilledCircle(screen, float32(spx), float32(spy), float32(3*g.camera.Zoom), color.RGBA{0, 0, 0, 255}, false)

				// Label
				label := fmt.Sprintf("%s:%s", port.Name, port.Type)
				ebitenutil.DebugPrintAt(screen, label, int(spx-20*g.camera.Zoom), int(spy-20*g.camera.Zoom))
			}
		}
	}

	hoverStatus := "None"
	if hoveredCard != nil {
		hoverStatus = hoveredCard.Title
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(
		"Camera: (%.1f, %.1f) Zoom: %.2f\n"+
			"Mouse World: (%.1f, %.1f)\n"+
			"Hovering: %s\n"+
			"Drag: Left Click to move cards\n"+
			"Pan: Left Drag (Empty Space) or Middle Drag",
		g.camera.X, g.camera.Y, g.camera.Zoom,
		wx, wy,
		hoverStatus,
	), 10, 10)

	// --- UI Buttons (Top Right) ---
	buttonColor := color.RGBA{60, 60, 70, 200}

	// Plus Button
	vector.DrawFilledRect(screen, float32(g.screenWidth-40), 10, 30, 30, buttonColor, false)
	ebitenutil.DebugPrintAt(screen, "+", g.screenWidth-30, 18)

	// Minus Button
	vector.DrawFilledRect(screen, float32(g.screenWidth-80), 10, 30, 30, buttonColor, false)
	ebitenutil.DebugPrintAt(screen, "-", g.screenWidth-70, 18)

	// --- Save Screenshot ---
	if g.screenshotRequested {
		g.screenshotRequested = false
		f, err := os.Create("screenshot.png")
		if err != nil {
			log.Println("screenshot error:", err)
		} else {
			defer f.Close()
			if err := png.Encode(f, screen); err != nil {
				log.Println("screenshot error:", err)
			} else {
				log.Println("Screenshot saved as screenshot.png")
			}
		}
	}
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

	gridGrey := color.RGBA{32, 32, 32, 10}

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

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight
	return outsideWidth, outsideHeight
}
