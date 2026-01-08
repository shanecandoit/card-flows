package main

import (
	"testing"
)

func TestRegisterSubscription(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{}
	g.arrows = []*Arrow{}

	// Create two cards
	card1 := g.AddTextCard(100, 100)
	card1.Text = "Source"
	card2 := g.AddTextCard(100, 300)
	card2.Text = "Target"

	// Register subscription
	g.RegisterSubscription(card1.ID, card2.ID, "text")

	// Verify subscription was added
	if len(card1.Subscribers) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", len(card1.Subscribers))
	}

	if card1.Subscribers[0].CardID != card2.ID {
		t.Errorf("Expected subscriber CardID %s, got %s", card2.ID, card1.Subscribers[0].CardID)
	}

	if card1.Subscribers[0].Port != "text" {
		t.Errorf("Expected subscriber Port 'text', got %s", card1.Subscribers[0].Port)
	}
}

func TestUnregisterSubscription(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{}
	g.arrows = []*Arrow{}

	// Create two cards
	card1 := g.AddTextCard(100, 100)
	card2 := g.AddTextCard(100, 300)

	// Register then unregister subscription
	g.RegisterSubscription(card1.ID, card2.ID, "text")
	g.UnregisterSubscription(card1.ID, card2.ID, "text")

	// Verify subscription was removed
	if len(card1.Subscribers) != 0 {
		t.Errorf("Expected 0 subscribers, got %d", len(card1.Subscribers))
	}
}

func TestPropagateText(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{}
	g.arrows = []*Arrow{}

	// Create two cards
	card1 := g.AddTextCard(100, 100)
	card1.Text = "Hello"
	card2 := g.AddTextCard(100, 300)
	card2.Text = ""

	// Register subscription
	g.RegisterSubscription(card1.ID, card2.ID, "text")

	// Propagate text
	g.PropagateText(card1)

	// Verify text was copied
	if card2.Text != "Hello" {
		t.Errorf("Expected card2 text 'Hello', got '%s'", card2.Text)
	}
}

func TestPropagateTextChain(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{}
	g.arrows = []*Arrow{}

	// Create three cards in a chain: A -> B -> C
	cardA := g.AddTextCard(100, 100)
	cardA.Text = "Source"
	cardB := g.AddTextCard(100, 300)
	cardB.Text = ""
	cardC := g.AddTextCard(100, 500)
	cardC.Text = ""

	// Register subscriptions: A -> B -> C
	g.RegisterSubscription(cardA.ID, cardB.ID, "text")
	g.RegisterSubscription(cardB.ID, cardC.ID, "text")

	// Propagate from A
	g.PropagateText(cardA)

	// Verify text propagated through the chain
	if cardB.Text != "Source" {
		t.Errorf("Expected cardB text 'Source', got '%s'", cardB.Text)
	}
	if cardC.Text != "Source" {
		t.Errorf("Expected cardC text 'Source', got '%s'", cardC.Text)
	}
}

func TestGetInputValue_Connected(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{}
	g.arrows = []*Arrow{}

	// Create two cards
	card1 := g.AddTextCard(100, 100)
	card1.Text = "Input Text"
	card2 := g.AddTextCard(100, 300)
	card2.Text = "Own Text"

	// Create arrow connection
	arrow := &Arrow{
		FromCardID: card1.ID,
		FromPort:   "text",
		ToCardID:   card2.ID,
		ToPort:     "text",
		Color:      ColorArrowDefault,
	}
	g.arrows = append(g.arrows, arrow)

	// Get input value for connected card
	value := g.GetInputValue(card2.ID, "text")

	// Should return source card's text
	if value != "Input Text" {
		t.Errorf("Expected 'Input Text', got '%s'", value)
	}
}

func TestGetInputValue_NotConnected(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{}
	g.arrows = []*Arrow{}

	// Create a card with no connections
	card := g.AddTextCard(100, 100)
	card.Text = "Own Text"

	// Get input value for unconnected card
	value := g.GetInputValue(card.ID, "text")

	// Should return card's own text
	if value != "Own Text" {
		t.Errorf("Expected 'Own Text', got '%s'", value)
	}
}

func TestArrowCreation_RegistersSubscription(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{}
	g.arrows = []*Arrow{}

	// Create two cards
	card1 := g.AddTextCard(100, 100)
	card1.Text = "Source"
	card2 := g.AddTextCard(100, 300)
	card2.Text = ""

	// Simulate arrow creation (what happens in input.go)
	arrow := &Arrow{
		FromCardID: card1.ID,
		FromPort:   "text",
		ToCardID:   card2.ID,
		ToPort:     "text",
		Color:      ColorArrowDefault,
	}
	g.arrows = append(g.arrows, arrow)
	g.RegisterSubscription(card1.ID, card2.ID, "text")
	g.PropagateText(card1)

	// Verify subscription was registered
	if len(card1.Subscribers) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", len(card1.Subscribers))
	}

	// Verify text was propagated
	if card2.Text != "Source" {
		t.Errorf("Expected card2 text 'Source', got '%s'", card2.Text)
	}
}

func TestMultipleSubscribers(t *testing.T) {
	g := NewGame()
	g.cards = []*Card{}
	g.arrows = []*Arrow{}

	// Create one source and two targets
	source := g.AddTextCard(100, 100)
	source.Text = "Broadcast"
	target1 := g.AddTextCard(100, 300)
	target1.Text = ""
	target2 := g.AddTextCard(300, 300)
	target2.Text = ""

	// Register both subscriptions
	g.RegisterSubscription(source.ID, target1.ID, "text")
	g.RegisterSubscription(source.ID, target2.ID, "text")

	// Propagate text
	g.PropagateText(source)

	// Verify both targets received the text
	if target1.Text != "Broadcast" {
		t.Errorf("Expected target1 text 'Broadcast', got '%s'", target1.Text)
	}
	if target2.Text != "Broadcast" {
		t.Errorf("Expected target2 text 'Broadcast', got '%s'", target2.Text)
	}
}
