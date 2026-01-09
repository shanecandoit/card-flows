package engine

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"go.starlark.net/starlark"
)

// computeInputHash creates a hash of card inputs for cache key
func ComputeInputHash(cardID string, inputs map[string]interface{}) string {
	data := map[string]interface{}{
		"cardID": cardID,
		"inputs": inputs,
	}
	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash)
}

// ExecuteStarlark executes a script with provided inputs and returns a map of output names to native Go values.
func ExecuteStarlark(threadName string, script string, inputs map[string]interface{}) (map[string]interface{}, error) {
	thread := &starlark.Thread{Name: threadName, Print: func(_ *starlark.Thread, msg string) { fmt.Println(msg) }}

	globals := starlark.StringDict{}
	for k, v := range inputs {
		if val, err := toStarlarkValue(v); err == nil {
			globals[k] = val
		}
	}

	resultGlobals, err := starlark.ExecFile(thread, threadName, script, globals)
	if err != nil {
		return nil, err
	}

	out := make(map[string]interface{})
	for k, v := range resultGlobals {
		out[k] = FromStarlarkValue(v)
	}
	return out, nil
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

func FromStarlarkValue(v starlark.Value) interface{} {
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
