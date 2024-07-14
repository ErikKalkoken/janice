// Package ui contains the user interface.
package ui

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/ErikKalkoken/jsonviewer/internal/jsondocument"
)

const appTitle = "JSON Viewer"

// setting keys
const (
	settingWindowWidth  = "main-window-width"
	settingWindowHeight = "main-window-height"
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
	app                fyne.App
	currentFile        fyne.URI
	currentSelectedUID widget.TreeNodeID
	detailCopyValue    *widget.Button
	jumpToSelection    *widget.Button
	collapseButton     *widget.Button
	detailPath         *widget.Label
	detailType         *widget.Label
	detailValueMD      *widget.RichText
	detailValueRaw     string
	document           *jsondocument.JSONDocument
	fileMenu           *fyne.Menu
	searchEntry        *widget.Entry
	searchButton       *widget.Button
	statusPath         *widget.Label
	statusTreeSize     *widget.Label
	treeWidget         *widget.Tree
	welcomeMessage     *fyne.Container
	window             fyne.Window
}

// NewUI returns a new UI object.
func NewUI() (*UI, error) {
	a := app.NewWithID("com.github.ErikKalkoken.jsonviewer")
	u := &UI{
		app:            a,
		document:       jsondocument.New(),
		detailPath:     widget.NewLabel(""),
		detailType:     widget.NewLabel(""),
		detailValueMD:  widget.NewRichText(),
		statusTreeSize: widget.NewLabel(""),
		statusPath:     widget.NewLabel(""),
		searchEntry:    widget.NewEntry(),
		window:         a.NewWindow(appTitle),
	}
	u.treeWidget = u.makeTree()
	u.detailPath.Wrapping = fyne.TextWrapBreak
	u.detailValueMD.Wrapping = fyne.TextWrapWord
	u.statusPath.Wrapping = fyne.TextWrapWord
	u.detailCopyValue = widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		u.window.Clipboard().SetContent(u.detailValueRaw)
	})
	u.detailCopyValue.Disable()
	u.jumpToSelection = widget.NewButtonWithIcon("", theme.NewThemedResource(resourceReadmoreSvg), func() {
		u.showInTree(u.currentSelectedUID)
	})
	u.jumpToSelection.Disable()
	u.searchEntry.OnSubmitted = func(s string) {
		u.searchKey()
	}
	detail := container.NewBorder(
		container.NewBorder(nil, u.detailType, nil, nil, u.detailPath),
		nil,
		nil,
		nil,
		u.detailValueMD,
	)
	welcomeText := widget.NewLabel(
		"Welcome to JSON Viewer.\n" +
			"Open a JSON file by dropping it on the window\n" +
			"or through the File menu.",
	)
	welcomeText.Importance = widget.LowImportance
	welcomeText.Alignment = fyne.TextAlignCenter
	u.welcomeMessage = container.NewCenter(welcomeText)
	hsplit := container.NewHSplit(container.NewStack(u.welcomeMessage, u.treeWidget), detail)
	hsplit.Offset = 0.75
	u.searchEntry.Disable()
	u.searchButton = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		u.searchKey()
	})
	u.searchButton.Disable()
	u.collapseButton = widget.NewButtonWithIcon("", theme.NewThemedResource(resourceUnfoldlessSvg), func() {
		u.treeWidget.CloseAllBranches()
	})
	u.collapseButton.Disable()
	searchBar := container.NewBorder(
		nil,
		nil,
		nil,
		container.NewHBox(
			u.searchButton,
			layout.NewSpacer(),
			container.NewPadded(),
			u.collapseButton,
			container.NewPadded(),
			layout.NewSpacer(),
			u.jumpToSelection,
			u.detailCopyValue,
		),
		u.searchEntry,
	)
	c := container.NewBorder(
		searchBar,
		container.NewVBox(widget.NewSeparator(), u.statusTreeSize),
		nil,
		nil,
		hsplit,
	)
	u.window.SetContent(c)
	u.window.SetMainMenu(u.makeMenu())
	u.updateRecentFilesMenu()
	u.window.SetMaster()
	u.window.SetOnDropped(func(p fyne.Position, uris []fyne.URI) {
		if len(uris) < 1 {
			return
		}
		uri := uris[0]
		slog.Info("Loading dropped file", "uri", uri)
		reader, err := storage.Reader(uri)
		if err != nil {
			u.showErrorDialog(fmt.Sprintf("Failed to load file: %s", uri), err)
			return
		}
		u.loadDocument(reader)
	})
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

func (u *UI) searchKey() {
	text := u.searchEntry.Text
	if len(text) == 0 {
		return
	}
	uid, err := u.document.SearchKey(u.currentSelectedUID, text)
	if errors.Is(err, jsondocument.ErrNotFound) {
		d := dialog.NewInformation("Not found", fmt.Sprintf("The key %s was not found", text), u.window)
		d.Show()
		return
	} else if err != nil {
		u.showErrorDialog("Search failed", err)
		return
	}
	u.showInTree(uid)
}

// ShowAndRun shows the main window and runs the app. This method is blocking.
func (u *UI) ShowAndRun() {
	u.window.ShowAndRun()
}

func (u *UI) showErrorDialog(message string, err error) {
	if err != nil {
		slog.Error(message, "err", err)
	}
	d := dialog.NewInformation("Error", message, u.window)
	d.Show()
}

func (u *UI) showInTree(uid widget.TreeNodeID) {
	if uid == "" {
		return
	}
	p := u.document.Path(uid)
	for _, uid2 := range p {
		u.treeWidget.OpenBranch(uid2)
	}
	u.treeWidget.ScrollTo(uid)
	u.treeWidget.Select(uid)
}

// reset resets the app to it's initial state
func (u *UI) reset() {
	u.document.Reset()
	u.setTitle("")
	u.statusTreeSize.SetText("")
	u.welcomeMessage.Show()
	u.searchButton.Disable()
	u.searchEntry.Disable()
	u.collapseButton.Disable()
	u.detailPath.SetText("")
	u.detailType.SetText("")
	u.detailValueMD.ParseMarkdown("")
	u.detailCopyValue.Disable()
	u.jumpToSelection.Disable()
	u.currentSelectedUID = ""
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
		u.currentSelectedUID = uid
		path := u.renderPath(uid)
		u.statusPath.SetText(path)
		node := u.document.Value(uid)
		u.detailPath.SetText(path)
		u.detailType.SetText(fmt.Sprint(node.Type))
		u.detailCopyValue.Enable()
		var v string
		if u.document.IsBranch(uid) {
			u.detailCopyValue.Disable()
			v = "..."
		} else {
			u.jumpToSelection.Enable()
			u.detailValueRaw = fmt.Sprint(node.Value)
			switch node.Type {
			case jsondocument.String:
				v = fmt.Sprintf("\"%s\"", node.Value)
			case jsondocument.Null:
				v = "null"
			default:
				v = u.detailValueRaw
			}
		}
		u.detailValueMD.ParseMarkdown(fmt.Sprintf("```\n%s\n```", v))
	}
	return tree
}

func (u *UI) renderPath(uid string) string {
	p := u.document.Path(uid)
	keys := []string{"$"}
	for _, id := range p {
		node := u.document.Value(id)
		keys = append(keys, node.Key)
	}
	node := u.document.Value(uid)
	keys = append(keys, node.Key)
	return strings.Join(keys, ".")
}
