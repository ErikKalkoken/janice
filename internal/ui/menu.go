package ui

import (
	"log/slog"
	"net/url"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"github.com/ErikKalkoken/janice/internal/jsondocument"
)

const (
	preferencesRecentFiles = "recent-files"
	websiteURL             = "https://github.com/ErikKalkoken/janice"
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
			filterEnabled := u.app.Preferences().BoolWithFallback(settingExtensionFilter, settingExtensionDefault)
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
				defer f.Close()
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
		fyne.NewMenuItem("Settings...", func() {
			u.showSettingsDialog()
		}),
	)
	toogleSelectionFrame := fyne.NewMenuItem("Show selected element", func() {
		u.toogleViewSelection()
	})
	toogleSelectionFrame.Checked = u.selection.isShown()
	toogleDetailFrame := fyne.NewMenuItem("Show value detail", func() {
		u.toogleViewDetail()
	})
	toogleDetailFrame.Checked = u.value.isShown()
	u.viewMenu = fyne.NewMenu("View",
		fyne.NewMenuItem("Scroll to top", func() {
			u.treeWidget.ScrollToTop()
		}),
		fyne.NewMenuItem("Scroll to bottom", func() {
			u.treeWidget.ScrollToBottom()
		}),
		fyne.NewMenuItem("Scroll to selection", func() {
			u.scrollTo(u.selection.selectedUID)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Expand All", func() {
			u.treeWidget.OpenAllBranches()
		}),
		fyne.NewMenuItem("Collapse All", func() {
			u.treeWidget.CloseAllBranches()
		}),
		fyne.NewMenuItemSeparator(),
		toogleSelectionFrame,
		toogleDetailFrame,
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Report a bug", func() {
			url, _ := url.Parse(websiteURL + "/issues")
			_ = u.app.OpenURL(url)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("About...", func() {
			u.showAboutDialog()
		}),
	)
	main := fyne.NewMainMenu(u.fileMenu, u.viewMenu, helpMenu)
	return main
}

func (u *UI) extractSelection() ([]byte, error) {
	uid := u.selection.selectedUID
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
	max := u.app.Preferences().IntWithFallback(settingRecentFileCount, settingRecentFileCountDefault)
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
	recentFiles := u.fileMenu.Items[3]
	files := u.app.Preferences().StringList(preferencesRecentFiles)
	if len(files) == 0 {
		recentFiles.Disabled = true
	} else {
		recentFiles.Disabled = false
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
		recentFiles.ChildMenu.Items = items
	}
	u.fileMenu.Refresh()
}

func (u *UI) toogleViewSelection() {
	if u.selection.isShown() {
		u.selection.hide()
	} else {
		u.selection.show()
	}
	menuItem := u.viewMenu.Items[7]
	menuItem.Checked = u.selection.isShown()
	u.viewMenu.Refresh()
}

func (u *UI) toogleViewDetail() {
	if u.value.isShown() {
		u.value.hide()
	} else {
		u.value.show()
	}
	menuItem := u.viewMenu.Items[8]
	menuItem.Checked = u.value.isShown()
	u.viewMenu.Refresh()
}
