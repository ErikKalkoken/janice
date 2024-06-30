package main

import (
	"fmt"
	"slices"

	"fyne.io/fyne/v2/widget"
)

type JSONTree struct {
	ids    map[widget.TreeNodeID][]widget.TreeNodeID
	values map[widget.TreeNodeID]string
}

// Returns a new jsonTree object.
func NewJSONTree() *JSONTree {
	t := &JSONTree{
		values: make(map[widget.TreeNodeID]string),
		ids:    make(map[widget.TreeNodeID][]widget.TreeNodeID),
	}
	return t
}

func (t *JSONTree) childUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	return t.ids[uid]
}

func (t *JSONTree) isBranch(uid widget.TreeNodeID) bool {
	_, found := t.ids[uid]
	return found
}

func (t *JSONTree) value(uid widget.TreeNodeID) string {
	return t.values[uid]
}

func (t *JSONTree) addObjectWithList(parentUID widget.TreeNodeID, data map[string][]any, id int) int {
	var uid string
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		uid, id = t.add(parentUID, id, k)
		id = t.addSlice(uid, data[k], id)
	}
	return id
}

func (t *JSONTree) addObject(parentUID widget.TreeNodeID, data map[string]any, id int) int {
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
			uid, id = t.add(parentUID, id, k)
			id = t.addSlice(uid, v2, id)
		case string:
			_, id = t.add(uid, id, fmt.Sprintf("%s: \"%s\"", k, v2))
		default:
			_, id = t.add(uid, id, fmt.Sprintf("%s: %s", k, formatValue(v2)))
		}
	}
	return id
}

func (t *JSONTree) addSlice(parentUID string, a []any, id int) int {
	var uid string
	for i, v := range a {
		switch v2 := v.(type) {
		case []any:
			uid, id = t.add(parentUID, id, formatArrayIndex(i))
			id = t.addSlice(uid, v2, id)
		case map[string][]any:
			uid, id = t.add(parentUID, id, formatArrayIndex(i))
			id = t.addObjectWithList(uid, v2, id)
		case map[string]any:
			uid, id = t.add(parentUID, id, formatArrayIndex(i))
			id = t.addObject(uid, v2, id)
		case string:
			_, id = t.add(parentUID, id, fmt.Sprintf("%s: \"%s\"", formatArrayIndex(i), v2))
		default:
			_, id = t.add(parentUID, id, fmt.Sprintf("%s: %s", formatArrayIndex(i), formatValue(v2)))
		}
	}
	return id
}

// add adds a node to the tree and returns the UID.
// Nodes will be rendered in the same order they are added.
// Use "" as parentUID for adding nodes at the top level.
// Returns the generated UID for this node and the incremented ID
func (t *JSONTree) add(parentUID widget.TreeNodeID, id int, value string) (widget.TreeNodeID, int) {
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
	uid := fmt.Sprintf("%s-%d-%s", s, id, value)
	_, found := t.values[uid]
	if found {
		panic(fmt.Sprintf("UID for this node already exists: %v", uid))
	}
	t.ids[parentUID] = append(t.ids[parentUID], uid)
	t.values[uid] = value
	id++
	return uid, id
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
