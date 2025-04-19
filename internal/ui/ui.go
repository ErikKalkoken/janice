// Package ui contains the user interface.
package ui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	fynetooltip "github.com/dweymouth/fyne-tooltip"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	kxdialog "github.com/ErikKalkoken/fyne-kx/dialog"
	kxtheme "github.com/ErikKalkoken/fyne-kx/theme"
	kxwidget "github.com/ErikKalkoken/fyne-kx/widget"

	"github.com/ErikKalkoken/janice/internal/jsondocument"
)

const (
	colorThemeAuto         = "Automatic"
	colorThemeLight        = "Light"
	colorThemeDark         = "Dark"
	preferencesRecentFiles = "recent-files"
	websiteURL             = "https://github.com/ErikKalkoken/janice"
)

// preference keys
const (
	preferenceLastDetailShown    = "last-value-frame-shown"
	preferenceLastSelectionShown = "last-selection-frame-shown"
	preferenceLastWindowHeight   = "last-window-height"
	preferenceLastWindowWidth    = "last-window-width"
)

// setting keys and defaults
const (
	settingColorTheme             = "color-theme"
	settingExtensionDefault       = true
	settingExtensionFilter        = "extension-filter"
	settingNotifyUpdates          = "notify-updates"
	settingNotifyUpdatesDefault   = true
	settingRecentFileCount        = "recent-file-count"
	settingRecentFileCountDefault = 5
)

// UI represents the user interface of this app.
type UI struct {
	app                 fyne.App
	currentFile         fyne.URI
	detail              *detail
	document            *jsondocument.JSONDocument
	fileExportClipboard *fyne.MenuItem
	fileExportFile      *fyne.MenuItem
	fileNew             *fyne.MenuItem
	fileOpenRecent      *fyne.MenuItem
	fileReload          *fyne.MenuItem
	goBottom            *fyne.MenuItem
	goSelection         *fyne.MenuItem
	goTop               *fyne.MenuItem
	searchBar           *searchBar
	selection           *selection
	statusBar           *statusBar
	tree                *jsonTree
	viewCollapseAll     *fyne.MenuItem
	viewExpandAll       *fyne.MenuItem
	viewShowDetail      *fyne.MenuItem
	viewShowSelection   *fyne.MenuItem
	welcomeMessage      *fyne.Container
	window              fyne.Window
}

// NewUI returns a new UI object.
func NewUI(app fyne.App) (*UI, error) {
	appName := app.Metadata().Name
	u := &UI{
		app:      app,
		document: jsondocument.New(),
		window:   app.NewWindow(appName),
	}

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

	u.detail = newDetail(u)
	u.searchBar = newSearchBar(u)
	u.selection = newSelection(u)
	u.statusBar = newStatusBar(u)
	u.tree = newJSONTree(u)

	if u.app.Preferences().BoolWithFallback(preferenceLastSelectionShown, false) {
		u.selection.Show()
	} else {
		u.selection.Hide()
	}
	if u.app.Preferences().BoolWithFallback(preferenceLastDetailShown, false) {
		u.detail.Show()
	} else {
		u.detail.Hide()
	}

	c := container.NewBorder(
		container.NewVBox(u.searchBar, u.selection, u.detail, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), u.statusBar),
		nil,
		nil,
		container.NewStack(u.welcomeMessage, u.tree))

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
	u.window.SetOnClosed(func() {
		app.Preferences().SetFloat(preferenceLastWindowWidth, float64(u.window.Canvas().Size().Width))
		app.Preferences().SetFloat(preferenceLastWindowHeight, float64(u.window.Canvas().Size().Height))
		app.Preferences().SetBool(preferenceLastDetailShown, !u.detail.Hidden)
		app.Preferences().SetBool(preferenceLastSelectionShown, !u.selection.Hidden)
	})
	return u, nil
}

func (u *UI) selectElement(uid string) {
	u.selection.set(uid)
	u.selection.enable()
	u.detail.set(uid)
	u.fileExportFile.Disabled = false
	u.fileExportClipboard.Disabled = false
	u.window.MainMenu().Refresh()
}

