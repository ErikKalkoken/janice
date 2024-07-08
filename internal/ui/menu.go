package ui

import (
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/ErikKalkoken/jsonviewer/internal/jsondocument"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

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
			dialogOpen.Show()
		}),
		recentItem,
		fyne.NewMenuItem("Open From Clipboard", func() {
			r := strings.NewReader(u.window.Clipboard().Content())
			reader := makeURIReadCloser(r)
			if err := u.loadDocument(reader); err != nil {
				u.showErrorDialog("Failed to load JSON document from clipboard", err)
				return
			}
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Reload", func() {
			if u.currentFile == nil {
				return
			}
			if err := u.loadURI(u.currentFile); err != nil {
				u.showErrorDialog("Failed to reload file", err)
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
			url, _ := url.Parse("https://github.com/ErikKalkoken/jsonviewer/issues")
			_ = u.app.OpenURL(url)
		}),
		fyne.NewMenuItem("Website", func() {
			url, _ := url.Parse("https://github.com/ErikKalkoken/jsonviewer")
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

func (u *UI) loadURI(uri fyne.URI) error {
	reader, err := storage.Reader(uri)
	if err != nil {
		return err
	}
	if err := u.loadDocument(reader); err != nil {
		return err
	}
	return nil
}

func (u *UI) loadDocument(reader fyne.URIReadCloser) error {
	defer reader.Close()
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
		var text string
		switch info.CurrentStep {
		case 1:
			text = "Loading file from disk..."
		case 2:
			text = "Parsing file..."
		case 3:
			text = "Calculating size..."
		case 4:
			if pb2.Hidden {
				pb1.Stop()
				pb1.Hide()
				pb2.Show()
			}
			p := message.NewPrinter(language.English)
			text = p.Sprintf("Rendering document with %d elements...", info.Size)
			pb2.SetValue(info.Progress)
		default:
			text = "?"
		}
		message := fmt.Sprintf("%d / %d: %s", info.CurrentStep, info.TotalSteps, text)
		infoText.SetText(message)
	}))
	c := container.NewVBox(infoText, container.NewStack(pb1, pb2))
	d2 := dialog.NewCustomWithoutButtons("Loading", c, u.window)
	d2.Show()
	defer d2.Hide()
	if err := u.document.Load(reader, progressInfo); err != nil {
		return err
	}
	p := message.NewPrinter(language.English)
	out := p.Sprintf("%d elements", u.document.Size())
	u.statusTreeSize.SetText(out)
	u.welcomeMessage.Hide()
	u.treeWidget.Refresh()
	u.detailCopyValue.Show()
	x := reader.URI()
	if x != nil {
		u.setTitle(x.Name())
		u.addRecentFile(x)
	} else {
		u.setTitle("CLIPBOARD")
	}
	u.currentFile = x
	return nil
}

func (u *UI) showAboutDialog() {
	c := container.NewVBox()
	info := u.app.Metadata()
	name := u.appName()
	appData := widget.NewRichTextFromMarkdown(fmt.Sprintf(
		"## %s\n**Version:** %s", name, info.Version))
	c.Add(appData)
	c.Add(widget.NewLabel("(c) 2024 Erik Kalkoken"))
	d := dialog.NewCustom("About", "OK", c, u.window)
	d.Show()
}
