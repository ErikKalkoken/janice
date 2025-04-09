package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// treeNode represents a node in a JSON document tree.
type treeNode struct {
	widget.BaseWidget

	key   *widget.Label
	value *widget.Label
}

// newTreeNode returns a new instance of the [treeNode] widget.
func newTreeNode() *treeNode {
	w := &treeNode{
		key:   widget.NewLabel(""),
		value: widget.NewLabel(""),
	}
	w.ExtendBaseWidget(w)
	return w
}

func (w *treeNode) set(key string, value string, importance widget.Importance) {
	w.key.SetText(fmt.Sprintf("%s :", key))
	w.value.Importance = importance
	w.value.Text = strings.ReplaceAll(value, "\n", " ")
	w.value.Refresh()
	w.value.Truncation = fyne.TextTruncateEllipsis
}

func (w *treeNode) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewBorder(nil, nil, w.key, nil, w.value)
	return widget.NewSimpleRenderer(c)
}
