// Package ui contains the user interface.
package ui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/ErikKalkoken/jsonviewer/internal/github"
	"github.com/ErikKalkoken/jsonviewer/internal/jsondocument"
	"github.com/ErikKalkoken/jsonviewer/internal/widgets"
)

const (
	appTitle    = "JSON Viewer"
	githubOwner = "ErikKalkoken"
	githubRepo  = "jsonviewer"
	websiteURL  = "https://github.com/ErikKalkoken/jsonviewer"
)

// preference keys
const (
	preferenceLastSelectionFrameHidden = "last-selection-frame-hidden"
	preferenceLastValueFrameHidden     = "last-value-frame-hidden"
	preferenceLastWindowHeight         = "last-window-height"
	preferenceLastWindowWidth          = "last-window-width"
)

const (
	searchTypeKey     = "key"
	searchTypeString  = "string"
	searchTypeNumber  = "number"
	searchTypeKeyword = "keyword"
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
	document           *jsondocument.JSONDocument
	fileMenu           *fyne.Menu
	viewMenu           *fyne.Menu
	treeWidget         *widget.Tree
	welcomeMessage     *fyne.Container
	window             fyne.Window

	searchEntry  *widget.Entry
	searchButton *widget.Button
	searchType   *widget.Select
	scrollBottom *widget.Button
	scrollTop    *widget.Button
	collapseAll  *widget.Button

	selectionFrame   *fyne.Container
	selectedPath     *fyne.Container
	jumpToSelection  *widget.Button
	copyKeyClipboard *widget.Button

	detailFrame        *fyne.Container
	copyValueClipboard *widget.Button
	valueDisplay       *widget.RichText
	valueRaw           string

	statusTreeSize *widget.Label
}

