package ui

import (
	"fmt"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (u *UI) showAboutDialog() {
	c := container.NewVBox()
	info := u.app.Metadata()
	name := u.appName()
	current := info.Version
	appData := widget.NewRichTextFromMarkdown(fmt.Sprintf(
		"## %s\n**Version:** %s", name, current))
	c.Add(appData)
	c.Add(widget.NewLabel("(c) 2024 Erik Kalkoken"))
	d := dialog.NewCustom("About", "OK", c, u.window)
	d.Show()
}
