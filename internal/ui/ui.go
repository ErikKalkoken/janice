// Package ui contains the user interface.
package ui

import (
	"fmt"
	"log/slog"
	"strings"

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

var type2importance = map[jsondocument.JSONType]widget.Importance{
	jsondocument.Array:   widget.HighImportance,
	jsondocument.Object:  widget.HighImportance,
	jsondocument.String:  widget.WarningImportance,
	jsondocument.Number:  widget.SuccessImportance,
	jsondocument.Boolean: widget.DangerImportance,
	jsondocument.Null:    widget.DangerImportance,
}

// UI represents the user interface of this app.
type UI struct {
	app         fyne.App
	detailPath  *widget.Label
	detailType  *widget.Label
	detailValue *widget.RichText
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
		app:         a,
		document:    jsondocument.New(),
		detailPath:  widget.NewLabel(""),
		detailType:  widget.NewLabel(""),
		detailValue: widget.NewRichText(),
		statusbar:   widget.NewLabel(""),
		window:      a.NewWindow(appTitle),
	}
	u.treeWidget = u.makeTree()
	u.detailPath.Wrapping = fyne.TextWrapWord
	u.detailValue.Wrapping = fyne.TextWrapWord
	detail := container.NewBorder(
		container.NewBorder(nil, u.detailType, nil, nil, u.detailPath), nil, nil, nil, u.detailValue,
	)
	hsplit := container.NewHSplit(u.treeWidget, detail)
	hsplit.Offset = 0.75
	c := container.NewBorder(nil, u.statusbar, nil, nil, hsplit)
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
			return NewNodeWidget()
		},
		func(uid widget.TreeNodeID, branch bool, co fyne.CanvasObject) {
			node := u.document.Value(uid)
			obj := co.(*NodeWidget)
			var text string
			switch v := node.Value; node.Type {
			case jsondocument.Array:
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
				text = fmt.Sprintf("\"%s\"", v)
			case jsondocument.Number:
				text = fmt.Sprintf("%v", v)
			case jsondocument.Boolean:
				text = fmt.Sprintf("%v", v)
			case jsondocument.Null:
				text = "null"
			default:
				text = fmt.Sprintf("%v", v)
			}
			obj.Set(node.Key, text, type2importance[node.Type])
		})

	tree.OnSelected = func(uid widget.TreeNodeID) {
		if u.document.IsBranch(uid) {
			u.treeWidget.UnselectAll()
			return
		}
		p := u.document.Path(uid)
		keys := []string{"$"}
		for _, id := range p {
			node := u.document.Value(id)
			keys = append(keys, node.Key)
		}
		node := u.document.Value(uid)
		keys = append(keys, node.Key)
		u.detailPath.SetText(strings.Join(keys, "."))
		u.detailType.SetText(fmt.Sprint(node.Type))
		var v string
		switch node.Type {
		case jsondocument.String:
			v = fmt.Sprintf("\"%s\"", node.Value)
		case jsondocument.Null:
			v = "null"
		default:
			v = fmt.Sprint(node.Value)
		}
		u.detailValue.ParseMarkdown(fmt.Sprintf("```\n%s\n```", v))
	}
	return tree
}
