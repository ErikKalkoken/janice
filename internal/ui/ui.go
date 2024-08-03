// Package ui contains the user interface.
package ui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	fynetooltip "github.com/dweymouth/fyne-tooltip"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/ErikKalkoken/janice/internal/jsondocument"
	"github.com/ErikKalkoken/janice/internal/widgets"
)

// preference keys
const (
	preferenceLastSelectionFrameShown = "last-selection-frame-shown"
	preferenceLastValueFrameShown     = "last-value-frame-shown"
	preferenceLastWindowHeight        = "last-window-height"
	preferenceLastWindowWidth         = "last-window-width"
)

// UI represents the user interface of this app.
type UI struct {
	app    fyne.App
	window fyne.Window

	currentFile    fyne.URI
	document       *jsondocument.JSONDocument
	fileMenu       *fyne.Menu
	viewMenu       *fyne.Menu
	treeWidget     *widget.Tree
	welcomeMessage *fyne.Container

	searchBar *searchBarFrame
	selection *selectionFrame
	statusBar *statusBarFrame
	value     *valueFrame
}

// NewUI returns a new UI object.
func NewUI(app fyne.App) (*UI, error) {
	appName := app.Metadata().Name
	u := &UI{
		app:      app,
		document: jsondocument.New(),
		window:   app.NewWindow(appName),
	}
	u.treeWidget = u.makeTree()

	// main frame
	welcomeText := widget.NewLabel(
		"Welcome to " + appName + ".\n" +
			"Open a JSON file through the File Open menu\n" +
			"or drag and drop the file on this window\n" +
			"or import from the clipboard.\n",
	)
	welcomeText.Importance = widget.MediumImportance
	welcomeText.Alignment = fyne.TextAlignCenter
	u.welcomeMessage = container.NewCenter(welcomeText)

	u.searchBar = u.newSearchBarFrame()
	u.selection = u.newSelectionFrame()
	u.statusBar = u.newStatusBarFrame()
	u.value = u.newValueFrame()

	c := container.NewBorder(
		container.NewVBox(u.searchBar.content, u.selection.content, u.value.content, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), u.statusBar.content),
		nil,
		nil,
		container.NewStack(u.welcomeMessage, u.treeWidget))

	u.window.SetContent(fynetooltip.AddWindowToolTipLayer(c, u.window.Canvas()))
	u.window.SetMainMenu(u.makeMenu())
	u.toogleHasDocument(false)
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
		Width:  float32(app.Preferences().FloatWithFallback(preferenceLastWindowWidth, 800)),
		Height: float32(app.Preferences().FloatWithFallback(preferenceLastWindowHeight, 600)),
	}
	u.window.Resize(s)
	u.setTheme(app.Preferences().StringWithFallback(settingTheme, settingThemeDefault))
	u.window.SetOnClosed(func() {
		app.Preferences().SetFloat(preferenceLastWindowWidth, float64(u.window.Canvas().Size().Width))
		app.Preferences().SetFloat(preferenceLastWindowHeight, float64(u.window.Canvas().Size().Height))
		app.Preferences().SetBool(preferenceLastValueFrameShown, u.value.isShown())
		app.Preferences().SetBool(preferenceLastSelectionFrameShown, u.selection.isShown())
	})
	return u, nil
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
			return widgets.NewTreeNode()
		},
		func(uid widget.TreeNodeID, branch bool, co fyne.CanvasObject) {
			node := u.document.Value(uid)
			obj := co.(*widgets.TreeNode)
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
		u.selectElement(uid)
	}
	return tree
}

func (u *UI) selectElement(uid string) {
	u.selection.set(uid)
	u.selection.enable()
	u.value.set(uid)
	u.fileMenu.Items[7].Disabled = false
	u.fileMenu.Items[8].Disabled = false
	u.fileMenu.Refresh()
}

