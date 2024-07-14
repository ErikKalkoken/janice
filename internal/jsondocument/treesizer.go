package jsondocument

import (
	"fmt"
)

// JSONTreeSizer is an object for calculating the number of nodes in a tree structure.
// It is specialized for data produced by un-marshalling a JSON document.
type JSONTreeSizer struct {
	count int
}

// Calculate returns the number of nodes in a tree structure.
func (t *JSONTreeSizer) Calculate(data any) (int, error) {
	t.count = 0
	switch v := data.(type) {
	case map[string]any:
		t.count++
		t.parseObject(v)
	case []any:
		t.count++
		t.parseArray(v)
	default:
		return 0, fmt.Errorf("unrecognized format")
	}
	return t.count, nil
}

// parseObject parses an object in a JSON tree.
func (t *JSONTreeSizer) parseObject(data map[string]any) {
	for _, v := range data {
		t.parseValue(v)
	}
}

// parseArray parses an array in a JSON tree.
func (t *JSONTreeSizer) parseArray(a []any) {
	for _, v := range a {
		t.parseValue(v)
	}
}

// parseArray parses a value in a JSON tree
func (t *JSONTreeSizer) parseValue(v any) {
	t.count++
	switch v2 := v.(type) {
	case map[string]any:
		t.parseObject(v2)
	case []any:
		t.parseArray(v2)
	}
}
