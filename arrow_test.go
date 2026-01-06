package main

import (
	"strings"
	"testing"
)

func TestDuplicateCardIDs(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{} // Clear default cards
	g.arrows = []*Arrow{}

	c1 := g.AddTextCard(100, 100)

	g.DuplicateCard(c1)
	if len(g.cards) != 2 {
		t.Fatalf("Expected 2 cards, got %d", len(g.cards))
	}

	c2 := g.cards[1]

	if c1.ID == c2.ID {
		t.Errorf("Duplicated card shares ID: %s", c1.ID)
	}
	// Check for new ID suffix format "Title (12345)"
	// We check if it ends with ')' and contains the suffix
	if !strings.HasSuffix(c2.Title, ")") || !strings.Contains(c2.Title, "(") {
		t.Errorf("Unexpected title format: %s", c2.Title)
	}
}

func TestWiring(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{}
	g.arrows = []*Arrow{} // Clear arrows
	c1 := g.AddTextCard(100, 100)
	c2 := g.AddTextCard(300, 100)

	// Simulate InputSystem wiring logic manually
	// We don't have direct access to internal InputSystem state logic easily without mocking events,
	// but we can test the core "connect" logic if we extract it, or reproduce it.
	// For now, let's replicate the logic from input.go's handleWiring drop section

	// Create Arrow C1 -> C2
	arrow := &Arrow{
		FromCardID: c1.ID,
		FromPort:   "text",
		ToCardID:   c2.ID,
		ToPort:     "text",
	}
	g.arrows = append(g.arrows, arrow)

	if len(g.arrows) != 1 {
		t.Errorf("Expected 1 arrow, got %d", len(g.arrows))
	}

	if !g.IsInputPortConnected(c2.ID, "text") {
		t.Errorf("Expected c2 input to be connected")
	}
}

func TestWiringOverwrite(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{}
	g.arrows = []*Arrow{} // Clear arrows
	c1 := g.AddTextCard(100, 100)
	c2 := g.AddTextCard(300, 100)
	c3 := g.AddTextCard(100, 300)

	// wire c1 -> c2
	g.arrows = append(g.arrows, &Arrow{
		FromCardID: c1.ID,
		FromPort:   "text",
		ToCardID:   c2.ID,
		ToPort:     "text",
	})

	// Logic to overwrite: Remove existing arrows to c2:text
	// This mirrors input.go logic
	targetCard := c2
	targetPort := "text"

	filteredArrows := g.arrows[:0]
	for _, arrow := range g.arrows {
		if arrow.ToCardID == targetCard.ID && arrow.ToPort == targetPort {
			continue
		}
		filteredArrows = append(filteredArrows, arrow)
	}
	g.arrows = filteredArrows

	// wire c3 -> c2
	g.arrows = append(g.arrows, &Arrow{
		FromCardID: c3.ID,
		FromPort:   "text",
		ToCardID:   c2.ID,
		ToPort:     "text",
	})

	if len(g.arrows) != 1 {
		t.Errorf("Expected 1 arrow after overwrite, got %d", len(g.arrows))
	}

	if g.arrows[0].FromCardID != c3.ID {
		t.Errorf("Expected arrow source to be c3, got %s", g.arrows[0].FromCardID)
	}
}
