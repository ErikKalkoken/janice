package jsondocument

import (
	"fmt"
)

// JSONTreeSizer allows the quick sizing of a tree structure.
// It is specialized for data coming from unmarshaled JSON.
type JSONTreeSizer struct {
	count int
}

// Load loads a new tree from a reader.
func (t *JSONTreeSizer) Run(data any) (int, error) {
	t.count = 0
	switch v := data.(type) {
	case map[string]any:
		t.addObject(v)
	case []any:
		t.addArray(v)
	default:
		return 0, fmt.Errorf("unrecognized format")
	}
	return t.count, nil
}

// addObject adds a JSON object to the tree.
func (t *JSONTreeSizer) addObject(data map[string]any) {
	for _, v := range data {
		t.addValue(v)
	}
}

// addArray adds a JSON array to the tree.
func (t *JSONTreeSizer) addArray(a []any) {
	for _, v := range a {
		t.addValue(v)
	}
}

// addValue adds a JSON value to the tree.
func (t *JSONTreeSizer) addValue(v any) {
	t.count++
	switch v2 := v.(type) {
	case map[string]any:
		t.addObject(v2)
	case []any:
		t.addArray(v2)
	}
}
