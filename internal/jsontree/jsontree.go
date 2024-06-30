package jsontree

import (
	"fmt"
	"slices"

	"fyne.io/fyne/v2/widget"
)

type JSONTree struct {
	ids    map[widget.TreeNodeID][]widget.TreeNodeID
	values map[widget.TreeNodeID]string
	id     int
}

// Returns a new jsonTree object.
func NewJSONTree() JSONTree {
	t := JSONTree{}
	t.reset()
	return t
}

// ChildUIDs returns the child UIDs for a given node.
// This can be used directly in the tree widget childUIDs() function.
func (t *JSONTree) ChildUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	return t.ids[uid]
}

// IsBranch reports wether a node is a branch.
// This can be used directly in the tree widget isBranch() function.
func (t *JSONTree) IsBranch(uid widget.TreeNodeID) bool {
	_, found := t.ids[uid]
	return found
}

// Value returns the value of a node
func (t *JSONTree) Value(uid widget.TreeNodeID) string {
	return t.values[uid]
}

// Set replaces the complete tree with the given data and returns the number of nodes.
func (t *JSONTree) Set(data any) (int, error) {
	t.reset()
	switch v := data.(type) {
	case map[string]any:
		t.addObject("", v)
	case []any:
		t.addSlice("", v)
	default:
		return 0, fmt.Errorf("unrecognized format")
	}
	return t.id, nil
}

func (t *JSONTree) reset() {
	t.values = make(map[widget.TreeNodeID]string)
	t.ids = make(map[widget.TreeNodeID][]widget.TreeNodeID)
	t.id = 0
}

func (t *JSONTree) addObjectWithList(parentUID widget.TreeNodeID, data map[string][]any) {
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

func (t *JSONTree) addObject(parentUID widget.TreeNodeID, data map[string]any) {
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

func (t *JSONTree) addSlice(parentUID string, a []any) {
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
func (t *JSONTree) add(parentUID widget.TreeNodeID, value string) widget.TreeNodeID {
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
	uid := fmt.Sprintf("%s-%d-%s", s, t.id, value)
	_, found := t.values[uid]
	if found {
		panic(fmt.Sprintf("UID for this node already exists: %v", uid))
	}
	t.ids[parentUID] = append(t.ids[parentUID], uid)
	t.values[uid] = value
	t.id++
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
