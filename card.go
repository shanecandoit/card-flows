package main

import (
	"fmt"
	"image/color"
	"math"

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
	Inputs        []Port
	Outputs       []Port
}

func (c *Card) Draw(screen *ebiten.Image, g *Game, cw, ch float64, hovered bool) {
	screenX, screenY := g.worldToScreen(c.X, c.Y, cw, ch)
	screenW := c.Width * g.camera.Zoom
	screenH := c.Height * g.camera.Zoom

	// Shadow
	vector.DrawFilledRect(screen, float32(screenX+5*g.camera.Zoom), float32(screenY+5*g.camera.Zoom), float32(screenW), float32(screenH), color.RGBA{0, 0, 0, 100}, false)

	// Body
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(screenW), float32(screenH), c.Color, false)

	// Border Logic
	borderColor := color.RGBA{0, 0, 0, 0}
	showBorder := false

	if c == g.input.activeCard {
		showBorder = true
		if g.input.isHot {
			borderColor = color.RGBA{255, 140, 0, 255} // Orange
		} else {
			borderColor = color.RGBA{50, 205, 50, 255} // Green
		}
	} else if hovered && g.input.activeCard == nil {
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
		radius := float32(6 * g.camera.Zoom)
		vector.DrawFilledCircle(screen, float32(scx), float32(scy), radius, color.RGBA{255, 255, 255, 200}, false)
		vector.StrokeCircle(screen, float32(scx), float32(scy), radius, 2, color.RGBA{0, 120, 255, 255}, false)
	}

	// Title
	msg := fmt.Sprintf("%s\n(%.0f, %.0f)", c.Title, c.X, c.Y)
	ebitenutil.DebugPrintAt(screen, msg, int(screenX+5), int(screenY+5))

	// --- Dividers ---
	dividerColor := color.RGBA{0, 0, 0, 50}
	headerHeight := 50.0
	hy := c.Y + headerHeight
	shx1, shy := g.worldToScreen(c.X, hy, cw, ch)
	shx2, _ := g.worldToScreen(c.X+c.Width, hy, cw, ch)
	vector.StrokeLine(screen, float32(shx1), float32(shy), float32(shx2), float32(shy), 1, dividerColor, false)

	footerHeight := 0.0
	if len(c.Outputs) > 0 {
		footerHeight = 30.0
		fy := c.Y + c.Height - footerHeight
		sfx1, sfy := g.worldToScreen(c.X, fy, cw, ch)
		sfx2, _ := g.worldToScreen(c.X+c.Width, fy, cw, ch)
		vector.StrokeLine(screen, float32(sfx1), float32(sfy), float32(sfx2), float32(sfy), 1, dividerColor, false)
	}

	// --- Ports Rendering ---
	portSize := 10.0 * g.camera.Zoom

	// Inputs (Left edge)
	if len(c.Inputs) > 0 {
		usableHeight := c.Height - headerHeight - footerHeight
		ySpacing := usableHeight / float64(len(c.Inputs)+1)
		for i, port := range c.Inputs {
			py := c.Y + headerHeight + ySpacing*float64(i+1)
			spx, spy := g.worldToScreen(c.X, py, cw, ch)
			vector.DrawFilledRect(screen, float32(spx-portSize/2), float32(spy-portSize/2), float32(portSize), float32(portSize), color.RGBA{150, 150, 150, 255}, false)
			vector.DrawFilledCircle(screen, float32(spx), float32(spy), float32(3*g.camera.Zoom), color.RGBA{0, 0, 0, 255}, false)
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
			vector.DrawFilledRect(screen, float32(spx-portSize/2), float32(spy-portSize/2), float32(portSize), float32(portSize), color.RGBA{150, 150, 150, 255}, false)
			vector.DrawFilledCircle(screen, float32(spx), float32(spy), float32(3*g.camera.Zoom), color.RGBA{0, 0, 0, 255}, false)
			label := fmt.Sprintf("%s:%s", port.Name, port.Type)
			ebitenutil.DebugPrintAt(screen, label, int(spx-20*g.camera.Zoom), int(spy-20*g.camera.Zoom))
		}
	}
}

func (c *Card) GetCornerAt(wx, wy, zoom float64) int {
	threshold := 15.0 // world units
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
