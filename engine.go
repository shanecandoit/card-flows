package main

import (
	"fmt"
	"image/color"
	"sort"

	"go.starlark.net/starlark"
)

// Engine handles the execution of the flow
type Engine struct {
	game   *Game
	Memory map[string]interface{} // Cache outputs: Key = CardID + InputValuesHash
}

func NewEngine(g *Game) *Engine {
	return &Engine{
		game:   g,
		Memory: make(map[string]interface{}),
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

func (e *Engine) executeCard(c *Card) {
	// 1. Gather Inputs
	inputs := make(map[string]interface{})

	// Find arrows pointing to this card
	for _, arrow := range e.game.arrows {
		if arrow.ToCardID == c.ID {
			// Get value from Memory using FromCardID
			// We need to know which output port of the source card connects to which input port of this card
			// But simple memory model: Memory[CardID] = map[string]interface{} (outputs)
			// Wait, Memory definition was: key = CardID + InputHash.
			// Simplest first: Memory stores latest output for a CardID.
			// Memory[CardID] -> map[outputName]value

			// Let's refine Memory.
			// For now: Memory map[string]interface{} where Key is "CardID:PortName"
			key := fmt.Sprintf("%s:%s", arrow.FromCardID, arrow.FromPort)
			if val, ok := e.Memory[key]; ok {
				inputs[arrow.ToPort] = val
			}
		}
	}

	// 2. Prepare Starlark Thread
	thread := &starlark.Thread{Name: c.Title, Print: func(_ *starlark.Thread, msg string) { fmt.Println(msg) }}

	// 3. Define Predefined Functions/Globals
	globals := starlark.StringDict{}

	// Inject Inputs as variables
	for k, v := range inputs {
		val, err := toStarlarkValue(v)
		if err == nil {
			globals[k] = val
		}
	}

	// 4. Get Script based on Card Type
	script := e.getCardScript(c)

	// 5. Execute
	// We want to extract outputs.
	// In Starlark, we can just execute and check globals?
	// Or return a dict?
	// Let's assume the script defines variables that match output ports.
	globals, err := starlark.ExecFile(thread, c.Title, script, globals)
	if err != nil {
		fmt.Printf("Error executing card %s: %v\n", c.Title, err)
		c.Color = color.RGBA{255, 100, 100, 255} // Error Red
		return
	}

	// Validation Success Color (Green-ish)
	c.Color = color.RGBA{50, 205, 50, 255}

	// 6. Store Outputs
	for _, p := range c.Outputs {
		if val, ok := globals[p.Name]; ok {
			goVal := fromStarlarkValue(val)
			key := fmt.Sprintf("%s:%s", c.ID, p.Name)
			e.Memory[key] = goVal
			fmt.Printf("Stored Output: %s = %v\n", key, goVal)
		}
	}
}

func (e *Engine) getCardScript(c *Card) string {
	switch c.Title {
	case "Text Card":
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

	case "String:find_replace":
		return `
def run():
    if "input" not in locals() or "find" not in locals() or "replace" not in locals():
        return ""
    return input.replace(find, replace)

result = run()
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
