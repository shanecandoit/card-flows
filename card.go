package main

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Port represents an input or output on a block
type Port struct {
	Name string
	Type string
}

// Card represents a node on the canvas
type Card struct {
	ID            string
	X, Y          float64
	Width, Height float64
	Color         color.Color
	Title         string
	Text          string
	IsEditing     bool
	IsCommit      bool
	Inputs        []Port
	Outputs       []Port
}

func (g *Game) AddTextCard(x, y float64) *Card {
	card := &Card{
		ID:     NewID(),
		X:      math.Round(x/SnapGridLarge) * SnapGridLarge,
		Y:      math.Round(y/SnapGridLarge) * SnapGridLarge,
		Width:  DefaultCardWidth,
		Height: DefaultCardHeight,
		Color:  ColorCardDefault,
		Title:  "Text Card",
		Text:   "",
		Inputs: []Port{
			{Name: "text", Type: "string"},
		},
		Outputs: []Port{
			{Name: "text", Type: "string"},
		},
	}
	g.cards = append(g.cards, card)
	return card
}

func (c *Card) Draw(screen *ebiten.Image, g *Game, cw, ch float64, hovered bool) {
	sx, sy := g.camera.WorldToScreen(c.X, c.Y, cw, ch)
	sw := c.Width * g.camera.Zoom
	sh := c.Height * g.camera.Zoom

	mx, my := ebiten.CursorPosition()
	wx, wy := g.camera.ScreenToWorld(float64(mx), float64(my), cw, ch)

	c.drawBody(screen, g, sx, sy, sw, sh)
	c.drawSelectionBorder(screen, g, hovered, sx, sy, sw, sh)
	c.drawResizeHandles(screen, g, hovered, wx, wy, cw, ch)
	c.drawHeader(screen, g, sx, sy, sw, wx, wy)

	headerHeight := HeaderHeight
	footerHeight := 0.0
	if len(c.Outputs) > 0 {
		footerHeight = FooterHeight
	}

	c.drawContent(screen, g, sx, sy, headerHeight)
	c.drawDividers(screen, g, sx, sy, sw, sh, headerHeight, footerHeight, cw, ch)
	c.drawPorts(screen, g, sx, sy, sw, sh, headerHeight, footerHeight, cw, ch)
}

func (c *Card) drawBody(screen *ebiten.Image, g *Game, sx, sy, sw, sh float64) {
	// Shadow
	zoom := g.camera.Zoom
	vector.DrawFilledRect(screen, float32(sx+ShadowOffset*zoom), float32(sy+ShadowOffset*zoom), float32(sw), float32(sh), ColorShadow, false)
	// Body
	vector.DrawFilledRect(screen, float32(sx), float32(sy), float32(sw), float32(sh), c.Color, false)
}

func (c *Card) drawSelectionBorder(screen *ebiten.Image, g *Game, hovered bool, sx, sy, sw, sh float64) {
	borderColor := color.RGBA{0, 0, 0, 0}
	showBorder := false

	if c == g.input.activeCard {
		showBorder = true
		if g.input.isHot {
			borderColor = ColorCardHot
		} else {
			borderColor = ColorCardActive
		}
	} else if hovered && g.input.activeCard == nil {
		showBorder = true
		borderColor = ColorCardHover
	}

	if showBorder {
		zoom := g.camera.Zoom
		thickness := float32(BorderThickness * zoom)
		offset := float32(BorderOffset * zoom)
		vector.StrokeRect(screen,
			float32(sx)-offset-thickness/2,
			float32(sy)-offset-thickness/2,
			float32(sw)+2*(offset+thickness/2),
			float32(sh)+2*(offset+thickness/2),
			thickness, borderColor, false)
	}
}

func (c *Card) drawResizeHandles(screen *ebiten.Image, g *Game, hovered bool, wx, wy float64, cw, ch float64) {
	resizingThis := (c == g.input.resizingCard)
	hCorner := c.GetCornerAt(wx, wy, g.camera.Zoom)
	if (hovered && hCorner != -1) || resizingThis {
		cIdx := hCorner
		if resizingThis {
			cIdx = g.input.resizingCorner
		}

		var cx, cy float64
		switch cIdx {
		case 0:
			cx, cy = c.X, c.Y
		case 1:
			cx, cy = c.X+c.Width, c.Y
		case 2:
			cx, cy = c.X, c.Y+c.Height
		case 3:
			cx, cy = c.X+c.Width, c.Y+c.Height
		}

		scx, scy := g.camera.WorldToScreen(cx, cy, cw, ch)
		radius := float32(CornerRadius * g.camera.Zoom)
		vector.DrawFilledCircle(screen, float32(scx), float32(scy), radius, ColorCornerHandle, false)
		vector.StrokeCircle(screen, float32(scx), float32(scy), radius, 2, ColorCardHover, false)
	}
}

