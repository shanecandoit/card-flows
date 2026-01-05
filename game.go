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
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	cards        []*Card
	camera       Camera
	screenWidth  int
	screenHeight int

	// Sub-systems
	input *InputSystem
	ui    *UISystem

	screenshotRequested bool
}

func NewGame() *Game {
	g := &Game{
		camera: Camera{X: 400, Y: 200, Zoom: 1.0},
		cards:  []*Card{},
	}

	g.input = NewInputSystem(g)
	g.ui = NewUISystem(g)

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
	// Delegate to sub-systems
	g.input.Update()
	g.ui.Update()
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

		if card == g.input.activeCard {
			showBorder = true
			if g.input.isHot {
				borderColor = color.RGBA{255, 140, 0, 255} // Orange
			} else {
				borderColor = color.RGBA{50, 205, 50, 255} // Green
			}
		} else if card == hoveredCard && g.input.activeCard == nil {
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
		resizingThis := (card == g.input.resizingCard)
		hCorner := card.GetCornerAt(wx, wy, g.camera.Zoom)
		if (card == hoveredCard && hCorner != -1) || resizingThis {
			cIdx := hCorner
			if resizingThis {
				cIdx = g.input.resizingCorner
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

	g.ui.Draw(screen)

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
