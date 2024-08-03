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

// valueFrame represents the value frame in the UI.
type valueFrame struct {
	content *fyne.Container
	ui      *UI

	copyValueClipboard *ttwidget.Button
	valueDisplay       *widget.RichText
	valueRaw           string
}

func (u *UI) newValueFrame() *valueFrame {
	f := &valueFrame{
		ui:           u,
		valueDisplay: widget.NewRichText(),
	}
	// value frame
	f.copyValueClipboard = ttwidget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		u.window.Clipboard().SetContent(f.valueRaw)
	})
	f.copyValueClipboard.SetToolTip("Copy value to clipboard")
	f.copyValueClipboard.Disable()
	c := container.NewBorder(
		nil,
		nil,
		nil,
		f.copyValueClipboard,
		container.NewScroll(f.valueDisplay),
	)
	c.Hidden = !u.app.Preferences().BoolWithFallback(preferenceLastValueFrameShown, true)
	f.content = c
	return f
}

func (f *valueFrame) isShown() bool {
	return !f.content.Hidden
}

func (f *valueFrame) show() {
	f.content.Show()
}

func (f *valueFrame) hide() {
	f.content.Hide()
}
func (f *valueFrame) reset() {
	f.valueDisplay.ParseMarkdown("")
	f.copyValueClipboard.Disable()
}

func (f *valueFrame) set(uid widget.TreeNodeID) {
	node := f.ui.document.Value(uid)
	typeText := fmt.Sprint(node.Type)
	var v string
	if f.ui.document.IsBranch(uid) {
		f.copyValueClipboard.Disable()
		switch node.Type {
		case jsondocument.Array:
			v = "[...]"
		case jsondocument.Object:
			v = "{...}"
		}
		ids := f.ui.document.ChildUIDs(uid)
		typeText += fmt.Sprintf(", %d elements", len(ids))
	} else {
		f.copyValueClipboard.Enable()
		switch node.Type {
		case jsondocument.String:
			x := node.Value.(string)
			v = fmt.Sprintf("\"%s\"", x)
			f.valueRaw = x
		case jsondocument.Number:
			x := node.Value.(float64)
			v = strconv.FormatFloat(x, 'f', -1, 64)
			f.valueRaw = v
		case jsondocument.Null:
			v = "null"
			f.valueRaw = v
		default:
			v = fmt.Sprint(node.Value)
			f.valueRaw = v
		}
	}
	f.valueDisplay.ParseMarkdown(fmt.Sprintf("```\n%s\n```", v))
}
