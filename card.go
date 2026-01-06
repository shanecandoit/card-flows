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
		X:      math.Round(x/SnapGridLarge) * SnapGridLarge,
		Y:      math.Round(y/SnapGridLarge) * SnapGridLarge,
		Width:  DefaultCardWidth,
		Height: DefaultCardHeight,
		Color:  ColorCardDefault,
		Title:  "Text Card",
		Text:   "",
		Outputs: []Port{
			{Name: "text", Type: "string"},
		},
	}
	g.cards = append(g.cards, card)
	return card
}

func (c *Card) Draw(screen *ebiten.Image, g *Game, cw, ch float64, hovered bool) {
	screenX, screenY := g.worldToScreen(c.X, c.Y, cw, ch)
	screenW := c.Width * g.camera.Zoom
	screenH := c.Height * g.camera.Zoom

	// Shadow
	vector.DrawFilledRect(screen, float32(screenX+ShadowOffset*g.camera.Zoom), float32(screenY+ShadowOffset*g.camera.Zoom), float32(screenW), float32(screenH), ColorShadow, false)

	// Body
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(screenW), float32(screenH), c.Color, false)

	// Border Logic
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
		borderThickness := float32(BorderThickness * g.camera.Zoom)
		borderOffset := float32(BorderOffset * g.camera.Zoom)
		vector.StrokeRect(screen,
			float32(screenX)-borderOffset-borderThickness/2,
			float32(screenY)-borderOffset-borderThickness/2,
			float32(screenW)+2*(borderOffset+borderThickness/2),
			float32(screenH)+2*(borderOffset+borderThickness/2),
			borderThickness, borderColor, false)
	}

	// Draw Corner Handle if hovering or resizing
	mx, my := ebiten.CursorPosition()
	wx, wy := g.screenToWorld(float64(mx), float64(my))
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

		scx, scy := g.worldToScreen(cx, cy, cw, ch)
		radius := float32(CornerRadius * g.camera.Zoom)
		vector.DrawFilledCircle(screen, float32(scx), float32(scy), radius, ColorCornerHandle, false)
		vector.StrokeCircle(screen, float32(scx), float32(scy), radius, 2, ColorCardHover, false)
	}

	// Title
	msg := fmt.Sprintf("%s\n(%.0f, %.0f)", c.Title, c.X, c.Y)
	ebitenutil.DebugPrintAt(screen, msg, int(screenX+5), int(screenY+5))

	// --- Action Buttons (dup and X) ---
	mx, my = ebiten.CursorPosition()
	wx, wy = g.screenToWorld(float64(mx), float64(my))

	btnW := CardActionButtonWidth * g.camera.Zoom
	btnH := CardActionButtonHeight * g.camera.Zoom
	btnMargin := 5.0 * g.camera.Zoom

	// X button (Delete)
	xBtnX := screenX + screenW - btnW - btnMargin
	xBtnY := screenY + btnMargin
	xHover := wx >= c.X+(c.Width-CardActionButtonWidth-5) && wx <= c.X+c.Width-5 &&
		wy >= c.Y+5 && wy <= c.Y+5+CardActionButtonHeight

	xColor := ColorCardActionDelete
	if xHover {
		xColor.A = 200
	} else {
		xColor.A = 100
	}
	vector.DrawFilledRect(screen, float32(xBtnX), float32(xBtnY), float32(btnW), float32(btnH), xColor, false)
	ebitenutil.DebugPrintAt(screen, "X", int(xBtnX+btnW/2-4), int(xBtnY+btnH/2-8))

	// dup button (Duplicate)
	dBtnX := xBtnX - btnW - btnMargin
	dBtnY := screenY + btnMargin
	dHover := wx >= c.X+(c.Width-2*CardActionButtonWidth-10) && wx <= c.X+(c.Width-CardActionButtonWidth-10) &&
		wy >= c.Y+5 && wy <= c.Y+5+CardActionButtonHeight

	dColor := ColorCardActionDuplicate
	if dHover {
		dColor.A = 200
	} else {
		dColor.A = 100
	}
	vector.DrawFilledRect(screen, float32(dBtnX), float32(dBtnY), float32(btnW), float32(btnH), dColor, false)
	ebitenutil.DebugPrintAt(screen, "++", int(dBtnX+btnW/2-6), int(dBtnY+btnH/2-8))

	// --- Text Content ---
	headerHeight := HeaderHeight
	textContent := c.Text
	if c.IsEditing {
		// Blinking cursor
		if (time.Now().UnixMilli()/CursorBlinkRate)%2 == 0 {
			textContent += "|"
		}
	}

	// Simple text wrapping / positioning
	// For now, just print it in the middle
	ebitenutil.DebugPrintAt(screen, textContent, int(screenX+10), int(screenY+headerHeight+10))

	// --- Dividers ---
	dividerColor := ColorDivider
	hy := c.Y + headerHeight
	shx1, shy := g.worldToScreen(c.X, hy, cw, ch)
	shx2, _ := g.worldToScreen(c.X+c.Width, hy, cw, ch)
	vector.StrokeLine(screen, float32(shx1), float32(shy), float32(shx2), float32(shy), 1, dividerColor, false)

	footerHeight := 0.0
	if len(c.Outputs) > 0 {
		footerHeight = FooterHeight
		fy := c.Y + c.Height - footerHeight
		sfx1, sfy := g.worldToScreen(c.X, fy, cw, ch)
		sfx2, _ := g.worldToScreen(c.X+c.Width, fy, cw, ch)
		vector.StrokeLine(screen, float32(sfx1), float32(sfy), float32(sfx2), float32(sfy), 1, dividerColor, false)
	}

	// --- Ports Rendering ---
	portSize := PortSize * g.camera.Zoom

	// Inputs (Left edge)
	if len(c.Inputs) > 0 {
		usableHeight := c.Height - headerHeight - footerHeight
		ySpacing := usableHeight / float64(len(c.Inputs)+1)
		for i, port := range c.Inputs {
			py := c.Y + headerHeight + ySpacing*float64(i+1)
			spx, spy := g.worldToScreen(c.X, py, cw, ch)
			vector.DrawFilledRect(screen, float32(spx-portSize/2), float32(spy-portSize/2), float32(portSize), float32(portSize), ColorPortBody, false)
			vector.DrawFilledCircle(screen, float32(spx), float32(spy), float32(3*g.camera.Zoom), ColorPortDot, false)
			label := fmt.Sprintf("%s:%s", port.Name, port.Type)
			ebitenutil.DebugPrintAt(screen, label, int(spx+portSize), int(spy-8*g.camera.Zoom))
		}
	}

	// Outputs (Bottom edge)
	if len(c.Outputs) > 0 {
		xSpacing := c.Width / float64(len(c.Outputs)+1)
		for i, port := range c.Outputs {
			px := c.X + xSpacing*float64(i+1)
			spx, spy := g.worldToScreen(px, c.Y+c.Height, cw, ch)
			vector.DrawFilledRect(screen, float32(spx-portSize/2), float32(spy-portSize/2), float32(portSize), float32(portSize), ColorPortBody, false)
			vector.DrawFilledCircle(screen, float32(spx), float32(spy), float32(3*g.camera.Zoom), ColorPortDot, false)
			label := fmt.Sprintf("%s:%s", port.Name, port.Type)
			ebitenutil.DebugPrintAt(screen, label, int(spx-20*g.camera.Zoom), int(spy-20*g.camera.Zoom))
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
