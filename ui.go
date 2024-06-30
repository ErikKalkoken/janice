package main

import (
	"fmt"
	"log"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ui struct {
	app      fyne.App
	tree     *widget.Tree
	treeWrap *treeWrap
	window   fyne.Window
}

func newUI() *ui {
	u := &ui{
		app:      app.New(),
		treeWrap: newTreeWrap(),
	}
	u.window = u.app.NewWindow("JSON Viewer")
	u.tree = makeTree(u)
	b1 := widget.NewButton("Close All", func() {
		u.tree.CloseAllBranches()
	})
	c := container.NewBorder(container.NewHBox(b1), nil, nil, nil, u.tree)
	u.window.SetContent(c)
	u.window.Resize(fyne.Size{Width: 800, Height: 600})
	return u
}

func makeTree(u *ui) *widget.Tree {
	tree := widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			return u.treeWrap.childUIDs(id)
		},
		func(id widget.TreeNodeID) bool {
			return u.treeWrap.isBranch(id)
		},
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Leaf template")
		},
		func(uid widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			text := u.treeWrap.value(uid)
			o.(*widget.Label).SetText(text)
		})

	tree.OnSelected = func(uid widget.TreeNodeID) {
		u.tree.UnselectAll()
	}
	return tree
}

func (u *ui) showAndRun() {
	u.window.ShowAndRun()
}

func (u *ui) setData(data any) error {
	var id int
	switch v := data.(type) {
	case map[string]any:
		id = addObject("", u.treeWrap, v, 0)
	case []any:
		id = addSlice("", u.treeWrap, v, 0)
	default:
		return fmt.Errorf("unrecognized format")
	}
	log.Printf("Loaded JSON file into tree with %d nodes", id)
	u.tree.Refresh()
	return nil
}

func addObjectWithList(parentUID widget.TreeNodeID, wrap *treeWrap, data map[string][]any, id int) int {
	var uid string
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		uid, id = wrap.add(parentUID, id, k)
		id = addSlice(uid, wrap, data[k], id)
	}
	return id
}

func addObject(parentUID widget.TreeNodeID, wrap *treeWrap, data map[string]any, id int) int {
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
			uid, id = wrap.add(parentUID, id, k)
			id = addSlice(uid, wrap, v2, id)
		case string:
			_, id = wrap.add(uid, id, fmt.Sprintf("%s: \"%s\"", k, v2))
		default:
			_, id = wrap.add(uid, id, fmt.Sprintf("%s: %s", k, formatValue(v2)))
		}
	}
	return id
}

func addSlice(parentUID string, wrap *treeWrap, a []any, id int) int {
	var uid string
	for i, v := range a {
		switch v2 := v.(type) {
		case []any:
			uid, id = wrap.add(parentUID, id, formatArrayIndex(i))
			id = addSlice(uid, wrap, v2, id)
		case map[string][]any:
			uid, id = wrap.add(parentUID, id, formatArrayIndex(i))
			id = addObjectWithList(uid, wrap, v2, id)
		case map[string]any:
			uid, id = wrap.add(parentUID, id, formatArrayIndex(i))
			id = addObject(uid, wrap, v2, id)
		case string:
			_, id = wrap.add(parentUID, id, fmt.Sprintf("%s: \"%s\"", formatArrayIndex(i), v2))
		default:
			_, id = wrap.add(parentUID, id, fmt.Sprintf("%s: %s", formatArrayIndex(i), formatValue(v2)))
		}
	}
	return id
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