// ShowAndRun shows the main window and runs the app. This method is blocking.
func (u *UI) ShowAndRun(path string) {
	if path != "" {
		u.app.Lifecycle().SetOnStarted(func() {
			path2, err := filepath.Abs(path)
			if err != nil {
				u.showErrorDialog(fmt.Sprintf("Not a valid path: %s", path), err)
				return
			}
			uri := storage.NewFileURI(path2)
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

func (u *UI) scrollTo(uid widget.TreeNodeID) {
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
	u.statusBar.reset()
	u.welcomeMessage.Show()
	u.toogleHasDocument(false)
	u.selection.reset()
	u.value.reset()
}

func (u *UI) setTitle(fileName string) {
	var s string
	name := u.app.Metadata().Name
	if fileName != "" {
		s = fmt.Sprintf("%s - %s", fileName, name)
	} else {
		s = name
	}
	u.window.SetTitle(s)
}

func (u *UI) setTheme(themeName string) {
	switch themeName {
	case themeDark:
		u.app.Settings().SetTheme(theme.DarkTheme())
	case themeLight:
		u.app.Settings().SetTheme(theme.LightTheme())
	}
}

// loadDocument loads a JSON file
// Shows a loader modal while loading
func (u *UI) loadDocument(reader fyne.URIReadCloser) {
	infoText := widget.NewLabel("")
	pb1 := widget.NewProgressBarInfinite()
	pb2 := widget.NewProgressBar()
	pb2.Hide()
	progressInfo := binding.NewUntyped()
	progressInfo.AddListener(binding.NewDataListener(func() {
		x, err := progressInfo.Get()
		if err != nil {
			slog.Warn("Failed to get progress info", "err", err)
			return
		}
		info, ok := x.(jsondocument.ProgressInfo)
		if !ok {
			return
		}
		uri := reader.URI()
		name := uri.Name()
		var text string
		switch info.CurrentStep {
		case 1:
			text = fmt.Sprintf("Loading file from disk: %s", name)
		case 2:
			text = fmt.Sprintf("Calculating document size: %s", name)
		case 3:
			if pb2.Hidden {
				pb1.Stop()
				pb1.Hide()
				pb2.Show()
			}
			p := message.NewPrinter(language.English)
			text = p.Sprintf("Rendering document with %d elements: %s", info.Size, name)
			pb2.SetValue(info.Progress)
		default:
			text = "?"
		}
		message := fmt.Sprintf("%d / %d: %s", info.CurrentStep, info.TotalSteps, text)
		infoText.SetText(message)
	}))
	ctx, cancel := context.WithCancel(context.TODO())
	b := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		cancel()
	})
	c := container.NewVBox(
		infoText,
		container.NewBorder(nil, nil, nil, b, container.NewStack(pb1, pb2)),
	)
	d2 := dialog.NewCustomWithoutButtons("Loading", c, u.window)
	d2.SetOnClosed(func() {
		cancel()
	})
	d2.Show()
	go func() {
		doc := jsondocument.New()
		if err := doc.Load(ctx, reader, progressInfo); err != nil {
			d2.Hide()
			if errors.Is(err, jsondocument.ErrCallerCanceled) {
				return
			}
			u.showErrorDialog(fmt.Sprintf("Failed to open document: %s", reader.URI()), err)
			return
		}
		u.document = doc
		u.statusBar.set(u.document.Size())
		u.welcomeMessage.Hide()
		u.toogleHasDocument(true)
		if doc.Size() > 1000 {
			u.viewMenu.Items[4].Disabled = true
		} else {
			u.viewMenu.Items[4].Disabled = false
		}
		u.viewMenu.Refresh()
		u.treeWidget.Refresh()
		uri := reader.URI()
		if uri.Scheme() == "file" {
			u.addRecentFile(uri)
		}
		u.setTitle(uri.Name())
		u.currentFile = uri
		u.selection.reset()
		u.value.reset()
		d2.Hide()
	}()
}

func (u *UI) toogleHasDocument(enabled bool) {
	if enabled {
		u.searchBar.enable()
		u.fileMenu.Items[0].Disabled = false
		u.fileMenu.Items[5].Disabled = false
		u.fileMenu.Items[7].Disabled = u.selection.selectedUID == ""
		u.fileMenu.Items[8].Disabled = u.selection.selectedUID == ""
		u.viewMenu.Items[0].Disabled = false
		u.viewMenu.Items[1].Disabled = false
		u.viewMenu.Items[2].Disabled = false
		u.viewMenu.Items[4].Disabled = false
		u.viewMenu.Items[5].Disabled = false
	} else {
		u.searchBar.disable()
		u.fileMenu.Items[0].Disabled = true
		u.fileMenu.Items[5].Disabled = true
		u.fileMenu.Items[7].Disabled = true
		u.fileMenu.Items[8].Disabled = true
		u.viewMenu.Items[0].Disabled = true
		u.viewMenu.Items[1].Disabled = true
		u.viewMenu.Items[2].Disabled = true
		u.viewMenu.Items[4].Disabled = true
		u.viewMenu.Items[5].Disabled = true
	}
	u.fileMenu.Refresh()
	u.viewMenu.Refresh()
}
