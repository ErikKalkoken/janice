package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type tappableLabel struct {
	widget.Label

	action func()
}

func newTappableLabel(text string, action func()) *tappableLabel {
	l := &tappableLabel{action: action}
	l.ExtendBaseWidget(l)
	l.SetText(text)
	return l
}

func (l *tappableLabel) Tapped(_ *fyne.PointEvent) {
	if l.action != nil {
		l.action()
	}
}