// ShowAndRun shows the main window and runs the app. This method is blocking.
func (u *UI) ShowAndRun(path string) {
	u.app.Lifecycle().SetOnStarted(func() {
		u.setColorTheme(u.app.Preferences().StringWithFallback(settingColorTheme, colorThemeAuto))
		if path != "" {
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
		}
	})
	u.window.ShowAndRun()
}

func (u *UI) showErrorDialog(message string, err error) {
	if err != nil {
		slog.Error(message, "err", err)
	}
	d := dialog.NewInformation("Error", message, u.window)
	kxdialog.AddDialogKeyHandler(d, u.window)
	d.Show()
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
	kxdialog.AddDialogKeyHandler(d2, u.window)
	d2.Show()
	go func() {
		doc := jsondocument.New()
		if err := doc.Load(ctx, reader, progressInfo); err != nil {
			fyne.Do(func() {
				d2.Hide()
			})
			if errors.Is(err, jsondocument.ErrCallerCanceled) {
				return
			}
			fyne.Do(func() {
				u.showErrorDialog(fmt.Sprintf("Failed to open document: %s", reader.URI()), err)
			})
			return
		}
		fyne.Do(func() {
			u.document = doc
			u.statusBar.set(u.document.Size())
			u.welcomeMessage.Hide()
			u.toogleHasDocument(true)
			if doc.Size() > 1000 {
				u.viewExpandAll.Disabled = true
			} else {
				u.viewExpandAll.Disabled = false
			}
			u.window.MainMenu().Refresh()
			u.tree.Refresh()
			uri := reader.URI()
			if uri.Scheme() == "file" {
				u.addRecentFile(uri)
			}
			u.setTitle(uri.Name())
			u.currentFile = uri
			u.selection.reset()
			u.detail.reset()
			d2.Hide()
		})
	}()
}

func (u *UI) toogleHasDocument(enabled bool) {
	if enabled {
		u.searchBar.enable()
		u.fileExportClipboard.Disabled = false
		u.fileExportFile.Disabled = u.selection.selectedUID == ""
		u.fileNew.Disabled = false
		u.fileReload.Disabled = false
		u.goBottom.Disabled = false
		u.goSelection.Disabled = false
		u.goTop.Disabled = false
		u.viewCollapseAll.Disabled = false
		u.viewExpandAll.Disabled = false

	} else {
		u.searchBar.disable()
		u.fileExportClipboard.Disabled = true
		u.fileExportFile.Disabled = true
		u.fileNew.Disabled = true
		u.fileReload.Disabled = true
		u.goBottom.Disabled = true
		u.goSelection.Disabled = true
		u.goTop.Disabled = true
		u.viewCollapseAll.Disabled = true
		u.viewExpandAll.Disabled = true
	}
	u.window.MainMenu().Refresh()
}

func (u *UI) setColorTheme(s string) {
	switch s {
	case colorThemeLight:
		u.app.Settings().SetTheme(kxtheme.DefaultWithFixedVariant(theme.VariantLight))
	case colorThemeDark:
		u.app.Settings().SetTheme(kxtheme.DefaultWithFixedVariant(theme.VariantDark))
	default:
		u.app.Settings().SetTheme(theme.DefaultTheme())
	}
}

func (u *UI) showAboutDialog() {
	data := u.app.Metadata()
	current := data.Version
	x, err := url.Parse(data.Custom["Website"])
	if err != nil || x.Path == "" {
		x, _ = url.Parse(websiteURL)
	}
	c := container.NewVBox(
		widget.NewRichTextFromMarkdown(
			fmt.Sprintf("## %s\n\n"+
				"**Version:** %s\n\n"+
				"(c) 2024-2025 Erik Kalkoken", data.Name, current),
		),
		widget.NewLabel("A desktop app for viewing large JSON files."),
		widget.NewHyperlink("Website", x),
	)
	d := dialog.NewCustom("About", "OK", c, u.window)
	kxdialog.AddDialogKeyHandler(d, u.window)
	d.Show()
}

