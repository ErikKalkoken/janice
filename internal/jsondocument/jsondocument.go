// Package jsondocument contains the logic for processing large JSON documents.
package jsondocument

import (
	"fmt"
	"slices"

	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

const (
	// Update progress after x added nodes
	progressUpdateTick = 10_000
)

type Node struct {
	Key   string
	Value any
}

var Empty = struct{}{}

// JSONDocument represents a JSON document.
//
// The purpose of this type is to compose a formatted string tree from a nested data structure,
// so it can be rendered directly with a Fyne tree widget.
type JSONDocument struct {
	// Current progress while loading a new tree from data
	Progress binding.Float

	ids          map[widget.TreeNodeID][]widget.TreeNodeID
	values       map[widget.TreeNodeID]Node
	ids2         map[widget.TreeNodeID][]widget.TreeNodeID
	values2      map[widget.TreeNodeID]Node
	n            int
	sizeEstimate int
}

// Returns a new JSONDocument object.
func NewJSONDocument() (*JSONDocument, error) {
	t := &JSONDocument{Progress: binding.NewFloat()}
	if err := t.Reset(); err != nil {
		return t, err
	}
	return t, nil
}

// ChildUIDs returns the child UIDs for a given node.
// This can be used directly in the tree widget childUIDs() function.
func (t *JSONDocument) ChildUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	return t.ids2[uid]
}

// IsBranch reports wether a node is a branch.
// This can be used directly in the tree widget isBranch() function.
func (t *JSONDocument) IsBranch(uid widget.TreeNodeID) bool {
	_, found := t.ids2[uid]
	return found
}

// Value returns the value of a node
func (t *JSONDocument) Value(uid widget.TreeNodeID) Node {
	return t.values2[uid]
}

// Load loads a new tree from data.
func (t *JSONDocument) Load(data any, sizeEstimate int) error {
	if err := t.Reset(); err != nil {
		return err
	}
	t.sizeEstimate = sizeEstimate
	switch v := data.(type) {
	case map[string]any:
		t.addObject("", v)
	case []any:
		t.addSlice("", v)
	default:
		return fmt.Errorf("unrecognized format")
	}
	t.ids2 = t.ids
	t.values2 = t.values
	t.ids = make(map[widget.TreeNodeID][]widget.TreeNodeID)
	t.values = make(map[widget.TreeNodeID]Node)
	return nil
}

// Size returns the number of nodes.
func (t *JSONDocument) Size() int {
	return t.n
}

// Reset resets the tree.
func (t *JSONDocument) Reset() error {
	t.values = make(map[widget.TreeNodeID]Node)
	t.ids = make(map[widget.TreeNodeID][]widget.TreeNodeID)
	t.values2 = make(map[widget.TreeNodeID]Node)
	t.ids2 = make(map[widget.TreeNodeID][]widget.TreeNodeID)
	t.n = 0
	t.sizeEstimate = 1
	return t.Progress.Set(0)
}

func (t *JSONDocument) addObject(parentUID widget.TreeNodeID, data map[string]any) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		v := data[k]
		t.addValue(parentUID, k, v)
	}
}

func (t *JSONDocument) addSlice(parentUID string, a []any) {
	for i, v := range a {
		k := fmt.Sprintf("[%d]", i)
		t.addValue(parentUID, k, v)
	}
}

func (t *JSONDocument) addValue(parentUID widget.TreeNodeID, k string, v any) {
	switch v2 := v.(type) {
	case map[string]any:
		uid := t.addNode(parentUID, k, Empty)
		t.addObject(uid, v2)
	case []any:
		uid := t.addNode(parentUID, k, Empty)
		t.addSlice(uid, v2)
	default:
		t.addNode(parentUID, k, v2)
	}
}

// addNode adds a node to the tree and returns the UID.
// Nodes will be rendered in the same order they are added.
// Use "" as parentUID for adding nodes at the top level.
// Returns the generated UID for this node and the incremented ID
func (t *JSONDocument) addNode(parentUID widget.TreeNodeID, key string, value any) widget.TreeNodeID {
	if parentUID != "" {
		_, found := t.values[parentUID]
		if !found {
			panic(fmt.Sprintf("parent UID does not exist: %s", parentUID))
		}
	}
	s := parentUID
	if parentUID == "" {
		s = "ID"
	}
	uid := fmt.Sprintf("%s-%d", s, t.n)
	_, found := t.values[uid]
	if found {
		panic(fmt.Sprintf("UID for this node already exists: %v", uid))
	}
	t.ids[parentUID] = append(t.ids[parentUID], uid)
	t.values[uid] = Node{Key: key, Value: value}
	t.n++
	if t.n%progressUpdateTick == 0 {
		t.Progress.Set(min(1, float64(t.n)/float64(t.sizeEstimate)))
	}
	return uid
}
