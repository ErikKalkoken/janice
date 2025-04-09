package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"

	"github.com/ErikKalkoken/janice/internal/jsondocument"
)

// detail shows the value of a selected item in the UI.
type detail struct {
	widget.BaseWidget

	copyValueClipboard *ttwidget.Button
	u                  *UI
	valueDisplay       *widget.RichText
	valueRaw           string
}

func newDetail(u *UI) *detail {
	w := &detail{
		u:            u,
		valueDisplay: widget.NewRichText(),
	}
	w.ExtendBaseWidget(w)
	w.copyValueClipboard = ttwidget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		u.window.Clipboard().SetContent(w.valueRaw)
	})
	w.copyValueClipboard.SetToolTip("Copy value to clipboard")
	w.copyValueClipboard.Disable()
	return w
}

func (w *detail) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewBorder(
		nil,
		nil,
		nil,
		w.copyValueClipboard,
		container.NewScroll(w.valueDisplay),
	)
	return widget.NewSimpleRenderer(c)
}

func (w *detail) isShown() bool {
	return !w.Hidden
}

func (w *detail) reset() {
	w.valueDisplay.ParseMarkdown("")
	w.copyValueClipboard.Disable()
}

func (w *detail) set(uid widget.TreeNodeID) {
	node := w.u.document.Value(uid)
	typeText := fmt.Sprint(node.Type)
	var v string
	if w.u.document.IsBranch(uid) {
		w.copyValueClipboard.Disable()
		switch node.Type {
		case jsondocument.Array:
			v = "[...]"
		case jsondocument.Object:
			v = "{...}"
		}
		ids := w.u.document.ChildUIDs(uid)
		typeText += fmt.Sprintf(", %d elements", len(ids))
	} else {
		w.copyValueClipboard.Enable()
		switch node.Type {
		case jsondocument.String:
			x := node.Value.(string)
			v = fmt.Sprintf("\"%s\"", x)
			w.valueRaw = x
		case jsondocument.Number:
			x := node.Value.(float64)
			v = strconv.FormatFloat(x, 'f', -1, 64)
			w.valueRaw = v
		case jsondocument.Null:
			v = "null"
			w.valueRaw = v
		default:
			v = fmt.Sprint(node.Value)
			w.valueRaw = v
		}
	}
	w.valueDisplay.ParseMarkdown(fmt.Sprintf("```\n%s\n```", v))
}
