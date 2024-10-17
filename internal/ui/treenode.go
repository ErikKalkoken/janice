package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// TreeNode represents a node in JSON document tree.
type TreeNode struct {
	widget.BaseWidget
	key   *widget.Label
	value *widget.Label
}

// NewTreeNode returns a new instance of the [TreeNode] widget.
func NewTreeNode() *TreeNode {
	n := &TreeNode{
		key:   widget.NewLabel(""),
		value: widget.NewLabel(""),
	}
	n.ExtendBaseWidget(n)
	return n
}

func (n *TreeNode) Set(key string, value string, importance widget.Importance) {
	n.key.SetText(fmt.Sprintf("%s :", key))
	n.value.Importance = importance
	n.value.Text = strings.ReplaceAll(value, "\n", " ")
	n.value.Refresh()
	n.value.Truncation = fyne.TextTruncateEllipsis
}

func (n *TreeNode) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewBorder(nil, nil, n.key, nil, n.value)
	return widget.NewSimpleRenderer(c)
}
