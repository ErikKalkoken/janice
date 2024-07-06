// Package ui contains the user interface.
package ui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/ErikKalkoken/jsonviewer/internal/jsondocument"
)

const appTitle = "JSON Viewer"

// setting keys
const (
	settingWindowWidth  = "main-window-width"
	settingWindowHeight = "main-window-height"
	settingRecentFiles  = "recent-files"
)

// UI represents the user interface of this app.
type UI struct {
	app         fyne.App
	document    *jsondocument.JSONDocument
	fileMenu    *fyne.Menu
	statusbar   *widget.Label
	treeWidget  *widget.Tree
	currentFile fyne.URI
	window      fyne.Window
}

// NewUI returns a new UI object.
func NewUI() (*UI, error) {
	a := app.NewWithID("com.github.ErikKalkoken.jsonviewer")
	u := &UI{
		app:       a,
		document:  jsondocument.New(),
		statusbar: widget.NewLabel(""),
		window:    a.NewWindow(appTitle),
	}
	u.treeWidget = u.makeTree()
	c := container.NewBorder(
		nil,
		u.statusbar,
		nil,
		nil,
		u.treeWidget,
	)
	u.window.SetContent(c)
	u.window.SetMainMenu(u.makeMenu())
	u.updateRecentFilesMenu()
	u.window.SetMaster()
	s := fyne.Size{
		Width:  float32(a.Preferences().FloatWithFallback(settingWindowWidth, 800)),
		Height: float32(a.Preferences().FloatWithFallback(settingWindowHeight, 600)),
	}
	u.window.Resize(s)

	u.window.SetOnClosed(func() {
		a.Preferences().SetFloat(settingWindowWidth, float64(u.window.Canvas().Size().Width))
		a.Preferences().SetFloat(settingWindowHeight, float64(u.window.Canvas().Size().Height))
	})
	return u, nil
}

// ShowAndRun shows the main window and runs the app. This method is blocking.
func (u *UI) ShowAndRun() {
	u.window.ShowAndRun()
}

func (u *UI) showErrorDialog(message string, err error) {
	slog.Error(message, "err", err)
	d := dialog.NewInformation("Error", message, u.window)
	d.Show()
}

func (u *UI) setTitle(fileName string) {
	var s string
	if fileName != "" {
		s = fmt.Sprintf("%s - %s", fileName, u.appName())
	} else {
		s = u.appName()
	}
	u.window.SetTitle(s)
}

func (u *UI) appName() string {
	info := u.app.Metadata()
	if info.Name != "" {
		return info.Name
	}
	return appTitle
}

func (u *UI) makeTree() *widget.Tree {
	tree := widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			return u.document.ChildUIDs(id)
		},
		func(id widget.TreeNodeID) bool {
			return u.document.IsBranch(id)
		},
		func(branch bool) fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Key template"), widget.NewLabel("Value template"))
		},
		func(uid widget.TreeNodeID, branch bool, co fyne.CanvasObject) {
			node := u.document.Value(uid)
			hbox := co.(*fyne.Container)
			key := hbox.Objects[0].(*widget.Label)
			key.SetText(fmt.Sprintf("%s :", node.Key))
			value := hbox.Objects[1].(*widget.Label)
			var text string
			var importance widget.Importance
			switch v := node.Value; node.Type {
			case jsondocument.Array:
				importance = widget.HighImportance
				if branch {
					if t := u.treeWidget; t != nil && t.IsBranchOpen(uid) {
						text = ""
					} else {
						text = "[...]"
					}
				} else {
					text = "[]"
				}
			case jsondocument.Object:
				importance = widget.HighImportance
				if branch {
					if t := u.treeWidget; t != nil && t.IsBranchOpen(uid) {
						text = ""
					} else {
						text = "{...}"
					}
				} else {
					text = "{}"
				}
			case jsondocument.String:
				importance = widget.WarningImportance
				text = fmt.Sprintf("\"%s\"", v)
			case jsondocument.Number:
				importance = widget.SuccessImportance
				text = fmt.Sprintf("%v", v)
			case jsondocument.Boolean:
				importance = widget.DangerImportance
				text = fmt.Sprintf("%v", v)
			case jsondocument.Null:
				importance = widget.DangerImportance
				text = "null"
			default:
				text = fmt.Sprintf("%v", v)
			}
			value.Text = text
			value.Importance = importance
			value.Refresh()
		})

	tree.OnSelected = func(uid widget.TreeNodeID) {
		defer u.treeWidget.UnselectAll()
		if u.document.IsBranch(uid) {
			tree.ToggleBranch(uid)
		}
	}
	return tree
}
