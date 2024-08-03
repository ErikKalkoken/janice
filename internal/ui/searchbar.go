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

// searchBarFrame represents the search bar frame in the UI.
type searchBarFrame struct {
	content *fyne.Container
	ui      *UI

	searchEntry  *widget.Entry
	searchButton *ttwidget.Button
	searchType   *ttwidget.Select
	scrollBottom *ttwidget.Button
	scrollTop    *ttwidget.Button
	collapseAll  *ttwidget.Button
}

func (u *UI) newSearchBarFrame() *searchBarFrame {
	f := &searchBarFrame{
		ui:          u,
		searchEntry: widget.NewEntry(),
	}
	// search frame
	f.searchType = ttwidget.NewSelect([]string{
		searchTypeKey,
		searchTypeKeyword,
		searchTypeNumber,
		searchTypeString,
	}, nil)
	f.searchType.SetSelected(searchTypeKey)
	f.searchType.SetToolTip("Select what to search")
	f.searchType.Disable()
	f.searchEntry.SetPlaceHolder(
		"Enter pattern to search for...")
	f.searchEntry.OnSubmitted = func(s string) {
		f.doSearch()
	}
	f.searchButton = ttwidget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		f.doSearch()
	})
	f.searchButton.SetToolTip("Search")
	f.scrollBottom = ttwidget.NewButtonWithIcon("", theme.NewThemedResource(resourceVerticalalignbottomSvg), func() {
		f.ui.treeWidget.ScrollToBottom()
	})
	f.scrollBottom.SetToolTip("Scroll to bottom")
	f.scrollTop = ttwidget.NewButtonWithIcon("", theme.NewThemedResource(resourceVerticalaligntopSvg), func() {
		f.ui.treeWidget.ScrollToTop()
	})
	f.scrollTop.SetToolTip("Scroll to top")
	f.collapseAll = ttwidget.NewButtonWithIcon("", theme.NewThemedResource(resourceUnfoldlessSvg), func() {
		f.ui.treeWidget.CloseAllBranches()
	})
	f.collapseAll.SetToolTip("Collapse all")
	c := container.NewBorder(
		nil,
		nil,
		f.searchType,
		container.NewHBox(
			f.searchButton,
			container.NewPadded(),
			layout.NewSpacer(),
			f.scrollTop,
			f.scrollBottom,
			f.collapseAll,
		),
		f.searchEntry,
	)
	f.content = c
	return f
}

func (f *searchBarFrame) enable() {
	f.searchButton.Enable()
	f.searchType.Enable()
	f.searchEntry.Enable()
	f.scrollBottom.Enable()
	f.scrollTop.Enable()
	f.collapseAll.Enable()
}

func (f *searchBarFrame) disable() {
	f.searchButton.Disable()
	f.searchType.Disable()
	f.searchEntry.Disable()
	f.scrollBottom.Disable()
	f.scrollTop.Disable()
	f.collapseAll.Disable()
}

func (f *searchBarFrame) doSearch() {
	search := f.searchEntry.Text
	if len(search) == 0 {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	spinner := widget.NewActivity()
	spinner.Start()
	searchType := f.searchType.Selected
	c := container.NewHBox(widget.NewLabel(fmt.Sprintf("Searching for %s with pattern: %s", searchType, search)), spinner)
	b := widget.NewButton("Cancel", func() {
		cancel()
	})
	d := dialog.NewCustomWithoutButtons("Search", container.NewVBox(c, b), f.ui.window)
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
				f.ui.showErrorDialog("Allowed keywords are: true, false, null", nil)
				return
			}
		case searchTypeString:
			typ = jsondocument.SearchString
		case searchTypeNumber:
			typ = jsondocument.SearchNumber
		}
		uid, err := f.ui.document.Search(ctx, f.ui.selection.selectedUID, search, typ)
		d.Hide()
		if errors.Is(err, jsondocument.ErrCallerCanceled) {
			return
		} else if errors.Is(err, jsondocument.ErrNotFound) {
			d2 := dialog.NewInformation(
				"No match",
				fmt.Sprintf("No %s found matching %s", searchType, search),
				f.ui.window,
			)
			d2.Show()
			return
		} else if err != nil {
			f.ui.showErrorDialog("Search failed", err)
			return
		}
		f.ui.scrollTo(uid)
	}()
}
