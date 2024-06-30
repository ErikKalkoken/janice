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

// JSONDocument represents a JSON document.
//
// The purpose of this type is to compose a formatted string tree from a nested data structure,
// so it can be rendered directly with a Fyne tree widget.
type JSONDocument struct {
	// Current progress while loading a new tree from data
	Progress binding.Float

	ids          map[widget.TreeNodeID][]widget.TreeNodeID
	values       map[widget.TreeNodeID]string
	ids2         map[widget.TreeNodeID][]widget.TreeNodeID
	values2      map[widget.TreeNodeID]string
	n            int
	sizeEstimate int
}

// Returns a new JSONDocument object.
func NewJSONDocument() (JSONDocument, error) {
	t := JSONDocument{Progress: binding.NewFloat()}
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
func (t *JSONDocument) Value(uid widget.TreeNodeID) string {
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
	t.ids = nil
	t.values = nil
	return nil
}

// Size returns the number of nodes.
func (t *JSONDocument) Size() int {
	return t.n
}

// Reset resets the tree.
func (t *JSONDocument) Reset() error {
	t.values = make(map[widget.TreeNodeID]string)
	t.ids = make(map[widget.TreeNodeID][]widget.TreeNodeID)
	t.values2 = make(map[widget.TreeNodeID]string)
	t.ids2 = make(map[widget.TreeNodeID][]widget.TreeNodeID)
	t.n = 0
	t.sizeEstimate = 1
	return t.Progress.Set(0)
}

func (t *JSONDocument) addObjectWithList(parentUID widget.TreeNodeID, data map[string][]any) {
	var uid string
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		uid = t.add(parentUID, k)
		t.addSlice(uid, data[k])
	}
}

func (t *JSONDocument) addObject(parentUID widget.TreeNodeID, data map[string]any) {
	var uid string
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		v := data[k]
		switch v2 := v.(type) {
		case []any:
			uid = t.add(parentUID, k)
			t.addSlice(uid, v2)
		case string:
			t.add(uid, fmt.Sprintf("%s: \"%s\"", k, v2))
		default:
			t.add(uid, fmt.Sprintf("%s: %s", k, formatValue(v2)))
		}
	}
}

func (t *JSONDocument) addSlice(parentUID string, a []any) {
	var uid string
	for i, v := range a {
		switch v2 := v.(type) {
		case []any:
			uid = t.add(parentUID, formatArrayIndex(i))
			t.addSlice(uid, v2)
		case map[string][]any:
			uid = t.add(parentUID, formatArrayIndex(i))
			t.addObjectWithList(uid, v2)
		case map[string]any:
			uid = t.add(parentUID, formatArrayIndex(i))
			t.addObject(uid, v2)
		case string:
			t.add(parentUID, fmt.Sprintf("%s: \"%s\"", formatArrayIndex(i), v2))
		default:
			t.add(parentUID, fmt.Sprintf("%s: %s", formatArrayIndex(i), formatValue(v2)))
		}
	}
}

// add adds a node to the tree and returns the UID.
// Nodes will be rendered in the same order they are added.
// Use "" as parentUID for adding nodes at the top level.
// Returns the generated UID for this node and the incremented ID
func (t *JSONDocument) add(parentUID widget.TreeNodeID, value string) widget.TreeNodeID {
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
	uid := fmt.Sprintf("%s-%d-%s", s, t.n, value)
	_, found := t.values[uid]
	if found {
		panic(fmt.Sprintf("UID for this node already exists: %v", uid))
	}
	t.ids[parentUID] = append(t.ids[parentUID], uid)
	t.values[uid] = value
	t.n++
	if t.n%progressUpdateTick == 0 {
		t.Progress.Set(min(1, float64(t.n)/float64(t.sizeEstimate)))
	}
	return uid
}

func formatArrayIndex(v int) string {
	return fmt.Sprintf("[%d]", v)
}

func formatValue(v any) string {
	if v == nil {
		return "null"
	}
	return fmt.Sprint(v)
}
