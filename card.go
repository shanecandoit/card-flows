package main

import (
	"image/color"
	"math"
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
