package main

import (
	"testing"
)

func TestFindReplaceCard(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{}
	g.arrows = []*Arrow{}

	// Create input text card
	inputCard := g.AddTextCard(100, 100)
	inputCard.Text = "Hello World"

	// Create find/replace card
	frCard := g.AddFindReplaceCard(100, 300)

	// Create "find" text card
	findCard := g.AddTextCard(300, 200)
	findCard.Text = "World"

	// Create "replace" text card
	replaceCard := g.AddTextCard(300, 400)
	replaceCard.Text = "Universe"

	// Wire up: inputCard -> frCard.input
	arrow1 := &Arrow{
		FromCardID: inputCard.ID,
		FromPort:   "text",
		ToCardID:   frCard.ID,
		ToPort:     "input",
		Color:      ColorArrowDefault,
	}
	g.arrows = append(g.arrows, arrow1)

	// Wire up: findCard -> frCard.find
	arrow2 := &Arrow{
		FromCardID: findCard.ID,
		FromPort:   "text",
		ToCardID:   frCard.ID,
		ToPort:     "find",
		Color:      ColorArrowDefault,
	}
	g.arrows = append(g.arrows, arrow2)

	// Wire up: replaceCard -> frCard.replace
	arrow3 := &Arrow{
		FromCardID: replaceCard.ID,
		FromPort:   "text",
		ToCardID:   frCard.ID,
		ToPort:     "replace",
		Color:      ColorArrowDefault,
	}
	g.arrows = append(g.arrows, arrow3)

	// Run the engine
	g.engine.Run()

	// Check the result
	expected := "Hello Universe"
	if frCard.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, frCard.Text)
	}

	// Verify output is in memory
	key := frCard.ID + ":result"
	if val, ok := g.engine.Memory[key]; ok {
		if str, ok := val.(string); ok {
			if str != expected {
				t.Errorf("Expected memory value '%s', got '%s'", expected, str)
			}
		} else {
			t.Error("Memory value is not a string")
		}
	} else {
		t.Error("Result not found in engine memory")
	}
}

func TestFindReplaceCache(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{}
	g.arrows = []*Arrow{}

	// Create input cards
	inputCard := g.AddTextCard(100, 100)
	inputCard.Text = "foo bar"

	findCard := g.AddTextCard(300, 200)
	findCard.Text = "bar"

	replaceCard := g.AddTextCard(300, 400)
	replaceCard.Text = "baz"

	// Create find/replace card
	frCard := g.AddFindReplaceCard(100, 300)

	// Wire up
	g.arrows = append(g.arrows, &Arrow{
		FromCardID: inputCard.ID,
		FromPort:   "text",
		ToCardID:   frCard.ID,
		ToPort:     "input",
		Color:      ColorArrowDefault,
	})
	g.arrows = append(g.arrows, &Arrow{
		FromCardID: findCard.ID,
		FromPort:   "text",
		ToCardID:   frCard.ID,
		ToPort:     "find",
		Color:      ColorArrowDefault,
	})
	g.arrows = append(g.arrows, &Arrow{
		FromCardID: replaceCard.ID,
		FromPort:   "text",
		ToCardID:   frCard.ID,
		ToPort:     "replace",
		Color:      ColorArrowDefault,
	})

	// First execution
	g.engine.Run()

	// Check result
	if frCard.Text != "foo baz" {
		t.Errorf("Expected 'foo baz', got '%s'", frCard.Text)
	}

	// Verify cache entry exists
	if _, ok := g.engine.ExecutionCache[frCard.ID]; !ok {
		t.Error("Expected cache entry for find/replace card")
	}

	// Run again with same inputs - should hit cache
	initialCacheTime := g.engine.ExecutionCache[frCard.ID].ExecutedAt

	g.engine.Run()

	// Cache time should be the same (not re-executed)
	if g.engine.ExecutionCache[frCard.ID].ExecutedAt != initialCacheTime {
		t.Error("Expected cache hit, but card was re-executed")
	}

	// Change input and run again
	inputCard.Text = "new text"
	g.engine.Run()

	// Result should be updated
	if frCard.Text != "new text" {
		t.Errorf("Expected 'new text', got '%s'", frCard.Text)
	}
}