// NewUI returns a new UI object.
func NewUI(app fyne.App) (*UI, error) {
	myHBox := layout.NewCustomPaddedHBoxLayout(-5)
	u := &UI{
		app:            app,
		document:       jsondocument.New(),
		selectedPath:   container.New(myHBox),
		valueDisplay:   widget.NewRichText(),
		statusTreeSize: widget.NewLabel(""),
		searchEntry:    widget.NewEntry(),
		window:         app.NewWindow(appTitle),
	}
	u.treeWidget = u.makeTree()

	// search frame
	u.searchType = widget.NewSelect([]string{
		searchTypeKey,
		searchTypeKeyword,
		searchTypeNumber,
		searchTypeString,
	}, nil)
	u.searchType.SetSelected(searchTypeKey)
	u.searchType.Disable()
	u.searchEntry.SetPlaceHolder(
		"Enter pattern to search for...")
	u.searchEntry.OnSubmitted = func(s string) {
		u.doSearch()
	}
	u.searchButton = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		u.doSearch()
	})
	u.scrollBottom = widget.NewButtonWithIcon("", theme.NewThemedResource(resourceVerticalalignbottomSvg), func() {
		u.treeWidget.ScrollToBottom()
	})
	u.scrollTop = widget.NewButtonWithIcon("", theme.NewThemedResource(resourceVerticalaligntopSvg), func() {
		u.treeWidget.ScrollToTop()
	})
	u.collapseAll = widget.NewButtonWithIcon("", theme.NewThemedResource(resourceUnfoldlessSvg), func() {
		u.treeWidget.CloseAllBranches()
	})
	searchBar := container.NewBorder(
		nil,
		nil,
		u.searchType,
		container.NewHBox(
			u.searchButton,
			container.NewPadded(),
			layout.NewSpacer(),
			u.scrollTop,
			u.scrollBottom,
			u.collapseAll,
		),
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

	// selection frame
	u.jumpToSelection = widget.NewButtonWithIcon("", theme.NewThemedResource(resourceReadmoreSvg), func() {
		u.scrollTo(u.currentSelectedUID)
	})
	u.jumpToSelection.Disable()
	u.copyKeyClipboard = widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		n := u.document.Value(u.currentSelectedUID)
		u.window.Clipboard().SetContent(n.Key)
	})
	u.copyKeyClipboard.Disable()
	u.selectionFrame = container.NewBorder(
		nil,
		nil,
		nil,
		container.NewHBox(u.jumpToSelection, u.copyKeyClipboard),
		container.NewHScroll(u.selectedPath),
	)
	u.selectionFrame.Hidden = app.Preferences().BoolWithFallback(preferenceLastSelectionFrameHidden, false)

	// value frame
	u.copyValueClipboard = widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		u.window.Clipboard().SetContent(u.valueRaw)
	})
	u.copyValueClipboard.Disable()
	u.detailFrame = container.NewBorder(
		nil,
		nil,
		nil,
		u.copyValueClipboard,
		container.NewScroll(u.valueDisplay),
	)
	u.detailFrame.Hidden = app.Preferences().BoolWithFallback(preferenceLastValueFrameHidden, false)

	// status bar frame
	statusBar := container.NewHBox(u.statusTreeSize)
	notifyUpdates := u.app.Preferences().BoolWithFallback(settingNotifyUpdates, settingNotifyUpdatesDefault)
	if notifyUpdates {
		go func() {
			current := u.app.Metadata().Version
			latest, isNewer, err := github.AvailableUpdate(githubOwner, githubRepo, current)
			if err != nil {
				slog.Error("Failed to fetch latest version from github", "err", err)
				return
			}
			if !isNewer {
				return
			}
			statusBar.Add(layout.NewSpacer())
			l := widgets.NewTappableLabel("Update available", func() {
				u.showReleaseDialog(current, latest)
			})
			l.Importance = widget.HighImportance
			statusBar.Add(l)
		}()
	}

	c := container.NewBorder(
		container.NewVBox(searchBar, u.selectionFrame, u.detailFrame, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), statusBar),
		nil,
		nil,
		container.NewStack(u.welcomeMessage, u.treeWidget))

	u.window.SetContent(c)
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
		app.Preferences().SetBool(preferenceLastValueFrameHidden, u.detailFrame.Hidden)
		app.Preferences().SetBool(preferenceLastSelectionFrameHidden, u.selectionFrame.Hidden)
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
	u.currentSelectedUID = uid
	u.renderSelectedPath(uid)
	node := u.document.Value(uid)
	u.jumpToSelection.Enable()
	u.copyKeyClipboard.Enable()
	typeText := fmt.Sprint(node.Type)
	var v string
	if u.document.IsBranch(uid) {
		u.copyValueClipboard.Disable()
		switch node.Type {
		case jsondocument.Array:
			v = "[...]"
		case jsondocument.Object:
			v = "{...}"
		}
		ids := u.document.ChildUIDs(uid)
		typeText += fmt.Sprintf(", %d elements", len(ids))
	} else {
		u.copyValueClipboard.Enable()
		switch node.Type {
		case jsondocument.String:
			x := node.Value.(string)
			v = fmt.Sprintf("\"%s\"", x)
			u.valueRaw = x
		case jsondocument.Number:
			x := node.Value.(float64)
			v = strconv.FormatFloat(x, 'f', -1, 64)
			u.valueRaw = v
		case jsondocument.Null:
			v = "null"
			u.valueRaw = v
		default:
			v = fmt.Sprint(node.Value)
			u.valueRaw = v
		}
	}
	u.valueDisplay.ParseMarkdown(fmt.Sprintf("```\n%s\n```", v))
	u.fileMenu.Items[7].Disabled = false
	u.fileMenu.Items[8].Disabled = false
	u.fileMenu.Refresh()
}

func (u *UI) renderSelectedPath(uid string) {
	p := u.document.Path(uid)
	var path []jsondocument.Node
	for _, id := range p {
		path = append(path, u.document.Value(id))
	}
	path = append(path, u.document.Value(uid))
	u.selectedPath.RemoveAll()
	for i, n := range path {
		isLast := i == len(path)-1
		if !isLast {
			l := widgets.NewTappableLabel(n.Key, func() {
				u.scrollTo(n.UID)
				u.selectElement(n.UID)
			})
			u.selectedPath.Add(l)
		} else {
			l := widget.NewLabel(n.Key)
			l.TextStyle.Bold = true
			u.selectedPath.Add(l)
		}
		if !isLast {
			l := widget.NewLabel("＞")
			l.Importance = widget.LowImportance
			u.selectedPath.Add(l)
		}
	}
}

