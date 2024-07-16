package ui

import (
	"context"
	"errors"
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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ErikKalkoken/jsonviewer/internal/jsondocument"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	preferencesRecentFiles         = "recent-files"
	settingsRecentFileCount        = "settings-recent-files"
	settingsRecentFileCountDefault = 5
	settingsExtensionFilter        = "settings-extension-filter"
	settingsExtensionDefault       = true
	settingsNotifyUpdates          = "settings-notify-updates"
	settingsNotifyUpdatesDefault   = true
)

func (u *UI) makeMenu() *fyne.MainMenu {
	recentItem := fyne.NewMenuItem("Open Recent", nil)
	recentItem.ChildMenu = fyne.NewMenu("")
	u.fileMenu = fyne.NewMenu("File",
		fyne.NewMenuItem("New", func() {
			u.reset()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Open File...", func() {
			d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					u.showErrorDialog("Failed to read folder", err)
					return
				}
				if reader == nil {
					return
				}
				u.loadDocument(reader)
			}, u.window)
			d.Show()
			filterEnabled := u.app.Preferences().BoolWithFallback(settingsExtensionFilter, settingsExtensionDefault)
			if filterEnabled {
				f := storage.NewExtensionFileFilter([]string{".json"})
				d.SetFilter(f)
			}
		}),
		recentItem,
		fyne.NewMenuItem("Open From Clipboard", func() {
			r := strings.NewReader(u.window.Clipboard().Content())
			reader := jsondocument.MakeURIReadCloser(r, "CLIPBOARD")
			u.loadDocument(reader)
		}),
		fyne.NewMenuItem("Reload", func() {
			if u.currentFile == nil {
				return
			}
			reader, err := storage.Reader(u.currentFile)
			if err != nil {
				u.showErrorDialog("Failed to reload file", err)
				return
			}
			u.loadDocument(reader)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Export Selection To File...", func() {
			byt, err := u.extractSelection()
			if err != nil {
				u.showErrorDialog("Failed to extract selection", err)
			}
			d := dialog.NewFileSave(func(f fyne.URIWriteCloser, err error) {
				if err != nil {
					u.showErrorDialog("Failed to open save dialog", err)
					return
				}
				if f == nil {
					return
				}
				_, err = f.Write(byt)
				if err != nil {
					u.showErrorDialog("Failed to write file", err)
					return
				}
			}, u.window)
			d.Show()
		}),
		fyne.NewMenuItem("Export Selection To Clipboard", func() {
			byt, err := u.extractSelection()
			if err != nil {
				u.showErrorDialog("Failed to extract selection", err)
			}
			u.window.Clipboard().SetContent(string(byt))
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Preferences...", func() {
			u.showSettingsDialog()
		}),
	)
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Scroll to top", func() {
			u.treeWidget.ScrollToTop()
		}),
		fyne.NewMenuItem("Scroll to bottom", func() {
			u.treeWidget.ScrollToBottom()
		}),
		fyne.NewMenuItem("Scroll to selection", func() {
			u.showInTree(u.currentSelectedUID)
		}),
		fyne.NewMenuItemSeparator(),
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

func (u *UI) extractSelection() ([]byte, error) {
	uid := u.currentSelectedUID
	n := u.document.Value(uid)
	if n.Type != jsondocument.Array && n.Type != jsondocument.Object {
		uid = u.document.Parent(uid)
	}
	byt, err := u.document.Extract(uid)
	if err != nil {
		return nil, err
	}
	return byt, nil
}

func (u *UI) addRecentFile(uri fyne.URI) {
	files := u.app.Preferences().StringList(preferencesRecentFiles)
	uri2 := uri.String()
	max := u.app.Preferences().IntWithFallback(settingsRecentFileCount, settingsRecentFileCountDefault)
	files = addToListWithRotation(files, uri2, max)
	u.app.Preferences().SetStringList(preferencesRecentFiles, files)
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
	files := u.app.Preferences().StringList(preferencesRecentFiles)
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
			u.loadDocument(reader)
		})
	}
	u.fileMenu.Items[3].ChildMenu.Items = items
	u.fileMenu.Refresh()
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
			text = fmt.Sprintf("Parsing file: %s", name)
		case 3:
			text = fmt.Sprintf("Calculating document size: %s", name)
		case 4:
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
	c := container.NewVBox(infoText, container.NewBorder(nil, nil, nil, b, container.NewStack(pb1, pb2)))
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
		p := message.NewPrinter(language.English)
		out := p.Sprintf("%d elements", u.document.Size())
		u.statusTreeSize.SetText(out)
		u.welcomeMessage.Hide()
		u.searchButton.Enable()
		u.searchType.Enable()
		u.searchEntry.Enable()
		u.collapseButton.Enable()
		u.treeWidget.Refresh()
		uri := reader.URI()
		if uri.Scheme() == "file" {
			u.addRecentFile(uri)
		}
		u.setTitle(uri.Name())
		u.currentFile = uri
		d2.Hide()
	}()
}