func (c *Card) drawHeader(screen *ebiten.Image, g *Game, sx, sy, sw float64, wx, wy float64) {
	// Title
	msg := fmt.Sprintf("%s\n(%.0f, %.0f)", c.Title, c.X, c.Y)
	ebitenutil.DebugPrintAt(screen, msg, int(sx+5), int(sy+5))

	// Buttons
	zoom := g.camera.Zoom
	btnW := CardActionButtonWidth * zoom
	btnH := CardActionButtonHeight * zoom
	btnMargin := 5.0 * zoom

	// X (Delete)
	xBtnX := sx + sw - btnW - btnMargin
	xBtnY := sy + btnMargin
	xHover := wx >= c.X+(c.Width-CardActionButtonWidth-5) && wx <= c.X+c.Width-5 &&
		wy >= c.Y+5 && wy <= c.Y+5+CardActionButtonHeight

	xColor := ColorButtonBackground
	xColor.A = 30
	if xHover {
		xColor = ColorCardActionDelete
		xColor.A = 200
	}
	vector.DrawFilledRect(screen, float32(xBtnX), float32(xBtnY), float32(btnW), float32(btnH), xColor, false)
	ebitenutil.DebugPrintAt(screen, "X", int(xBtnX+btnW/2-4), int(xBtnY+btnH/2-8))

	// ++ (Duplicate)
	dBtnX := xBtnX - btnW - btnMargin
	dBtnY := sy + btnMargin
	dHover := wx >= c.X+(c.Width-2*CardActionButtonWidth-10) && wx <= c.X+(c.Width-CardActionButtonWidth-10) &&
		wy >= c.Y+5 && wy <= c.Y+5+CardActionButtonHeight

	dColor := ColorButtonBackground
	dColor.A = 30
	if dHover {
		dColor = ColorCardActionDuplicate
		dColor.A = 200
	}
	vector.DrawFilledRect(screen, float32(dBtnX), float32(dBtnY), float32(btnW), float32(btnH), dColor, false)
	ebitenutil.DebugPrintAt(screen, "++", int(dBtnX+btnW/2-6), int(dBtnY+btnH/2-8))
}

func (c *Card) drawContent(screen *ebiten.Image, g *Game, sx, sy, headerHeight float64) {
	var textContent string

	// Special handling for TextCard
	if c.Title == "Text Card" {
		isPortConnected := g.IsInputPortConnected(c.ID, "text")
		if isPortConnected {
			// Later, this will show the actual input value from the execution engine
			textContent = "[Connected]"
		} else {
			textContent = c.Text
			if c.IsEditing {
				if (time.Now().UnixMilli()/CursorBlinkRate)%2 == 0 {
					textContent += "|"
				}
			}
		}
	} else {
		textContent = c.Text
	}

	ebitenutil.DebugPrintAt(screen, textContent, int(sx+10), int(sy+headerHeight+10))
}

func (c *Card) drawDividers(screen *ebiten.Image, g *Game, sx, sy, sw, sh, headerHeight, footerHeight float64, cw, ch float64) {
	color := ColorDivider
	// Header divider
	hy := c.Y + headerHeight
	shx1, shy := g.camera.WorldToScreen(c.X, hy, cw, ch)
	shx2, _ := g.camera.WorldToScreen(c.X+c.Width, hy, cw, ch)
	vector.StrokeLine(screen, float32(shx1), float32(shy), float32(shx2), float32(shy), 1, color, false)

	// Footer divider
	if footerHeight > 0 {
		fy := c.Y + c.Height - footerHeight
		sfx1, sfy := g.camera.WorldToScreen(c.X, fy, cw, ch)
		sfx2, _ := g.camera.WorldToScreen(c.X+c.Width, fy, cw, ch)
		vector.StrokeLine(screen, float32(sfx1), float32(sfy), float32(sfx2), float32(sfy), 1, color, false)
	}
}