func (u *UI) doSearch() {
	search := u.searchEntry.Text
	if len(search) == 0 {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	spinner := widget.NewActivity()
	spinner.Start()
	searchType := u.searchType.Selected
	c := container.NewHBox(widget.NewLabel(fmt.Sprintf("Searching for %s with pattern: %s", searchType, search)), spinner)
	b := widget.NewButton("Cancel", func() {
		cancel()
	})
	d := dialog.NewCustomWithoutButtons("Search", container.NewVBox(c, b), u.window)
	d.Show()
	d.SetOnClosed(func() {
		cancel()
	})
	go func() {
		var typ jsondocument.SearchType
		switch searchType {
		case searchTypeKey:
			typ = jsondocument.SearchKey
		case searchTypeKeyword:
			typ = jsondocument.SearchKeyword
			search = strings.ToLower(search)
			if search != "true" && search != "false" && search != "null" {
				d.Hide()
				u.showErrorDialog("Allowed keywords are: true, false, null", nil)
				return
			}
		case searchTypeString:
			typ = jsondocument.SearchString
		case searchTypeNumber:
			typ = jsondocument.SearchNumber
		}
		uid, err := u.document.Search(ctx, u.currentSelectedUID, search, typ)
		d.Hide()
		if errors.Is(err, jsondocument.ErrCallerCanceled) {
			return
		} else if errors.Is(err, jsondocument.ErrNotFound) {
			d2 := dialog.NewInformation("No match", fmt.Sprintf("No %s found matching %s", searchType, search), u.window)
			d2.Show()
			return
		} else if err != nil {
			u.showErrorDialog("Search failed", err)
			return
		}
		u.scrollTo(uid)
	}()
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
	u.statusTreeSize.SetText("")
	u.welcomeMessage.Show()
	u.toogleHasDocument(false)
	u.selectedPath.RemoveAll()
	u.valueDisplay.ParseMarkdown("")
	u.copyValueClipboard.Disable()
	u.jumpToSelection.Disable()
	u.copyKeyClipboard.Disable()
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
		u.statusTreeSize.SetText(p.Sprintf("%d elements", u.document.Size()))
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
		d2.Hide()
	}()
}

func (u *UI) toogleHasDocument(enabled bool) {
	if enabled {
		u.searchButton.Enable()
		u.searchType.Enable()
		u.searchEntry.Enable()
		u.scrollBottom.Enable()
		u.scrollTop.Enable()
		u.collapseAll.Enable()
		u.fileMenu.Items[0].Disabled = false
		u.fileMenu.Items[5].Disabled = false
		u.fileMenu.Items[7].Disabled = u.currentSelectedUID == ""
		u.fileMenu.Items[8].Disabled = u.currentSelectedUID == ""
		for _, o := range u.viewMenu.Items {
			o.Disabled = false
		}
	} else {
		u.searchButton.Disable()
		u.searchType.Disable()
		u.searchEntry.Disable()
		u.scrollBottom.Disable()
		u.scrollTop.Disable()
		u.collapseAll.Disable()
		u.fileMenu.Items[0].Disabled = true
		u.fileMenu.Items[5].Disabled = true
		u.fileMenu.Items[7].Disabled = true
		u.fileMenu.Items[8].Disabled = true
		for _, o := range u.viewMenu.Items {
			o.Disabled = true
		}
	}
	u.fileMenu.Refresh()
	u.viewMenu.Refresh()
}

func (u *UI) showReleaseDialog(current, latest string) {
	x := websiteURL + "/releases"
	y, _ := url.Parse(x)
	link := widget.NewHyperlink(x, y)
	c := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("New release %s available. You have release %s.", latest, current)),
		widget.NewLabel("The link below will bring you to the releases page,\n"+
			"where you can download the latest version:"),
		link,
	)
	d := dialog.NewCustom("Update available", "OK", c, u.window)
	d.Show()
}
