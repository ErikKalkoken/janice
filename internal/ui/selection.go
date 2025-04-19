package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	kxwidget "github.com/ErikKalkoken/fyne-kx/widget"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"

	"github.com/ErikKalkoken/janice/internal/jsondocument"
)

// selection shows the currently selected item in the JSON document.
type selection struct {
	widget.BaseWidget

	copyKeyClipboard *ttwidget.Button
	jumpToSelection  *ttwidget.Button
	selectedPath     *fyne.Container
	selectedUID      widget.TreeNodeID
	u                *UI
}

func newSelection(u *UI) *selection {
	w := &selection{
		selectedPath: container.New(layout.NewCustomPaddedHBoxLayout(-5)),
		u:            u,
	}
	w.ExtendBaseWidget(w)
	w.jumpToSelection = ttwidget.NewButtonWithIcon("", theme.NewThemedResource(resourceReadmoreSvg), func() {
		u.tree.scrollTo(w.selectedUID)
	})
	w.jumpToSelection.SetToolTip("Jump to selection")
	w.jumpToSelection.Disable()
	w.copyKeyClipboard = ttwidget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		n := u.document.Value(w.selectedUID)
		u.app.Clipboard().SetContent(n.Key)
	})
	w.copyKeyClipboard.SetToolTip("Copy key to clipboard")
	w.copyKeyClipboard.Disable()
	return w
}

func (w *selection) enable() {
	w.jumpToSelection.Enable()
	w.copyKeyClipboard.Enable()
}

func (w *selection) disable() {
	w.jumpToSelection.Disable()
	w.copyKeyClipboard.Disable()
}

func (w *selection) reset() {
	w.selectedPath.RemoveAll()
	w.disable()
	w.selectedUID = ""
}

type NodePlus struct {
	jsondocument.Node
	UID string
}

func (w *selection) set(uid string) {
	w.selectedUID = uid
	p := w.u.document.Path(uid)
	var path []NodePlus
	for _, uid2 := range p {
		path = append(path, NodePlus{Node: w.u.document.Value(uid2), UID: uid2})
	}
	path = append(path, NodePlus{Node: w.u.document.Value(uid), UID: uid})
	w.selectedPath.RemoveAll()
	for i, n := range path {
		isLast := i == len(path)-1
		if !isLast {
			l := kxwidget.NewTappableLabel(n.Key, func() {
				w.u.tree.scrollTo(n.UID)
				w.u.selectElement(n.UID)
			})
			w.selectedPath.Add(l)
		} else {
			l := widget.NewLabel(n.Key)
			l.TextStyle.Bold = true
			w.selectedPath.Add(l)
		}
		if !isLast {
			l := widget.NewLabel("＞")
			l.Importance = widget.LowImportance
			w.selectedPath.Add(l)
		}
	}
}

func (w *selection) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewBorder(
		nil,
		nil,
		nil,
		container.NewHBox(w.jumpToSelection, w.copyKeyClipboard),
		container.NewHScroll(w.selectedPath),
	)
	return widget.NewSimpleRenderer(c)
}
