package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"card-flows/engine"
	"card-flows/graph"
)

// CacheEntry stores execution results with metadata
type CacheEntry struct {
	InputHash  string
	Output     interface{}
	ExecutedAt time.Time
}

// Engine handles the execution of the flow
type Engine struct {
	game           *Game
	Memory         map[string]interface{} // Cache outputs: Key = CardID + InputValuesHash
	ExecutionCache map[string]CacheEntry  // Cache with metadata
}

func NewEngine(g *Game) *Engine {
	return &Engine{
		game:           g,
		Memory:         make(map[string]interface{}),
		ExecutionCache: make(map[string]CacheEntry),
	}
}

// Run executes the flow starting from the roots or all nodes
func (e *Engine) Run() {
	// 1. Build Dependency Graph & Sort
	order, err := e.getExecutionOrder()
	if err != nil {
		fmt.Println("Execution Error:", err)
		return
	}

	// 2. Execute in Order
	for _, card := range order {
		e.executeCard(card)
	}
}

func (e *Engine) getExecutionOrder() ([]*Card, error) {
	// Build lightweight node/arrow lists for the graph package
	nodes := []graph.Node{}
	for _, c := range e.game.cards {
		nodes = append(nodes, graph.Node{ID: c.ID, X: c.X, Y: c.Y})
	}
	arrows := []graph.Arrow{}
	for _, a := range e.game.arrows {
		arrows = append(arrows, graph.Arrow{FromID: a.FromCardID, ToID: a.ToCardID})
	}

	orderedIDs, err := graph.TopologicalSort(nodes, arrows)
	if err != nil {
		return nil, err
	}

	// Map ordered IDs back to card pointers
	idToCard := make(map[string]*Card)
	for _, c := range e.game.cards {
		idToCard[c.ID] = c
	}
	result := []*Card{}
	for _, id := range orderedIDs {
		if c, ok := idToCard[id]; ok {
			result = append(result, c)
		}
	}

	// Deterministic fallback: ensure all cards included (shouldn't be necessary)
	if len(result) != len(e.game.cards) {
		missing := []*Card{}
		present := make(map[string]bool)
		for _, c := range result {
			present[c.ID] = true
		}
		for _, c := range e.game.cards {
			if !present[c.ID] {
				missing = append(missing, c)
			}
		}
		sort.Slice(missing, func(i, j int) bool {
			if missing[i].Y != missing[j].Y {
				return missing[i].Y < missing[j].Y
			}
			return missing[i].X < missing[j].X
		})
		result = append(result, missing...)
	}
	return result, nil
}

// Note: input hashing moved to the `engine` package (engine.ComputeInputHash).

func (e *Engine) executeCard(c *Card) {
	// 1. Gather Inputs
	inputs := make(map[string]interface{})

	// Find arrows pointing to this card
	for _, arrow := range e.game.arrows {
		if arrow.ToCardID == c.ID {
			// Get value from source card's output
			key := fmt.Sprintf("%s:%s", arrow.FromCardID, arrow.FromPort)
			if val, ok := e.Memory[key]; ok {
				inputs[arrow.ToPort] = val
			} else {
				// Try to get from source card directly (for text cards)
				sourceCard := e.game.getCardByID(arrow.FromCardID)
				if sourceCard != nil && strings.HasPrefix(sourceCard.Title, "Text Card") {
					inputs[arrow.ToPort] = sourceCard.Text
				}
			}
		}
	}

	// 2. Check Cache
	// For text cards, include their text in the cache key since they're "input" nodes
	cacheInputs := inputs
	if strings.HasPrefix(c.Title, "Text Card") {
		cacheInputs = map[string]interface{}{"_text": c.Text}
	}
	inputHash := engine.ComputeInputHash(c.ID, cacheInputs)
	if cached, ok := e.ExecutionCache[c.ID]; ok && cached.InputHash == inputHash {
		fmt.Printf("[%s] Cache hit - using cached result\n", c.Title)
		// Use cached result
		for _, p := range c.Outputs {
			key := fmt.Sprintf("%s:%s", c.ID, p.Name)
			e.Memory[key] = cached.Output
		}
		// Update card text with cached result for display
		if c.Title == "String:find_replace" {
			if str, ok := cached.Output.(string); ok {
				c.Text = str
				// Propagate to subscribers
				e.game.PropagateText(c)
			}
		}
		return
	}

	// 3. Log execution start
	fmt.Printf("[%s] Executing with inputs: %v\n", c.Title, inputs)

	// 4. Execute based on card type
	var result interface{}
	var err error

	if c.Type == "text" {
		// Text cards just output their text
		result = c.Text
	} else {
		// Execute Starlark for functional cards
		result, err = e.executeStarlark(c, inputs)
		if err != nil {
			fmt.Printf("[%s] Execution error: %v\n", c.Title, err)
			c.LastErrorFlash = time.Now()
			return
		}
	}

	// 5. Log execution result
	fmt.Printf("[%s] Result: %v\n", c.Title, result)

	// 6. Success flash (200ms)
	c.LastSuccessFlash = time.Now()

	// 7. Store in cache
	e.ExecutionCache[c.ID] = CacheEntry{
		InputHash:  inputHash,
		Output:     result,
		ExecutedAt: time.Now(),
	}

	// 8. Store Outputs in Memory
	for _, p := range c.Outputs {
		key := fmt.Sprintf("%s:%s", c.ID, p.Name)
		e.Memory[key] = result
	}

	// 9. Update card text with result for display and propagate
	if c.Title == "String:find_replace" {
		if str, ok := result.(string); ok {
			c.Text = str
			// Propagate to subscribers using pub-sub
			e.game.PropagateText(c)
		}
	}
}

// executeStarlark runs Starlark code for a card via enginepkg helper
func (e *Engine) executeStarlark(c *Card, inputs map[string]interface{}) (interface{}, error) {
	// Ensure defaults for find_replace
	if c.Type == "find_replace" {
		if _, ok := inputs["input"]; !ok {
			inputs["input"] = ""
		}
		if _, ok := inputs["find"]; !ok {
			inputs["find"] = ""
		}
		if _, ok := inputs["replace"]; !ok {
			inputs["replace"] = ""
		}
	}

	script := e.getCardScript(c)
	if script == "" {
		return nil, fmt.Errorf("no script defined for card type: %s", c.Title)
	}

	outputs, err := engine.ExecuteStarlark(c.Title, script, inputs)
	if err != nil {
		return nil, err
	}

	for _, p := range c.Outputs {
		if v, ok := outputs[p.Name]; ok {
			return v, nil
		}
	}
	return nil, fmt.Errorf("no output found")
}

func (e *Engine) getCardScript(c *Card) string {
	if c.Type == "text" {
		// Pass through or literal text
		// If 'text' input is connected, it passes through.
		// Else it uses c.Text
		return `
if "text" not in locals():
    text = "` + c.Text + `"
` // If input 'text' exists (in globals), this script does nothing, preserving it.
	}

	if c.Type == "find_replace" {
		return `
# Perform find and replace operation
result = input.replace(find, replace) if input and find else input
`
	}
	return ""
}

// Note: Starlark conversion helpers moved to enginepkg package.
