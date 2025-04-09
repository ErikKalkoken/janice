package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	kxdialog "github.com/ErikKalkoken/fyne-kx/dialog"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"

	"github.com/ErikKalkoken/janice/internal/jsondocument"
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

// searchBar represents the search bar frame in the UI.
type searchBar struct {
	widget.BaseWidget

	collapseAll  *ttwidget.Button
	scrollBottom *ttwidget.Button
	scrollTop    *ttwidget.Button
	searchButton *ttwidget.Button
	searchEntry  *widget.Entry
	searchType   *ttwidget.Select
	u            *UI
}

func newSearchBar(u *UI) *searchBar {
	w := &searchBar{
		searchEntry: widget.NewEntry(),
		u:           u,
	}
	w.ExtendBaseWidget(w)
	w.searchType = ttwidget.NewSelect(
		[]string{
			searchTypeKey,
			searchTypeKeyword,
			searchTypeNumber,
			searchTypeString,
		},
		nil,
	)
	w.searchType.SetSelected(searchTypeKey)
	w.searchType.SetToolTip("Select what to search")
	w.searchType.Disable()
	w.searchEntry.SetPlaceHolder(
		"Enter pattern to search for...")
	w.searchEntry.OnSubmitted = func(s string) {
		w.doSearch()
	}
	w.searchButton = ttwidget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		w.doSearch()
	})
	w.searchButton.SetToolTip("Search")
	w.scrollBottom = ttwidget.NewButtonWithIcon("", theme.NewThemedResource(resourceVerticalalignbottomSvg), func() {
		w.u.treeWidget.ScrollToBottom()
	})
	w.scrollBottom.SetToolTip("Scroll to bottom")
	w.scrollTop = ttwidget.NewButtonWithIcon("", theme.NewThemedResource(resourceVerticalaligntopSvg), func() {
		w.u.treeWidget.ScrollToTop()
	})
	w.scrollTop.SetToolTip("Scroll to top")
	w.collapseAll = ttwidget.NewButtonWithIcon("", theme.NewThemedResource(resourceUnfoldlessSvg), func() {
		w.u.treeWidget.CloseAllBranches()
	})
	w.collapseAll.SetToolTip("Collapse all")
	return w
}

func (w *searchBar) enable() {
	w.searchButton.Enable()
	w.searchType.Enable()
	w.searchEntry.Enable()
	w.scrollBottom.Enable()
	w.scrollTop.Enable()
	w.collapseAll.Enable()
}

func (w *searchBar) disable() {
	w.searchButton.Disable()
	w.searchType.Disable()
	w.searchEntry.Disable()
	w.scrollBottom.Disable()
	w.scrollTop.Disable()
	w.collapseAll.Disable()
}

func (w *searchBar) doSearch() {
	search := w.searchEntry.Text
	if len(search) == 0 {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	spinner := widget.NewActivity()
	spinner.Start()
	searchType := w.searchType.Selected
	c := container.NewHBox(widget.NewLabel(fmt.Sprintf("Searching for %s with pattern: %s", searchType, search)), spinner)
	b := widget.NewButton("Cancel", func() {
		cancel()
	})
	d := dialog.NewCustomWithoutButtons("Search", container.NewVBox(c, b), w.u.window)
	kxdialog.AddDialogKeyHandler(d, w.u.window)
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
				w.u.showErrorDialog("Allowed keywords are: true, false, null", nil)
				return
			}
		case searchTypeString:
			typ = jsondocument.SearchString
		case searchTypeNumber:
			typ = jsondocument.SearchNumber
		}
		uid, err := w.u.document.Search(ctx, w.u.selection.selectedUID, search, typ)
		d.Hide()
		if errors.Is(err, jsondocument.ErrCallerCanceled) {
			return
		} else if errors.Is(err, jsondocument.ErrNotFound) {
			d2 := dialog.NewInformation(
				"No match",
				fmt.Sprintf("No %s found matching %s", searchType, search),
				w.u.window,
			)
			kxdialog.AddDialogKeyHandler(d, w.u.window)
			d2.Show()
			return
		} else if err != nil {
			w.u.showErrorDialog("Search failed", err)
			return
		}
		w.u.scrollTo(uid)
	}()
}

func (w *searchBar) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewBorder(
		nil,
		nil,
		w.searchType,
		container.NewHBox(
			w.searchButton,
			container.NewPadded(),
			layout.NewSpacer(),
			w.scrollTop,
			w.scrollBottom,
			w.collapseAll,
		),
		w.searchEntry,
	)
	return widget.NewSimpleRenderer(c)
}
