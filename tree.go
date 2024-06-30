package main

import (
	"fmt"

	"fyne.io/fyne/v2/widget"
)

type treeWrap struct {
	ids    map[widget.TreeNodeID][]widget.TreeNodeID
	values map[widget.TreeNodeID]string
}

// Returns a new DataNodeTree object.
func newTreeWrap() *treeWrap {
	t := &treeWrap{
		values: make(map[widget.TreeNodeID]string),
		ids:    make(map[widget.TreeNodeID][]widget.TreeNodeID),
	}
	return t
}

// add adds a node to the tree and returns the UID.
// Nodes will be rendered in the same order they are added.
// Use "" as parentUID for adding nodes at the top level.
// Returns the generated UID for this node and the incremented ID
func (t treeWrap) add(parentUID widget.TreeNodeID, id int, value string) (widget.TreeNodeID, int) {
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

func (t treeWrap) childUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	return t.ids[uid]
}

func (t treeWrap) isBranch(uid widget.TreeNodeID) bool {
	_, found := t.ids[uid]
	return found
}

func (t treeWrap) value(uid widget.TreeNodeID) string {
	return t.values[uid]
}
