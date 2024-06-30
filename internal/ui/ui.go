package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"example/jsonviewer/internal/jsondocument"
)

const appTitle = "JSON Viewer"

// UI represents the user interface of this app.
type UI struct {
	app        fyne.App
	treeWidget *widget.Tree
	document   jsondocument.JSONDocument
	window     fyne.Window
}

// NewUI returns a new UI object.
func NewUI() (*UI, error) {
	u := &UI{app: app.New()}
	x, err := jsondocument.NewJSONDocument()
	if err != nil {
		return nil, err
	}
	u.document = x
	u.window = u.app.NewWindow(appTitle)
	u.treeWidget = makeTree(u)
	tb := widget.NewToolbar(
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.NewThemedResource(resourceUnfoldlessSvg), func() {
			u.treeWidget.CloseAllBranches()
		}),
	)
	c := container.NewBorder(
		container.NewVBox(tb, widget.NewSeparator()),
		nil,
		nil,
		nil,
		u.treeWidget,
	)
	u.window.SetContent(c)
	u.window.Resize(fyne.Size{Width: 800, Height: 600})
	u.window.SetMainMenu(makeMenu(u))
	u.window.SetMaster()
	return u, nil
}

// ShowAndRun shows the main window and runs the app. This method is blocking.
func (u *UI) ShowAndRun() {
	u.window.ShowAndRun()
}

func (u *UI) setData(data any, sizeEstimate int) error {
	n := int(0.75 * float64(sizeEstimate))
	if err := u.document.Load(data, n); err != nil {
		return err
	}
	log.Printf("Loaded JSON file into tree with %d nodes", u.document.Size())
	u.treeWidget.Refresh()
	return nil
}

func (u *UI) setTitle(fileName string) {
	var s string
	if fileName != "" {
		s = fmt.Sprintf("%s - %s", fileName, appTitle)
	} else {
		s = appTitle
	}
	u.window.SetTitle(s)
}

func makeTree(u *UI) *widget.Tree {
	tree := widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			return u.document.ChildUIDs(id)
		},
		func(id widget.TreeNodeID) bool {
			return u.document.IsBranch(id)
		},
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Leaf template")
		},
		func(uid widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			text := u.document.Value(uid)
			o.(*widget.Label).SetText(text)
		})

	tree.OnSelected = func(uid widget.TreeNodeID) {
		u.treeWidget.UnselectAll()
	}
	return tree
}
