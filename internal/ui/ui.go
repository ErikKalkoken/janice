// Package ui contains the user interface.
package ui

import (
	"fmt"
	"log/slog"
	"net/url"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

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
		document:  jsondocument.NewJSONDocument(),
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
		s = fmt.Sprintf("%s - %s", fileName, appTitle)
	} else {
		s = appTitle
	}
	u.window.SetTitle(s)
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

func (u *UI) makeMenu() *fyne.MainMenu {
	recentItem := fyne.NewMenuItem("Open Recent", nil)
	recentItem.ChildMenu = fyne.NewMenu("")
	u.fileMenu = fyne.NewMenu("File",
		fyne.NewMenuItem("Open File...", func() {
			dialogOpen := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					u.showErrorDialog("Failed to read folder", err)
					return
				}
				if reader == nil {
					return
				}
				defer reader.Close()
				if err := u.loadDocument(reader); err != nil {
					u.showErrorDialog("Failed to load document", err)
					return
				}
			}, u.window)
			// f := storage.NewExtensionFileFilter([]string{".json"})
			// dialogOpen.SetFilter(f)
			dialogOpen.Show()
		}),
		recentItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Reload", func() {
			reader, err := storage.Reader(u.currentFile)
			if err != nil {
				u.showErrorDialog("Failed to open file", err)
				return
			}
			defer reader.Close()
			if err := u.loadDocument(reader); err != nil {
				u.showErrorDialog("Failed to load document", err)
				return
			}
		}),
	)
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Expand All", func() {
			u.treeWidget.OpenAllBranches()
		}),
		fyne.NewMenuItem("Collapse All", func() {
			u.treeWidget.CloseAllBranches()
		}),
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Report a bug", func() {
			url, _ := url.Parse("https://github.com/ErikKalkoken/jsonviewer/issue")
			_ = u.app.OpenURL(url)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("About...", func() {
			u.showAboutDialog()
		}),
	)
	main := fyne.NewMainMenu(u.fileMenu, viewMenu, helpMenu)
	return main
}

func (u *UI) updateRecentFilesMenu() {
	files := u.app.Preferences().StringList(settingRecentFiles)
	items := make([]*fyne.MenuItem, len(files))
	for i, f := range files {
		uri, err := storage.ParseURI(f)
		if err != nil {
			slog.Error("Failed to parse URI", "URI", f, "err", err)
			continue
		}
		items[i] = fyne.NewMenuItem(uri.Path(), func() {
			reader, err := storage.Reader(uri)
			if err != nil {
				dialog.ShowError(err, u.window)
				return
			}
			defer reader.Close()
			if err := u.loadDocument(reader); err != nil {
				dialog.ShowError(err, u.window)
				return
			}
		})
	}
	u.fileMenu.Items[1].ChildMenu.Items = items
	u.fileMenu.Refresh()
}

func (u *UI) addRecentFile(uri fyne.URI) {
	files := u.app.Preferences().StringList(settingRecentFiles)
	uri2 := uri.String()
	files = addToListWithRotation(files, uri2, 5)
	u.app.Preferences().SetStringList(settingRecentFiles, files)
	u.updateRecentFilesMenu()
}

func (u *UI) showAboutDialog() {
	c := container.NewVBox()
	info := u.app.Metadata()
	appData := widget.NewRichTextFromMarkdown(
		"## " + appTitle + "\n**Version:** " + info.Version)
	c.Add(appData)
	x, _ := url.Parse("https://github.com/ErikKalkoken/jsombuddy")
	c.Add(widget.NewHyperlink("Website", x))
	c.Add(widget.NewLabel("(c) 2024 Erik Kalkoken"))
	d := dialog.NewCustom("About", "OK", c, u.window)
	d.Show()
}

func (u *UI) loadDocument(reader fyne.URIReadCloser) error {
	defer reader.Close()
	text1 := widget.NewLabel("Loading file 1 / 2. Please wait...")
	text2 := widget.NewLabel("")
	pb := widget.NewProgressBarInfinite()
	pb.Start()
	c := container.NewVBox(text1, container.NewStack(pb, text2))
	d2 := dialog.NewCustomWithoutButtons("Loading", c, u.window)
	d2.Show()
	defer d2.Hide()
	text1.SetText("Loading file 2 / 2. Please wait...")
	currentSize := binding.NewInt()
	currentSize.AddListener(binding.NewDataListener(func() {
		v, err := currentSize.Get()
		if err != nil {
			slog.Error("Failed to retrieve value for current size", "err", err)
			return
		}
		p := message.NewPrinter(language.English)
		t := p.Sprintf("%d elements loaded", v)
		text2.SetText(t)
	}))
	if err := u.document.Load(reader, currentSize); err != nil {
		return err
	}
	p := message.NewPrinter(language.English)
	out := p.Sprintf("Size: %d", u.document.Size())
	u.statusbar.SetText(out)
	u.treeWidget.Refresh()
	x := reader.URI()
	u.setTitle(x.Name())
	u.addRecentFile(x)
	u.currentFile = x
	return nil
}

func addToListWithRotation(s []string, v string, max int) []string {
	if max < 1 {
		panic("max must be 1 or higher")
	}
	i := slices.Index(s, v)
	if i != -1 {
		s = slices.Delete(s, i, i+1)
	}
	s = slices.Insert(s, 0, v)
	if len(s) > max {
		s = s[0:max]
	}
	return s
}
