package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Arrow struct {
	FromCardID string
	FromPort   string
	ToCardID   string
	ToPort     string
	Color      color.Color
}

func (a *Arrow) Draw(screen *ebiten.Image, g *Game, cw, ch float64) {
	fromCard := g.getCardByID(a.FromCardID)
	toCard := g.getCardByID(a.ToCardID)

	if fromCard == nil || toCard == nil {
		// fmt.Printf("Miss Draw: %s->%s (%v, %v)\n", a.FromCardID, a.ToCardID, fromCard, toCard)
		return
	}
	// fmt.Printf("Drawing Arrow: %s:%s -> %s:%s\n", fromCard.Title, a.FromPort, toCard.Title, a.ToPort)

	// Calculate start and end points (world coordinates)
	// We need to know the index of the port.
	// For now, let's just use center of card or approximate based on port name if possible.
	// Ideally we'd have a method on Card to get Port World Position.

	startWX, startWY := fromCard.GetOutputPortPosition(a.FromPort)
	endWX, endWY := toCard.GetInputPortPosition(a.ToPort)

	// Convert to screen
	sx, sy := g.camera.WorldToScreen(startWX, startWY, cw, ch)
	ex, ey := g.camera.WorldToScreen(endWX, endWY, cw, ch)

	// Draw Curve
	// Simple cubic bezier: Control points?
	// CP1 = Start + (Right * dist)
	// CP2 = End - (Right * dist) (or Left * dist)
	dist := math.Abs(ex-sx) * 0.5
	if dist < 50 {
		dist = 50
	}

	cp1x, cp1y := sx+dist, sy
	cp2x, cp2y := ex-dist, ey

	// Draw Curve using segments
	segments := 20
	prevX, prevY := float32(sx), float32(sy)

	zoom := g.camera.Zoom
	thickness := float32(2 * zoom)
	if thickness < 1 {
		thickness = 1
	}

	for i := 1; i <= segments; i++ {
		t := float64(i) / float64(segments)

		// Cubic Bezier Formula
		u := 1 - t
		tt := t * t
		uu := u * u
		uuu := uu * u
		ttt := tt * t

		px := uuu*sx + 3*uu*t*cp1x + 3*u*tt*cp2x + ttt*ex
		py := uuu*sy + 3*uu*t*cp1y + 3*u*tt*cp2y + ttt*ey

		curX, curY := float32(px), float32(py)

		vector.StrokeLine(screen, prevX, prevY, curX, curY, thickness, a.Color, true)
		prevX, prevY = curX, curY
	}
}
