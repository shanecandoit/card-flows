package main

import (
	"image/color"
	"os"
	"strings"

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
	ID      string      `yaml:"id"`
	Type    string      `yaml:"type"`
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

type ArrowState struct {
	FromCardID string `yaml:"from_card_id"`
	FromPort   string `yaml:"from_port"`
	ToCardID   string `yaml:"to_card_id"`
	ToPort     string `yaml:"to_port"`
}

type AppState struct {
	Cards  []CardState  `yaml:"cards"`
	Arrows []ArrowState `yaml:"arrows"`
	Camera CameraState  `yaml:"camera"`
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
			ID:     c.ID,
			Type:   c.Type,
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

	for _, arrow := range g.arrows {
		state.Arrows = append(state.Arrows, ArrowState{
			FromCardID: arrow.FromCardID,
			FromPort:   arrow.FromPort,
			ToCardID:   arrow.ToCardID,
			ToPort:     arrow.ToPort,
		})
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
	g.arrows = nil

	for _, cs := range state.Cards {
		id := cs.ID
		if id == "" {
			id = NewID()
		}

		// Migration: Infer Type from Title if not set
		cardType := cs.Type
		if cardType == "" {
			if strings.HasPrefix(cs.Title, "Text Card") {
				cardType = "text"
			} else if cs.Title == "String:find_replace" {
				cardType = "find_replace"
			}
		}

		card := &Card{
			ID:     id,
			Type:   cardType,
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

		// Migration: Ensure Text Cards have the default output if missing
		if card.Title == "Text Card" && len(card.Outputs) == 0 {
			card.Outputs = append(card.Outputs, Port{Name: "text", Type: "string"})
		}

		g.cards = append(g.cards, card)
	}

	// Load Arrows
	// Note: If IDs changed during load (due to empty IDs), arrows will be broken.
	// We filter out any arrows that point to non-existent cards to prevent "Miss Draw" errors.
	validArrows := []*Arrow{}
	for _, as := range state.Arrows {
		// Check validity
		fromExists := false
		toExists := false
		for _, c := range g.cards {
			if c.ID == as.FromCardID {
				fromExists = true
			}
			if c.ID == as.ToCardID {
				toExists = true
			}
		}

		if fromExists && toExists {
			arrow := &Arrow{
				FromCardID: as.FromCardID,
				FromPort:   as.FromPort,
				ToCardID:   as.ToCardID,
				ToPort:     as.ToPort,
				Color:      ColorArrowDefault,
			}
			validArrows = append(validArrows, arrow)
		} else {
			// Optional: Log dropped arrow?
			// fmt.Printf("Dropping invalid arrow: %s->%s\n", as.FromCardID, as.ToCardID)
		}
	}
	g.arrows = validArrows

	return nil
}
