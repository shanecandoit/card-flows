package main

import (
	"fmt"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
		camera: Camera{X: DefaultCameraX, Y: DefaultCameraY, Zoom: DefaultCameraZoom},
		cards:  []*Card{},
	}

	g.input = NewInputSystem(g)
	g.ui = NewUISystem(g)

	err := LoadState(g, "state.yaml")
	if err == nil {
		return g
	}

	// Default dummy cards if load fails
	g.cards = append(g.cards, &Card{
		X: 50, Y: 50, Width: 200, Height: 120, Color: color.RGBA{100, 149, 237, 255}, Title: "Input Data",
		Outputs: []Port{{Name: "data", Type: "any"}},
	})
	g.cards = append(g.cards, &Card{
		X: 300, Y: 200, Width: 180, Height: 100, Color: color.RGBA{255, 105, 180, 255}, Title: "Transformation",
		Inputs:  []Port{{Name: "in", Type: "any"}},
		Outputs: []Port{{Name: "out", Type: "any"}},
	})
	g.cards = append(g.cards, &Card{
		X: 100, Y: 400, Width: 220, Height: 140, Color: color.RGBA{60, 179, 113, 255}, Title: "Output Plot",
		Inputs: []Port{{Name: "data", Type: "any"}},
	})

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

func (g *Game) DeleteCard(c *Card) {
	newCards := []*Card{}
	for _, card := range g.cards {
		if card != c {
			newCards = append(newCards, card)
		}
	}
	g.cards = newCards
}

func (g *Game) DuplicateCard(c *Card) {
	newCard := &Card{
		X:      c.X + DuplicateOffset,
		Y:      c.Y + DuplicateOffset,
		Width:  c.Width,
		Height: c.Height,
		Color:  c.Color,
		Title:  c.Title + " (Copy)",
		Text:   c.Text,
	}
	// Copy ports
	for _, p := range c.Inputs {
		newCard.Inputs = append(newCard.Inputs, Port{Name: p.Name, Type: p.Type})
	}
	for _, p := range c.Outputs {
		newCard.Outputs = append(newCard.Outputs, Port{Name: p.Name, Type: p.Type})
	}
	g.cards = append(g.cards, newCard)
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
	screen.Fill(ColorBackground)

	cw := float64(g.screenWidth) / 2
	ch := float64(g.screenHeight) / 2

	g.drawBackgroundGrid(screen, cw, ch)

	mx, my := ebiten.CursorPosition()
	wx, wy := g.screenToWorld(float64(mx), float64(my))
	hoveredCard := g.getCardAt(wx, wy)

	for _, card := range g.cards {
		card.Draw(screen, g, cw, ch, card == hoveredCard)
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

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight
	return outsideWidth, outsideHeight
}
