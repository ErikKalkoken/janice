// Package jsondocument contains the logic for processing large JSON documents.
package jsondocument

import (
	"fmt"
	"slices"
	"sync"

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
	infoText binding.Int
	ids      map[widget.TreeNodeID][]widget.TreeNodeID
	values   map[widget.TreeNodeID]Node
	n        int
	mu       sync.RWMutex
}

// Returns a new JSONDocument object.
func NewJSONDocument() *JSONDocument {
	t := &JSONDocument{}
	t.reset()
	return t
}

// ChildUIDs returns the child UIDs for a given node.
// This can be used directly in the tree widget childUIDs() function.
func (t *JSONDocument) ChildUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	if !t.mu.TryRLock() {
		return []widget.TreeNodeID{}
	}
	defer t.mu.RUnlock()
	return t.ids[uid]
}

// IsBranch reports wether a node is a branch.
// This can be used directly in the tree widget isBranch() function.
func (t *JSONDocument) IsBranch(uid widget.TreeNodeID) bool {
	if !t.mu.TryRLock() {
		return false
	}
	defer t.mu.RUnlock()
	_, found := t.ids[uid]
	return found
}

// Value returns the value of a node
func (t *JSONDocument) Value(uid widget.TreeNodeID) Node {
	if !t.mu.TryRLock() {
		return Node{}
	}
	defer t.mu.RUnlock()
	return t.values[uid]
}

// Load loads a new tree from data.
func (t *JSONDocument) Load(data any, infoText binding.Int) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.reset()
	t.infoText = infoText
	switch v := data.(type) {
	case map[string]any:
		t.addObject("", v)
	case []any:
		t.addSlice("", v)
	default:
		return fmt.Errorf("unrecognized format")
	}
	return nil
}

// Size returns the number of nodes.
func (t *JSONDocument) Size() int {
	if !t.mu.TryRLock() {
		return 0
	}
	defer t.mu.RUnlock()
	return t.n
}

func (t *JSONDocument) reset() {
	t.values = make(map[widget.TreeNodeID]Node)
	t.ids = make(map[widget.TreeNodeID][]widget.TreeNodeID)
	t.n = 0
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
		t.infoText.Set(t.n)
	}
	return uid
}
