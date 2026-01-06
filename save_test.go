package main

import (
	"os"
	"testing"
)

func TestSaveLoadArrows(t *testing.T) {
	// Setup
	filename := "test_state.yaml"
	defer os.Remove(filename)

	g := NewGame()
	g.cards = []*Card{} // Clear default cards
	g.arrows = []*Arrow{}

	c1 := g.AddTextCard(100, 100)
	c2 := g.AddTextCard(300, 100)

	// Create Arrow
	arrow := &Arrow{
		FromCardID: c1.ID,
		FromPort:   "text",
		ToCardID:   c2.ID,
		ToPort:     "text",
	}
	g.arrows = append(g.arrows, arrow)

	// Save
	if err := SaveState(g, filename); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// New Game instance for loading
	g2 := NewGame()
	g2.cards = []*Card{}
	g2.arrows = []*Arrow{}

	// Load
	if err := LoadState(g2, filename); err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	// Verify
	if len(g2.cards) != 2 {
		t.Errorf("Expected 2 cards loaded, got %d", len(g2.cards))
	}

	if len(g2.arrows) != 1 {
		t.Fatalf("Expected 1 arrow loaded, got %d", len(g2.arrows))
	}

	loadedArrow := g2.arrows[0]
	if loadedArrow.FromCardID != c1.ID || loadedArrow.ToCardID != c2.ID {
		t.Errorf("Loaded arrow has incorrect IDs. Expected %s->%s, got %s->%s",
			c1.ID, c2.ID, loadedArrow.FromCardID, loadedArrow.ToCardID)
	}
}
