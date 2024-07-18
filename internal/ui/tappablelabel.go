package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type tappableLabel struct {
	widget.Label

	OnTapped func()
}

func newTappableLabel(text string, tapped func()) *tappableLabel {
	l := &tappableLabel{OnTapped: tapped}
	l.ExtendBaseWidget(l)
	l.SetText(text)
	return l
}

func (l *tappableLabel) Tapped(_ *fyne.PointEvent) {
	if l.OnTapped != nil {
		l.OnTapped()
	}
}
