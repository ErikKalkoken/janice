// Package ui contains the user interface.
package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
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
	app        fyne.App
	document   jsondocument.JSONDocument
	fileMenu   *fyne.Menu
	statusbar  *widget.Label
	treeWidget *widget.Tree
	window     fyne.Window
}

// NewUI returns a new UI object.
func NewUI() (*UI, error) {
	a := app.NewWithID("com.github.ErikKalkoken.jsonviewer")
	u := &UI{
		app:       a,
		statusbar: widget.NewLabel(""),
		window:    a.NewWindow(appTitle),
	}
	d, err := jsondocument.NewJSONDocument()
	if err != nil {
		return nil, err
	}
	u.document = d
	u.treeWidget = makeTree(u)
	tb := widget.NewToolbar(
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.NewThemedResource(resourceUnfoldlessSvg), func() {
			u.treeWidget.CloseAllBranches()
		}),
	)
	c := container.NewBorder(
		tb,
		u.statusbar,
		nil,
		nil,
		u.treeWidget,
	)
	u.window.SetContent(c)
	u.window.SetMainMenu(makeMenu(u))
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

func (u *UI) loadData(data any, sizeEstimate int) error {
	n := int(0.75 * float64(sizeEstimate))
	if err := u.document.Load(data, n); err != nil {
		return err
	}
	p := message.NewPrinter(language.English)
	log.Printf("Loaded JSON file into tree with %d nodes", u.document.Size())
	out := p.Sprintf("Size: %d", u.document.Size())
	u.statusbar.SetText(out)
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

func (u *UI) updateRecentFilesMenu() {
	files := u.app.Preferences().StringList(settingRecentFiles)
	items := make([]*fyne.MenuItem, len(files))
	for i, f := range files {
		uri, err := storage.ParseURI(f)
		if err != nil {
			log.Printf("Failed to parse URI %s: %s", f, err)
			continue
		}
		items[i] = fyne.NewMenuItem(uri.Path(), func() {
			reader, err := storage.Reader(uri)
			if err != nil {
				log.Printf("Failed to read from URI %s: %s", uri, err)
				return
			}
			if err := u.loadDocument(reader); err != nil {
				dialog.ShowError(err, u.window)
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

func makeMenu(u *UI) *fyne.MainMenu {
	recentItem := fyne.NewMenuItem("Open Recent", nil)
	recentItem.ChildMenu = fyne.NewMenu("")
	u.fileMenu = fyne.NewMenu("File",
		fyne.NewMenuItem("Open File...", func() {
			dialogOpen := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					dialog.ShowError(err, u.window)
					return
				}
				if reader == nil {
					return
				}
				if err := u.loadDocument(reader); err != nil {
					dialog.ShowError(err, u.window)
					return
				}
			}, u.window)
			f := storage.NewExtensionFileFilter([]string{".json"})
			dialogOpen.SetFilter(f)
			dialogOpen.Show()
		}),
		recentItem,
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Documentation", func() {
			url, _ := url.Parse("https://developer.fyne.io")
			_ = u.app.OpenURL(url)
		}))

	main := fyne.NewMainMenu(u.fileMenu, helpMenu)
	return main
}

func (u *UI) loadDocument(reader fyne.URIReadCloser) error {
	defer reader.Close()
	infoText := binding.NewString()
	c := container.NewVBox(
		widget.NewLabelWithData(infoText),
		widget.NewProgressBarWithData(u.document.Progress),
	)
	d2 := dialog.NewCustomWithoutButtons("Loading", c, u.window)
	d2.Show()
	defer d2.Hide()
	infoText.Set("Loading file... Please Wait.")
	data, n, err := loadFile(reader)
	if err != nil {
		return err
	}
	infoText.Set("Parsing file... Please Wait.")
	if err := u.loadData(data, n); err != nil {
		return err
	}
	u.setTitle(reader.URI().Name())
	u.addRecentFile(reader.URI())
	return nil
}

func loadFile(reader fyne.URIReadCloser) (any, int, error) {
	dat, err := io.ReadAll(reader)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read file: %s", err)
	}
	var data any
	if err := json.Unmarshal(dat, &data); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal JSON: %s", err)
	}
	log.Printf("Read and unmarshaled JSON file")
	n := bytes.Count(dat, []byte{'\n'})
	log.Printf("File has %d LOC", n)
	return data, n, nil
}
