package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"example/jsonviewer/internal/jsontree"
)

type UI struct {
	app        fyne.App
	treeWidget *widget.Tree
	treeData   *jsontree.JSONTree
	window     fyne.Window
}

func NewUI() *UI {
	u := &UI{
		app:      app.New(),
		treeData: jsontree.NewJSONTree(),
	}
	u.window = u.app.NewWindow("JSON Viewer")
	u.treeWidget = makeTree(u)
	b1 := widget.NewButton("Close All", func() {
		u.treeWidget.CloseAllBranches()
	})
	c := container.NewBorder(container.NewHBox(b1), nil, nil, nil, u.treeWidget)
	u.window.SetContent(c)
	u.window.Resize(fyne.Size{Width: 800, Height: 600})
	return u
}

func (u *UI) ShowAndRun() {
	u.window.ShowAndRun()
}

func (u *UI) SetData(data any) error {
	id, err := u.treeData.Set(data)
	if err != nil {
		return err
	}
	log.Printf("Loaded JSON file into tree with %d nodes", id)
	u.treeWidget.Refresh()
	return nil
}

func makeTree(u *UI) *widget.Tree {
	tree := widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			return u.treeData.ChildUIDs(id)
		},
		func(id widget.TreeNodeID) bool {
			return u.treeData.IsBranch(id)
		},
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Leaf template")
		},
		func(uid widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			text := u.treeData.Value(uid)
			o.(*widget.Label).SetText(text)
		})

	tree.OnSelected = func(uid widget.TreeNodeID) {
		u.treeWidget.UnselectAll()
	}
	return tree
}