func (u *UI) showSettingsDialog() {
	// recent files
	recentEntry := kxwidget.NewSlider(3, 20)
	x := u.app.Preferences().IntWithFallback(settingRecentFileCount, settingRecentFileCountDefault)
	recentEntry.SetValue(float64(x))
	recentEntry.OnChangeEnded = func(v float64) {
		u.app.Preferences().SetInt(settingRecentFileCount, int(v))
	}

	// apply file filter
	extFilter := kxwidget.NewSwitch(func(v bool) {
		u.app.Preferences().SetBool(settingExtensionFilter, v)
	})
	y := u.app.Preferences().BoolWithFallback(settingExtensionFilter, settingExtensionDefault)
	extFilter.SetOn(y)

	notifyUpdates := kxwidget.NewSwitch(func(v bool) {
		u.app.Preferences().SetBool(settingNotifyUpdates, v)
	})
	z := u.app.Preferences().BoolWithFallback(settingNotifyUpdates, settingNotifyUpdatesDefault)
	notifyUpdates.SetOn(z)

	// theme
	theme := widget.NewRadioGroup([]string{colorThemeAuto, colorThemeLight, colorThemeDark}, func(s string) {
		u.setColorTheme(s)
		u.app.Preferences().SetString(settingColorTheme, s)
	})
	theme.Selected = u.app.Preferences().StringWithFallback(settingColorTheme, colorThemeAuto)
	items := []*widget.FormItem{
		{
			Text:   "Max recent files",
			Widget: recentEntry, HintText: "Maximum number of recent files remembered",
		},
		{
			Text: "JSON file filter", Widget: extFilter,
			HintText: "Wether to show files with .json extension only",
		},
		{
			Text:   "Notify about updates",
			Widget: notifyUpdates, HintText: "Wether to notify when an update is available (requires restart)",
		},
		{
			Text: "Appearance", Widget: theme,
			HintText: "Choose the color scheme. Automatic uses the current OS theme.",
		},
	}
	d := dialog.NewCustom("Settings", "Close", widget.NewForm(items...), u.window)
	kxdialog.AddDialogKeyHandler(d, u.window)
	d.Show()
}

func (u *UI) makeMenu() *fyne.MainMenu {
	// File menu
	u.fileNew = fyne.NewMenuItem("New", u.newFile)
	u.fileNew.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl}
	u.window.Canvas().AddShortcut(addShortcutFromMenuItem(u.fileNew))

	u.fileOpenRecent = fyne.NewMenuItem("Open Recent", nil)
	u.fileOpenRecent.ChildMenu = fyne.NewMenu("")

	fileSettingsItem := fyne.NewMenuItem("Settings...", u.showSettingsDialog)
	fileSettingsItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: fyne.KeyModifierControl}
	u.window.Canvas().AddShortcut(addShortcutFromMenuItem(fileSettingsItem))

	u.fileReload = fyne.NewMenuItem("Reload", u.reloadFile)
	u.fileReload.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: fyne.KeyModifierAlt}
	u.window.Canvas().AddShortcut(addShortcutFromMenuItem(u.fileReload))

	fileOpenItem := fyne.NewMenuItem("Open File...", u.openFile)
	fileOpenItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierControl}
	u.window.Canvas().AddShortcut(addShortcutFromMenuItem(fileOpenItem))

	u.fileExportFile = fyne.NewMenuItem("Export Selection To File...", func() {
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
		kxdialog.AddDialogKeyHandler(d, u.window)
		d.Show()
	})
	u.fileExportClipboard = fyne.NewMenuItem("Export Selection To Clipboard", func() {
		byt, err := u.extractSelection()
		if err != nil {
			u.showErrorDialog("Failed to extract selection", err)
		}
		u.app.Clipboard().SetContent(string(byt))
	})
	fileMenu := fyne.NewMenu("File",
		u.fileNew,
		fyne.NewMenuItemSeparator(),
		fileOpenItem,
		u.fileOpenRecent,
		fyne.NewMenuItem("Open From Clipboard", func() {
			r := strings.NewReader(u.app.Clipboard().Content())
			reader := jsondocument.MakeURIReadCloser(r, "CLIPBOARD")
			u.loadDocument(reader)
		}),
		u.fileReload,
		fyne.NewMenuItemSeparator(),
		u.fileExportFile,
		u.fileExportClipboard,
		fyne.NewMenuItemSeparator(),
		fileSettingsItem,
	)

	// View menu
	u.viewExpandAll = fyne.NewMenuItem("Expand All", func() {
		u.tree.OpenAllBranches()
	})
	u.viewCollapseAll = fyne.NewMenuItem("Collapse All", func() {
		u.tree.CloseAllBranches()
	})
	u.viewShowSelection = fyne.NewMenuItem("Show selected element", func() {
		u.toogleViewSelection()
	})
	u.viewShowSelection.Checked = !u.selection.Hidden
	u.viewShowDetail = fyne.NewMenuItem("Show value detail", func() {
		u.toogleViewDetail()
	})
	u.viewShowDetail.Checked = !u.detail.Hidden
	viewMenu := fyne.NewMenu("View",
		u.viewExpandAll,
		u.viewCollapseAll,
		fyne.NewMenuItemSeparator(),
		u.viewShowSelection,
		u.viewShowDetail,
	)

	// Go menu
	u.goTop = fyne.NewMenuItem("Go to top", u.tree.ScrollToTop)
	u.goTop.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyHome, Modifier: fyne.KeyModifierControl}
	u.window.Canvas().AddShortcut(addShortcutFromMenuItem(u.goTop))

	u.goBottom = fyne.NewMenuItem("Go to bottom", u.tree.ScrollToBottom)
	u.goBottom.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyEnd, Modifier: fyne.KeyModifierControl}
	u.window.Canvas().AddShortcut(addShortcutFromMenuItem(u.goBottom))

	u.goSelection = fyne.NewMenuItem("Go to selection", func() {
		u.tree.scrollTo(u.selection.selectedUID)
	})
	goMenu := fyne.NewMenu("Go",
		u.goTop,
		u.goBottom,
		u.goSelection,
	)

	// Help menu
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

	main := fyne.NewMainMenu(fileMenu, viewMenu, goMenu, helpMenu)
	return main
}

