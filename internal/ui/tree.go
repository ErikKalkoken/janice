package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/ErikKalkoken/janice/internal/jsondocument"
)

// jsonTree shows a JSON document in a tree structure.
type jsonTree struct {
	widget.Tree
	u *UI
}

func newJSONTree(u *UI) *jsonTree {
	w := &jsonTree{u: u}
	w.ExtendBaseWidget(w)

	w.ChildUIDs = func(id widget.TreeNodeID) []widget.TreeNodeID {
		return u.document.ChildUIDs(id)
	}
	w.IsBranch = func(id widget.TreeNodeID) bool {
		return u.document.IsBranch(id)
	}
	w.CreateNode = func(branch bool) fyne.CanvasObject {
		return newTreeNode()
	}
	w.UpdateNode = func(uid widget.TreeNodeID, branch bool, co fyne.CanvasObject) {
		node := u.document.Value(uid)
		obj := co.(*treeNode)
		var text string
		switch v := node.Value; node.Type {
		case jsondocument.Array:
			if branch {
				if t := u.tree; t != nil && t.IsBranchOpen(uid) {
					text = ""
				} else {
					text = "[...]"
				}
			} else {
				text = "[]"
			}
		case jsondocument.Object:
			if branch {
				if t := u.tree; t != nil && t.IsBranchOpen(uid) {
					text = ""
				} else {
					text = "{...}"
				}
			} else {
				text = "{}"
			}
		case jsondocument.String:
			text = fmt.Sprintf("\"%s\"", v)
		case jsondocument.Number:
			x := v.(float64)
			text = strconv.FormatFloat(x, 'f', -1, 64)
		case jsondocument.Boolean:
			text = fmt.Sprintf("%v", v)
		case jsondocument.Null:
			text = "null"
		default:
			text = fmt.Sprintf("%v", v)
		}
		obj.set(node.Key, text, type2importance[node.Type])
	}
	w.OnSelected = func(uid widget.TreeNodeID) {
		u.selectElement(uid)
	}
	return w
}

func (w *jsonTree) scrollTo(uid widget.TreeNodeID) {
	if uid == "" {
		return
	}
	p := w.u.document.Path(uid)
	for _, uid2 := range p {
		w.OpenBranch(uid2)
	}
	w.ScrollTo(uid)
	w.Select(uid)
}
