package main

import (
	"image/color"
	"os"

	"gopkg.in/yaml.v3"
)

type ColorState struct {
	R uint8 `yaml:"r"`
	G uint8 `yaml:"g"`
	B uint8 `yaml:"b"`
	A uint8 `yaml:"a"`
}

type PortState struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type CardState struct {
	X       float64     `yaml:"x"`
	Y       float64     `yaml:"y"`
	Width   float64     `yaml:"width"`
	Height  float64     `yaml:"height"`
	Color   ColorState  `yaml:"color"`
	Title   string      `yaml:"title"`
	Text    string      `yaml:"text"`
	Inputs  []PortState `yaml:"inputs"`
	Outputs []PortState `yaml:"outputs"`
}

type CameraState struct {
	X    float64 `yaml:"x"`
	Y    float64 `yaml:"y"`
	Zoom float64 `yaml:"zoom"`
}

type AppState struct {
	Cards  []CardState `yaml:"cards"`
	Camera CameraState `yaml:"camera"`
}

func SaveState(g *Game, filename string) error {
	state := AppState{
		Camera: CameraState{
			X:    g.camera.X,
			Y:    g.camera.Y,
			Zoom: g.camera.Zoom,
		},
	}

	for _, c := range g.cards {
		r, g_val, b, a := c.Color.RGBA()
		cardState := CardState{
			X:      c.X,
			Y:      c.Y,
			Width:  c.Width,
			Height: c.Height,
			Color: ColorState{
				R: uint8(r >> 8),
				G: uint8(g_val >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			},
			Title: c.Title,
			Text:  c.Text,
		}
		for _, p := range c.Inputs {
			cardState.Inputs = append(cardState.Inputs, PortState{Name: p.Name, Type: p.Type})
		}
		for _, p := range c.Outputs {
			cardState.Outputs = append(cardState.Outputs, PortState{Name: p.Name, Type: p.Type})
		}
		state.Cards = append(state.Cards, cardState)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	err = enc.Encode(&state)
	if err != nil {
		return err
	}
	return enc.Close()
}

func LoadState(g *Game, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var state AppState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return err
	}

	g.camera.X = state.Camera.X
	g.camera.Y = state.Camera.Y
	g.camera.Zoom = state.Camera.Zoom

	g.cards = nil
	for _, cs := range state.Cards {
		card := &Card{
			X:      cs.X,
			Y:      cs.Y,
			Width:  cs.Width,
			Height: cs.Height,
			Color:  color.RGBA{cs.Color.R, cs.Color.G, cs.Color.B, cs.Color.A},
			Title:  cs.Title,
			Text:   cs.Text,
		}
		for _, ps := range cs.Inputs {
			card.Inputs = append(card.Inputs, Port{Name: ps.Name, Type: ps.Type})
		}
		for _, ps := range cs.Outputs {
			card.Outputs = append(card.Outputs, Port{Name: ps.Name, Type: ps.Type})
		}
		g.cards = append(g.cards, card)
	}

	return nil
}
