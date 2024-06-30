package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ui struct {
	app      fyne.App
	tree     *widget.Tree
	treeWrap *JSONTree
	window   fyne.Window
}

func newUI() *ui {
	u := &ui{
		app:      app.New(),
		treeWrap: NewJSONTree(),
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
		id = u.treeWrap.addObject("", v, 0)
	case []any:
		id = u.treeWrap.addSlice("", v, 0)
	default:
		return fmt.Errorf("unrecognized format")
	}
	log.Printf("Loaded JSON file into tree with %d nodes", id)
	u.tree.Refresh()
	return nil
}
