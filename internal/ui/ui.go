// Package ui contains the user interface.
package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"

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

const (
	appTitle     = "JSON Viewer"
	windowWidth  = "main-window-width"
	windowHeight = "main-window-height"
)

// UI represents the user interface of this app.
type UI struct {
	app        fyne.App
	treeWidget *widget.Tree
	document   jsondocument.JSONDocument
	window     fyne.Window
	statusbar  *widget.Label
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
	u.window.SetMaster()
	s := fyne.Size{
		Width:  float32(a.Preferences().FloatWithFallback(windowWidth, 800)),
		Height: float32(a.Preferences().FloatWithFallback(windowHeight, 600)),
	}
	u.window.Resize(s)

	u.window.SetOnClosed(func() {
		a.Preferences().SetFloat(windowWidth, float64(u.window.Canvas().Size().Width))
		a.Preferences().SetFloat(windowHeight, float64(u.window.Canvas().Size().Height))
	})
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
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open File...", func() {
			d1 := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					dialog.ShowError(err, u.window)
					return
				}
				if reader == nil {
					return
				}
				defer reader.Close()
				infoText := binding.NewString()
				c := container.NewVBox(
					widget.NewLabelWithData(infoText),
					widget.NewProgressBarWithData(u.document.Progress),
				)
				d2 := dialog.NewCustomWithoutButtons("Loading", c, u.window)
				d2.Show()
				infoText.Set("Loading file... Please Wait.")
				data, n := loadFile(reader)
				infoText.Set("Parsing file... Please Wait.")
				u.setData(data, n)
				d2.Hide()
				u.setTitle(reader.URI().Name())
			}, u.window)
			f := storage.NewExtensionFileFilter([]string{".json"})
			d1.SetFilter(f)
			d1.Show()
		}))
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Documentation", func() {
			url, _ := url.Parse("https://developer.fyne.io")
			_ = u.app.OpenURL(url)
		}))

	main := fyne.NewMainMenu(fileMenu, helpMenu)
	return main
}

func loadFile(reader fyne.URIReadCloser) (any, int) {
	dat, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}
	var data any
	if err := json.Unmarshal(dat, &data); err != nil {
		log.Fatalf("failed to unmarshal JSON: %s", err)
	}
	log.Printf("Read and unmarshaled JSON file")
	n := bytes.Count(dat, []byte{'\n'})
	log.Printf("File has %d LOC", n)
	return data, n
}
