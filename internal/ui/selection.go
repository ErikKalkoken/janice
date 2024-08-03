package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"

	"github.com/ErikKalkoken/janice/internal/jsondocument"
	"github.com/ErikKalkoken/janice/internal/widgets"
)

// selection represents the selection frame in the UI.
type selectionFrame struct {
	content *fyne.Container
	ui      *UI

	selectedUID      widget.TreeNodeID
	selectedPath     *fyne.Container
	jumpToSelection  *ttwidget.Button
	copyKeyClipboard *ttwidget.Button
}

func (u *UI) newSelectionFrame() *selectionFrame {
	myHBox := layout.NewCustomPaddedHBoxLayout(-5)

	f := &selectionFrame{
		ui:           u,
		selectedPath: container.New(myHBox),
	}
	f.jumpToSelection = ttwidget.NewButtonWithIcon("", theme.NewThemedResource(resourceReadmoreSvg), func() {
		u.scrollTo(f.selectedUID)
	})
	f.jumpToSelection.SetToolTip("Jump to selection")
	f.jumpToSelection.Disable()
	f.copyKeyClipboard = ttwidget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		n := u.document.Value(f.selectedUID)
		u.window.Clipboard().SetContent(n.Key)
	})
	f.copyKeyClipboard.SetToolTip("Copy key to clipboard")
	f.copyKeyClipboard.Disable()
	c := container.NewBorder(
		nil,
		nil,
		nil,
		container.NewHBox(f.jumpToSelection, f.copyKeyClipboard),
		container.NewHScroll(f.selectedPath),
	)
	c.Hidden = !u.app.Preferences().BoolWithFallback(preferenceLastSelectionFrameShown, false)
	f.content = c
	return f
}

func (f *selectionFrame) isShown() bool {
	return !f.content.Hidden
}

func (f *selectionFrame) show() {
	f.content.Show()
}

func (f *selectionFrame) hide() {
	f.content.Hide()
}

func (f *selectionFrame) enable() {
	f.jumpToSelection.Enable()
	f.copyKeyClipboard.Enable()
}

func (f *selectionFrame) disable() {
	f.jumpToSelection.Disable()
	f.copyKeyClipboard.Disable()
}

func (f *selectionFrame) reset() {
	f.selectedPath.RemoveAll()
	f.disable()
	f.selectedUID = ""
}

type NodePlus struct {
	jsondocument.Node
	UID string
}

func (f *selectionFrame) set(uid string) {
	f.selectedUID = uid
	p := f.ui.document.Path(uid)
	var path []NodePlus
	for _, uid2 := range p {
		path = append(path, NodePlus{Node: f.ui.document.Value(uid2), UID: uid2})
	}
	path = append(path, NodePlus{Node: f.ui.document.Value(uid), UID: uid})
	f.selectedPath.RemoveAll()
	for i, n := range path {
		isLast := i == len(path)-1
		if !isLast {
			l := widgets.NewTappableLabel(n.Key, func() {
				f.ui.scrollTo(n.UID)
				f.ui.selectElement(n.UID)
			})
			f.selectedPath.Add(l)
		} else {
			l := widget.NewLabel(n.Key)
			l.TextStyle.Bold = true
			f.selectedPath.Add(l)
		}
		if !isLast {
			l := widget.NewLabel("＞")
			l.Importance = widget.LowImportance
			f.selectedPath.Add(l)
		}
	}
}
