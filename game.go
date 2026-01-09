package main

import (
	"fmt"
	"image/color"
	"image/png"
	"log"
	"os"
	"regexp"
	"strings"

	"card-flows/canvas"
	"card-flows/input"
	"card-flows/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/font"
)

type Game struct {
	cards        []*Card
	arrows       []*Arrow
	camera       canvas.Camera
	screenWidth  int
	screenHeight int

	// Sub-systems
	input  *input.InputSystem
	ui     *ui.UISystem
	engine *Engine

	screenshotRequested bool
	FontFace            font.Face
}

func NewGame() *Game {
	g := &Game{
		camera: canvas.Camera{X: DefaultCameraX, Y: DefaultCameraY, Zoom: DefaultCameraZoom},
		cards:  []*Card{},
	}

	g.FontFace = LoadUIFont()

	g.input = input.NewInputSystem(g)
	g.ui = ui.NewUISystem(
		func() font.Face { return g.FontFace },
		func() (int, int) { return g.screenWidth, g.screenHeight },
		func() {
			newZoom := g.camera.Zoom * (1 + ZoomSpeed)
			if newZoom < ZoomLimitMax {
				g.camera.Zoom = newZoom
			}
		},
		func() {
			newZoom := g.camera.Zoom / (1 + ZoomSpeed)
			if newZoom > ZoomLimitMin {
				g.camera.Zoom = newZoom
			}
		},
		DrawTextLines,
	)
	g.engine = NewEngine(g)

	err := LoadState(g, "state.yaml")
	if err == nil {
		return g
	}

	// Default dummy cards if load fails
	g.cards = append(g.cards, &Card{
		ID: NewID(),
		X:  50, Y: 50, Width: 200, Height: 120, Color: color.RGBA{100, 149, 237, 255}, Title: "Text Card",
		Text:    "Hello World",
		Inputs:  []Port{{Name: "text", Type: "string"}},
		Outputs: []Port{{Name: "text", Type: "string"}},
	})

	g.cards = append(g.cards, &Card{
		ID:     NewID(),
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
	newID := NewID()

	// Determine base title (strip previous ID suffix or (Copy) if present)
	// Simple heuristic: if it ends with ')', try to strip the last parenthesized group
	baseTitle := c.Title
	if matches := regexp.MustCompile(`^(.*)\s\([a-f0-9]+\)$`).FindStringSubmatch(c.Title); len(matches) > 1 {
		baseTitle = matches[1]
	} else {
		// Also clean up old " (Copy)" style just in case
		baseTitle = strings.ReplaceAll(baseTitle, " (Copy)", "")
	}

	shortID := newID
	if len(newID) > 5 {
		shortID = newID[:5]
	}

	newCard := &Card{
		ID:     newID,
		X:      c.X + DuplicateOffset,
		Y:      c.Y + DuplicateOffset,
		Width:  c.Width,
		Height: c.Height,
		Color:  c.Color,
		Title:  fmt.Sprintf("%s (%s)", baseTitle, shortID),
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

	canvas.DrawBackgroundGrid(&g.camera, screen, cw, ch, g.screenWidth, g.screenHeight, GridSizeSmall, GridSizeLarge, ColorGrid, ColorGridBlocked, ColorOriginCross)

	mx, my := ebiten.CursorPosition()
	wx, wy := g.screenToWorld(float64(mx), float64(my))
	hoveredCard := g.getCardAt(wx, wy)

	// Draw Arrows (Connections)
	for _, arrow := range g.arrows {
		arrow.Draw(screen, g, cw, ch)
	}

	g.drawTemporaryArrow(screen, cw, ch)

	for _, card := range g.cards {
		card.Draw(screen, g, cw, ch, card == hoveredCard)
	}

	hoverStatus := "None"
	if hoveredCard != nil {
		hoverStatus = hoveredCard.Title
	}

	DrawTextLines(screen, g.FontFace, fmt.Sprintf(
		"Camera: (%.1f, %.1f) Zoom: %.2f\n"+
			"Mouse World: (%.1f, %.1f)\n"+
			"Hovering: %s\n"+
			"Drag: Left Click to move cards\n"+
			"Pan: Left Drag (Empty Space) or Middle Drag",
		g.camera.X, g.camera.Y, g.camera.Zoom,
		wx, wy,
		hoverStatus,
	), 10, 10, color.White)

	// Print card IDs for debugging
	if hoveredCard != nil {
		DrawTextLines(screen, g.FontFace, fmt.Sprintf("ID: %s", hoveredCard.ID), 10, 100, color.White)
	}

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

func (g *Game) drawTemporaryArrow(screen *ebiten.Image, cw, ch float64) {
	if !g.input.DraggingArrow {
		return
	}

	if g.input.DragStartCard == nil {
		return
	}

	// startCard is an opaque handle; assert to *Card for drawing
	startCard, ok := g.input.DragStartCard.(*Card)
	if !ok || startCard == nil {
		return
	}

	// Get start position
	x1, y1 := startCard.GetOutputPortPosition(g.input.DragStartPort)
	sx1, sy1 := g.camera.WorldToScreen(x1, y1, cw, ch)

	// Get end position (current mouse position)
	mx, my := ebiten.CursorPosition()
	sx2, sy2 := float64(mx), float64(my)

	// Draw the line
	ebitenutil.DrawLine(screen, sx1, sy1, sx2, sy2, color.RGBA{255, 255, 255, 255})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func (g *Game) getCardByID(id string) *Card {
	for _, c := range g.cards {
		if c.ID == id {
			return c
		}
	}
	return nil
}

func (g *Game) screenToWorld(sx, sy float64) (float64, float64) {
	cw := float64(g.screenWidth) / 2
	ch := float64(g.screenHeight) / 2
	return g.camera.ScreenToWorld(sx, sy, cw, ch)
}

func (g *Game) IsInputPortConnected(cardID string, portName string) bool {
	for _, arrow := range g.arrows {
		if arrow.ToCardID == cardID && arrow.ToPort == portName {
			return true
		}
	}
	return false
}

// RegisterSubscription adds a subscription when an arrow is created
func (g *Game) RegisterSubscription(sourceCardID, targetCardID, targetPort string) {
	sourceCard := g.getCardByID(sourceCardID)
	if sourceCard == nil {
		return
	}

	// Check if already subscribed
	for _, sub := range sourceCard.Subscribers {
		if sub.CardID == targetCardID && sub.Port == targetPort {
			return // Already subscribed
		}
	}

	sourceCard.Subscribers = append(sourceCard.Subscribers, Subscription{
		CardID: targetCardID,
		Port:   targetPort,
	})
}

// UnregisterSubscription removes a subscription when an arrow is deleted
func (g *Game) UnregisterSubscription(sourceCardID, targetCardID, targetPort string) {
	sourceCard := g.getCardByID(sourceCardID)
	if sourceCard == nil {
		return
	}

	// Remove the subscription
	filtered := sourceCard.Subscribers[:0]
	for _, sub := range sourceCard.Subscribers {
		if sub.CardID != targetCardID || sub.Port != targetPort {
			filtered = append(filtered, sub)
		}
	}
	sourceCard.Subscribers = filtered
}

// PropagateText sends this card's text to all subscribers
func (g *Game) PropagateText(sourceCard *Card) {
	for _, sub := range sourceCard.Subscribers {
		targetCard := g.getCardByID(sub.CardID)
		if targetCard != nil {
			// For now, directly update the text
			// In the future, this might go through the execution engine
			targetCard.Text = sourceCard.Text

			// Recursively propagate if the target card also has subscribers
			g.PropagateText(targetCard)
		}
	}
}

// GetInputValue returns the value for an input port
// If connected, returns the source card's text; otherwise returns the card's own text
func (g *Game) GetInputValue(cardID, portName string) string {
	// Find if there's an arrow connected to this input
	for _, arrow := range g.arrows {
		if arrow.ToCardID == cardID && arrow.ToPort == portName {
			sourceCard := g.getCardByID(arrow.FromCardID)
			if sourceCard != nil {
				return sourceCard.Text
			}
		}
	}

	// No connection, return the card's own text
	card := g.getCardByID(cardID)
	if card != nil {
		return card.Text
	}
	return ""
}

// Host methods required by input.Host interface
func (g *Game) ScreenToWorld(sx, sy float64) (float64, float64) {
	return g.screenToWorld(sx, sy)
}

func (g *Game) IsMouseOver(mx, my int) bool {
	return g.ui.IsMouseOver(mx, my)
}

func (g *Game) RequestScreenshot() {
	g.screenshotRequested = true
}

func (g *Game) RunEngine() {
	if g.engine != nil {
		g.engine.Run()
	}
}

func (g *Game) SaveState(filename string) error {
	return SaveState(g, filename)
}

func (g *Game) GetCardAt(wx, wy float64) interface{} {
	c := g.getCardAt(wx, wy)
	if c == nil {
		return nil
	}
	return c
}

func (g *Game) AddTextCardHandle(wx, wy float64) interface{} {
	return g.AddTextCard(wx, wy)
}

func (g *Game) DeleteCardHandle(card interface{}) {
	if c, ok := card.(*Card); ok {
		g.DeleteCard(c)
	}
}

func (g *Game) DuplicateCardHandle(card interface{}) {
	if c, ok := card.(*Card); ok {
		g.DuplicateCard(c)
	}
}

func (g *Game) IsInputPortConnectedHandle(cardID, portName string) bool {
	return g.IsInputPortConnected(cardID, portName)
}

func (g *Game) GetCardID(card interface{}) string {
	if c, ok := card.(*Card); ok {
		return c.ID
	}
	return ""
}

func (g *Game) GetCardTitle(card interface{}) string {
	if c, ok := card.(*Card); ok {
		return c.Title
	}
	return ""
}

func (g *Game) GetCardBounds(card interface{}) (float64, float64, float64, float64) {
	if c, ok := card.(*Card); ok {
		return c.X, c.Y, c.Width, c.Height
	}
	return 0, 0, 0, 0
}

func (g *Game) SetCardBounds(card interface{}, x, y, w, h float64) {
	if c, ok := card.(*Card); ok {
		c.X = x
		c.Y = y
		c.Width = w
		c.Height = h
	}
}

func (g *Game) GetCornerAt(card interface{}, wx, wy, zoom float64) int {
	if c, ok := card.(*Card); ok {
		return c.GetCornerAt(wx, wy, zoom)
	}
	return -1
}

func (g *Game) GetPortAt(card interface{}, wx, wy, zoom float64) *input.PortInfo {
	if c, ok := card.(*Card); ok {
		p := c.GetPortAt(wx, wy, zoom)
		if p == nil {
			return nil
		}
		return &input.PortInfo{Name: p.Name, Type: p.Type, IsInput: p.IsInput}
	}
	return nil
}

func (g *Game) GetOutputPortPosition(card interface{}, portName string) (float64, float64) {
	if c, ok := card.(*Card); ok {
		return c.GetOutputPortPosition(portName)
	}
	return 0, 0
}

// ApplyPan applies a pan delta in screen pixels; called from input system.
func (g *Game) ApplyPan(dx, dy float64) {
	g.camera.X -= dx / g.camera.Zoom
	g.camera.Y -= dy / g.camera.Zoom

	if g.camera.X < CameraLimitMin {
		g.camera.X = CameraLimitMin
	}
	if g.camera.Y < CameraLimitMin {
		g.camera.Y = CameraLimitMin
	}
}

// CheckActionButton checks if world coordinates wx,wy are clicking on an action button
// Returns "delete", "duplicate", or "" if no button is clicked
func (g *Game) CheckActionButton(card interface{}, wx, wy float64) string {
	c, ok := card.(*Card)
	if !ok {
		return ""
	}

	// X button (Delete)
	if wx >= c.X+(c.Width-CardActionButtonWidth-5) && wx <= c.X+c.Width-5 &&
		wy >= c.Y+5 && wy <= c.Y+5+CardActionButtonHeight {
		return "delete"
	}

	// ++ button (Duplicate)
	if wx >= c.X+(c.Width-2*CardActionButtonWidth-10) && wx <= c.X+(c.Width-CardActionButtonWidth-10) &&
		wy >= c.Y+5 && wy <= c.Y+5+CardActionButtonHeight {
		return "duplicate"
	}

	return ""
}

func (g *Game) RegisterSubscriptionHandle(fromID, toID, toPort string) {
	g.RegisterSubscription(fromID, toID, toPort)
}

func (g *Game) UnregisterSubscriptionHandle(fromID, toID, toPort string) {
	g.UnregisterSubscription(fromID, toID, toPort)
}

func (g *Game) PropagateTextByID(cardID string) {
	c := g.getCardByID(cardID)
	if c != nil {
		g.PropagateText(c)
	}
}
