package ui

import (
	"fmt"
	"log/slog"
	"net/url"
	"slices"

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

func (u *UI) loadDocument(reader fyne.URIReadCloser) error {
	defer reader.Close()
	text := widget.NewLabel("")
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
		step := fmt.Sprintf("%d / %d", info.CurrentStep, info.TotalSteps)
		var t string
		switch info.CurrentStep {
		case 1:
			t = fmt.Sprintf("%s: Loading file from disk...", step)
		case 2:
			t = fmt.Sprintf("%s: Parsing file...", step)
		case 3:
			t = fmt.Sprintf("%s: Calculating size...", step)
		case 4:
			if pb2.Hidden {
				pb1.Stop()
				pb1.Hide()
				pb2.Show()
			}
			p := message.NewPrinter(language.English)
			t = p.Sprintf("%s: Rendering document for %d elements...", step, info.Size)
		}
		text.SetText(t)
		pb2.SetValue(info.Progress)
	}))
	c := container.NewVBox(text, container.NewStack(pb1, pb2))
	d2 := dialog.NewCustomWithoutButtons("Loading", c, u.window)
	d2.Show()
	defer d2.Hide()
	if err := u.document.Load(reader, progressInfo); err != nil {
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
