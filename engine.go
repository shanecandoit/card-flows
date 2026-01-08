package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image/color"
	"sort"
	"strings"
	"time"

	"go.starlark.net/starlark"
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
	// Kahn's Algorithm for Topological Sort
	// OR simple DFS based sort since we want a sequence.

	// Build Graph
	graph := make(map[string][]string) // FromID -> []ToIDs
	inDegree := make(map[string]int)

	// Initialize inDegree for all cards
	for _, c := range e.game.cards {
		inDegree[c.ID] = 0
	}

	for _, arrow := range e.game.arrows {
		graph[arrow.FromCardID] = append(graph[arrow.FromCardID], arrow.ToCardID)
		inDegree[arrow.ToCardID]++
	}

	// Queue for source nodes
	queue := []*Card{}
	for _, c := range e.game.cards {
		if inDegree[c.ID] == 0 {
			queue = append(queue, c)
		}
	}

	// Sort (optional: sort by position for deterministic order?)
	// Stable sort by Y then X
	sort.Slice(queue, func(i, j int) bool {
		if queue[i].Y != queue[j].Y {
			return queue[i].Y < queue[j].Y
		}
		return queue[i].X < queue[j].X
	})

	result := []*Card{}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		result = append(result, u)

		for _, vID := range graph[u.ID] {
			inDegree[vID]--
			if inDegree[vID] == 0 {
				v := e.game.getCardByID(vID)
				if v != nil {
					queue = append(queue, v)
				}
			}
		}
	}

	if len(result) != len(e.game.cards) {
		return nil, fmt.Errorf("cycle detected in flow")
	}

	return result, nil
}

// computeInputHash creates a hash of card inputs for cache key
func (e *Engine) computeInputHash(cardID string, inputs map[string]interface{}) string {
	// Serialize inputs to JSON and hash
	data := map[string]interface{}{
		"cardID": cardID,
		"inputs": inputs,
	}
	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash)
}

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
	inputHash := e.computeInputHash(c.ID, cacheInputs)
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

	if strings.HasPrefix(c.Title, "Text Card") {
		// Text cards just output their text
		result = c.Text
	} else {
		// Execute Starlark for functional cards
		result, err = e.executeStarlark(c, inputs)
		if err != nil {
			fmt.Printf("[%s] Execution error: %v\n", c.Title, err)
			c.Color = color.RGBA{255, 100, 100, 255} // Error Red
			return
		}
	}

	// 5. Log execution result
	fmt.Printf("[%s] Result: %v\n", c.Title, result)

	// 6. Success color
	c.Color = color.RGBA{50, 205, 50, 255}

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

// executeStarlark runs Starlark code for a card
func (e *Engine) executeStarlark(c *Card, inputs map[string]interface{}) (interface{}, error) {
	// 1. Prepare Starlark Thread
	thread := &starlark.Thread{Name: c.Title, Print: func(_ *starlark.Thread, msg string) { fmt.Println(msg) }}

	// 2. Define Predefined Functions/Globals
	globals := starlark.StringDict{}

	// For find_replace cards, ensure all inputs have defaults
	if c.Title == "String:find_replace" {
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

	// Inject Inputs as variables
	for k, v := range inputs {
		val, err := toStarlarkValue(v)
		if err == nil {
			globals[k] = val
		}
	}

	// 3. Get Script based on Card Type
	script := e.getCardScript(c)
	if script == "" {
		return nil, fmt.Errorf("no script defined for card type: %s", c.Title)
	}

	// 4. Execute
	resultGlobals, err := starlark.ExecFile(thread, c.Title, script, globals)
	if err != nil {
		return nil, err
	}

	// 5. Extract output
	for _, p := range c.Outputs {
		if val, ok := resultGlobals[p.Name]; ok {
			return fromStarlarkValue(val), nil
		}
	}

	return nil, fmt.Errorf("no output found")
}

func (e *Engine) getCardScript(c *Card) string {
	// Handle text cards (with or without ID suffix)
	if strings.HasPrefix(c.Title, "Text Card") {
		// Pass through or literal text
		// If 'text' input is connected, it passes through.
		// Else it uses c.Text
		// For now, let's say it outputs 'text' port.
		// If we have an input named 'text', pass it through.
		return `
if "text" not in locals():
    text = "` + c.Text + `"
` // If input 'text' exists (in globals), this script does nothing, preserving it.
		// Actually globals are pre-populated.
		// If 'text' is in globals, we are good.
		// If not, we define it from c.Text.
	}

	if c.Title == "String:find_replace" {
		return `
# Perform find and replace operation
result = input.replace(find, replace) if input and find else input
`
	}
	return ""
}

// Helpers for type conversion
func toStarlarkValue(v interface{}) (starlark.Value, error) {
	if v == nil {
		return starlark.None, nil
	}
	switch val := v.(type) {
	case string:
		return starlark.String(val), nil
	case int:
		return starlark.MakeInt(val), nil
	case float64:
		return starlark.Float(val), nil
	case bool:
		return starlark.Bool(val), nil
	}
	return starlark.None, fmt.Errorf("unsupported type: %T", v)
}

func fromStarlarkValue(v starlark.Value) interface{} {
	switch val := v.(type) {
	case starlark.String:
		return string(val)
	case starlark.Int:
		i, _ := val.Int64()
		return int(i)
	case starlark.Float:
		return float64(val)
	case starlark.Bool:
		return bool(val)
	}
	return nil
}