func (u *UI) openFile() {
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
	kxdialog.AddDialogKeyHandler(d, u.window)
	d.Show()
	filterEnabled := u.app.Preferences().BoolWithFallback(settingExtensionFilter, settingExtensionDefault)
	if filterEnabled {
		f := storage.NewExtensionFileFilter([]string{".json"})
		d.SetFilter(f)
	}
}

// newFile resets the app to it's initial state
func (u *UI) newFile() {
	u.document.Reset()
	u.setTitle("")
	u.statusBar.reset()
	u.welcomeMessage.Show()
	u.toogleHasDocument(false)
	u.selection.reset()
	u.detail.reset()
}

func (u *UI) reloadFile() {
	if u.currentFile == nil {
		return
	}
	reader, err := storage.Reader(u.currentFile)
	if err != nil {
		u.showErrorDialog("Failed to reload file", err)
		return
	}
	u.loadDocument(reader)
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
	files := u.app.Preferences().StringList(preferencesRecentFiles)
	if len(files) == 0 {
		u.fileOpenRecent.Disabled = true
	} else {
		u.fileOpenRecent.Disabled = false
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
		u.fileOpenRecent.ChildMenu.Items = items
	}
	u.window.MainMenu().Refresh()
}

func (u *UI) toogleViewSelection() {
	if u.selection.Hidden {
		u.selection.Show()
	} else {
		u.selection.Hide()
	}
	u.viewShowSelection.Checked = !u.selection.Hidden
	u.window.MainMenu().Refresh()
}

func (u *UI) toogleViewDetail() {
	if u.detail.Hidden {
		u.detail.Show()
	} else {
		u.detail.Hide()
	}
	u.viewShowDetail.Checked = !u.detail.Hidden
	u.window.MainMenu().Refresh()
}

// addShortcutFromMenuItem is a helper for defining shortcuts.
// It allows to add an already defined shortcut from a menu item to the canvas.
//
// For example:
//
//	window.Canvas().AddShortcut(menuItem)
func addShortcutFromMenuItem(item *fyne.MenuItem) (fyne.Shortcut, func(fyne.Shortcut)) {
	return item.Shortcut, func(s fyne.Shortcut) {
		item.Action()
	}
}