func (c *Card) drawPorts(screen *ebiten.Image, g *Game, sx, sy, sw, sh, headerHeight, footerHeight float64, cw, ch float64) {
	zoom := g.camera.Zoom
	portSize := PortSize * zoom

	// Inputs (Left)
	if len(c.Inputs) > 0 {
		usableHeight := c.Height - headerHeight - footerHeight
		ySpacing := usableHeight / float64(len(c.Inputs)+1)
		for i, port := range c.Inputs {
			py := c.Y + headerHeight + ySpacing*float64(i+1)
			spx, spy := g.camera.WorldToScreen(c.X, py, cw, ch)
			vector.DrawFilledRect(screen, float32(spx-portSize/2), float32(spy-portSize/2), float32(portSize), float32(portSize), ColorPortBody, false)
			vector.DrawFilledCircle(screen, float32(spx), float32(spy), float32(3*zoom), ColorPortDot, false)
			label := fmt.Sprintf("%s:%s", port.Name, port.Type)
			ebitenutil.DebugPrintAt(screen, label, int(spx+portSize), int(spy-8*zoom))
		}
	}

	// Outputs (Bottom)
	if len(c.Outputs) > 0 {
		xSpacing := c.Width / float64(len(c.Outputs)+1)
		for i, port := range c.Outputs {
			px := c.X + xSpacing*float64(i+1)
			spx, spy := g.camera.WorldToScreen(px, c.Y+c.Height, cw, ch)
			vector.DrawFilledRect(screen, float32(spx-portSize/2), float32(spy-portSize/2), float32(portSize), float32(portSize), ColorPortBody, false)
			vector.DrawFilledCircle(screen, float32(spx), float32(spy), float32(3*zoom), ColorPortDot, false)
			label := fmt.Sprintf("%s:%s", port.Name, port.Type)
			ebitenutil.DebugPrintAt(screen, label, int(spx-20*zoom), int(spy-20*zoom))
		}
	}
}

func (c *Card) GetCornerAt(wx, wy, zoom float64) int {
	threshold := CornerThreshold // world units
	corners := [][2]float64{
		{c.X, c.Y},
		{c.X + c.Width, c.Y},
		{c.X, c.Y + c.Height},
		{c.X + c.Width, c.Y + c.Height},
	}

	for i, cor := range corners {
		dx := wx - cor[0]
		dy := wy - cor[1]
		if math.Sqrt(dx*dx+dy*dy) < threshold/zoom {
			return i
		}
	}
	return -1
}

func (c *Card) GetInputPortPosition(name string) (float64, float64) {
	index := -1
	for i, p := range c.Inputs {
		if p.Name == name {
			index = i
			break
		}
	}
	if index == -1 {
		return c.X, c.Y // Fallback
	}

	headerHeight := HeaderHeight
	footerHeight := 0.0
	if len(c.Outputs) > 0 {
		footerHeight = FooterHeight
	}
	usableHeight := c.Height - headerHeight - footerHeight
	ySpacing := usableHeight / float64(len(c.Inputs)+1)
	py := c.Y + headerHeight + ySpacing*float64(index+1)
	return c.X, py
}

func (c *Card) GetOutputPortPosition(name string) (float64, float64) {
	index := -1
	for i, p := range c.Outputs {
		if p.Name == name {
			index = i
			break
		}
	}
	if index == -1 {
		return c.X + c.Width, c.Y + c.Height // Fallback
	}

	xSpacing := c.Width / float64(len(c.Outputs)+1)
	px := c.X + xSpacing*float64(index+1)
	return px, c.Y + c.Height
}

// PortInfo contains info about a hit port
type PortInfo struct {
	Name    string
	IsInput bool
	Type    string
}

func (c *Card) GetPortAt(wx, wy, zoom float64) *PortInfo {
	// portSizeWorld := PortSize // / zoom? No, PortSize is screen size constant?
	// constants are in pixels probably.
	// Wait, PortSize is 10.

	hitThreshold := CornerThreshold / zoom // Use same threshold metric logic as corners

	// Check Inputs
	for _, p := range c.Inputs {
		px, py := c.GetInputPortPosition(p.Name)
		dx := wx - px
		dy := wy - py
		if math.Sqrt(dx*dx+dy*dy) < hitThreshold {
			return &PortInfo{Name: p.Name, IsInput: true, Type: p.Type}
		}
	}

	// Check Outputs
	for _, p := range c.Outputs {
		px, py := c.GetOutputPortPosition(p.Name)
		dx := wx - px
		dy := wy - py
		if math.Sqrt(dx*dx+dy*dy) < hitThreshold {
			return &PortInfo{Name: p.Name, IsInput: false, Type: p.Type}
		}
	}

	return nil
}
