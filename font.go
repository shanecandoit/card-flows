package main

import (
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
)

// LoadUIFont attempts to load fonts/Roboto-Regular.ttf. If it fails, returns basicfont.Face7x13.
func LoadUIFont() font.Face {
	data, err := os.ReadFile("fonts/Roboto-Regular.ttf")
	if err != nil {
		log.Println("LoadUIFont: fonts/Roboto-Regular.ttf not found, using basic font:", err)
		return basicfont.Face7x13
	}
	f, err := opentype.Parse(data)
	if err != nil {
		log.Println("LoadUIFont: parse error, using basic font:", err)
		return basicfont.Face7x13
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: 16, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		log.Println("LoadUIFont: new face error, using basic font:", err)
		return basicfont.Face7x13
	}
	return face
}

// DrawTextLines draws multiline text with the provided font.Face and color starting at (x,y).
func DrawTextLines(screen *ebiten.Image, face font.Face, s string, x, y int, clr color.Color) {
	if face == nil {
		face = basicfont.Face7x13
	}
	lines := splitLines(s)
	// compute line height and baseline offset from metrics
	metrics := face.Metrics()
	ascent := int(metrics.Ascent >> 6)
	descent := int(metrics.Descent >> 6)
	lineHeight := ascent + descent
	if lineHeight <= 0 {
		lineHeight = 16
		ascent = 12
	}
	// Treat provided y as the top of the first line. text.Draw expects baseline y,
	// so shift by ascent.
	baseY := y + ascent
	for i, line := range lines {
		text.Draw(screen, line, face, x, baseY+(i*lineHeight), clr)
	}
}

func splitLines(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == '\n' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	out = append(out, cur)
	return out
}
