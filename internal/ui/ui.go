// Package ui contains the user interface.
package ui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/ErikKalkoken/jsonviewer/internal/github"
	"github.com/ErikKalkoken/jsonviewer/internal/jsondocument"
)

const (
	appTitle    = "JSON Viewer"
	githubOwner = "ErikKalkoken"
	githubRepo  = "jsonviewer"
)

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
func NewUI(app fyne.App) (*UI, error) {
	u := &UI{
		app:            app,
		document:       jsondocument.New(),
		detailPath:     widget.NewLabel(""),
		detailType:     widget.NewLabel(""),
		detailValueMD:  widget.NewRichText(),
		statusTreeSize: widget.NewLabel(""),
		statusPath:     widget.NewLabel(""),
		searchEntry:    widget.NewEntry(),
		window:         app.NewWindow(appTitle),
	}
	u.treeWidget = u.makeTree()
	u.detailPath.Wrapping = fyne.TextWrapBreak
	u.detailValueMD.Wrapping = fyne.TextWrapWord
	u.statusPath.Wrapping = fyne.TextWrapWord

	// search frame
	u.searchEntry.SetPlaceHolder(
		"Enter pattern with wildcards to search for a key...")
	u.searchEntry.OnSubmitted = func(s string) {
		u.searchKey()
	}
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
		container.NewHBox(u.searchButton, u.collapseButton),
		u.searchEntry,
	)

	// main frame
	welcomeText := widget.NewLabel(
		"Welcome to JSON Viewer.\n" +
			"Open a JSON file by dropping it on the window\n" +
			"or through File / Open File\n" +
			"or by importing it from clipboard.\n",
	)
	welcomeText.Importance = widget.LowImportance
	welcomeText.Alignment = fyne.TextAlignCenter
	u.welcomeMessage = container.NewCenter(welcomeText)
	document := container.NewBorder(
		searchBar,
		nil,
		nil,
		nil,
		container.NewStack(u.welcomeMessage, u.treeWidget),
	)

	// detail frame
	u.detailCopyValue = widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		u.window.Clipboard().SetContent(u.detailValueRaw)
	})
	u.detailCopyValue.Disable()
	u.jumpToSelection = widget.NewButtonWithIcon("", theme.NewThemedResource(resourceReadmoreSvg), func() {
		u.showInTree(u.currentSelectedUID)
	})
	u.jumpToSelection.Disable()
	detail := container.NewBorder(
		container.NewBorder(
			container.NewHBox(layout.NewSpacer(), u.jumpToSelection, u.detailCopyValue),
			u.detailType,
			nil,
			nil,
			u.detailPath,
		),
		nil,
		nil,
		nil,
		u.detailValueMD,
	)

	hsplit := container.NewHSplit(document, detail)
	hsplit.Offset = 0.75

	statusBar := container.NewHBox(u.statusTreeSize)
	notifyUpdates := u.app.Preferences().BoolWithFallback(settingsNotifyUpdates, settingsNotifyUpdatesDefault)
	if notifyUpdates {
		go func() {
			current := u.app.Metadata().Version
			latestVersion, isNewer, err := github.AvailableUpdate(githubOwner, githubRepo, current)
			if err != nil {
				slog.Error("Failed to fetch latest version from github", "err", err)
				return
			}
			if !isNewer {
				return
			}
			url, _ := url.Parse("https://github.com/ErikKalkoken/jsonviewer/releases")
			statusBar.Add(layout.NewSpacer())
			statusBar.Add(widget.NewHyperlink(fmt.Sprintf("New version %s available", latestVersion), url))
		}()
	}

	c := container.NewBorder(nil, container.NewVBox(widget.NewSeparator(), statusBar), nil, nil, hsplit)

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
		Width:  float32(app.Preferences().FloatWithFallback(settingWindowWidth, 800)),
		Height: float32(app.Preferences().FloatWithFallback(settingWindowHeight, 600)),
	}
	u.window.Resize(s)
	u.window.SetOnClosed(func() {
		app.Preferences().SetFloat(settingWindowWidth, float64(u.window.Canvas().Size().Width))
		app.Preferences().SetFloat(settingWindowHeight, float64(u.window.Canvas().Size().Height))
	})
	return u, nil
}

func (u *UI) searchKey() {
	pattern := u.searchEntry.Text
	if len(pattern) == 0 {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	spinner := widget.NewActivity()
	spinner.Start()
	c := container.NewHBox(widget.NewLabel(fmt.Sprintf("Searching for key with pattern: %s", pattern)), spinner)
	b := widget.NewButton("Cancel", func() {
		cancel()
	})
	d := dialog.NewCustomWithoutButtons("Search", container.NewVBox(c, b), u.window)
	d.Show()
	d.SetOnClosed(func() {
		cancel()
	})
	go func() {
		uid, err := u.document.SearchKey(ctx, u.currentSelectedUID, pattern)
		d.Hide()
		if errors.Is(err, jsondocument.ErrCallerCanceled) {
			return
		} else if errors.Is(err, jsondocument.ErrNotFound) {
			d2 := dialog.NewInformation("No match", fmt.Sprintf("No key found matching %s", pattern), u.window)
			d2.Show()
			return
		} else if err != nil {
			u.showErrorDialog("Search failed", err)
			return
		}
		u.showInTree(uid)
	}()
}

// ShowAndRun shows the main window and runs the app. This method is blocking.
func (u *UI) ShowAndRun(path string) {
	if path != "" {
		u.app.Lifecycle().SetOnStarted(func() {
			uri := storage.NewFileURI(path)
			reader, err := storage.Reader(uri)
			if err != nil {
				u.showErrorDialog(fmt.Sprintf("Failed to open file: %s", uri), err)
				return
			}
			u.loadDocument(reader)
		})
	}
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
				x := v.(float64)
				text = strconv.FormatFloat(x, 'f', -1, 64)
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
		u.jumpToSelection.Enable()
		typeText := fmt.Sprint(node.Type)
		var v string
		if u.document.IsBranch(uid) {
			u.detailCopyValue.Disable()
			v = "..."
			ids := u.document.ChildUIDs(uid)
			typeText += fmt.Sprintf(", %d elements", len(ids))
		} else {
			u.detailCopyValue.Enable()
			u.detailValueRaw = fmt.Sprint(node.Value)
			switch node.Type {
			case jsondocument.String:
				v = fmt.Sprintf("\"%s\"", node.Value)
			case jsondocument.Number:
				x := node.Value.(float64)
				v = strconv.FormatFloat(x, 'f', -1, 64)
			case jsondocument.Null:
				v = "null"
			default:
				v = u.detailValueRaw
			}
		}
		u.detailType.SetText(typeText)
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
