package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// TappableLabel is a Label that can be tapped.
type TappableLabel struct {
	widget.Label

	// The function that is called when the label is tapped.
	OnTapped func()

	hovered bool
}

var _ fyne.Tappable = (*TappableLabel)(nil)
var _ desktop.Hoverable = (*TappableLabel)(nil)

// NewTappableLabel returns a new TappableLabel instance.
func NewTappableLabel(text string, tapped func()) *TappableLabel {
	l := &TappableLabel{OnTapped: tapped}
	l.ExtendBaseWidget(l)
	l.SetText(text)
	return l
}

func (l *TappableLabel) Tapped(_ *fyne.PointEvent) {
	if l.OnTapped != nil {
		l.OnTapped()
	}
}

// Cursor returns the cursor type of this widget
func (l *TappableLabel) Cursor() desktop.Cursor {
	if l.hovered {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

// MouseIn is a hook that is called if the mouse pointer enters the element.
func (l *TappableLabel) MouseIn(e *desktop.MouseEvent) {
	l.hovered = true
}

func (l *TappableLabel) MouseMoved(*desktop.MouseEvent) {
	// needed to satisfy the interface only
}

// MouseOut is a hook that is called if the mouse pointer leaves the element.
func (l *TappableLabel) MouseOut() {
	l.hovered = false
}
