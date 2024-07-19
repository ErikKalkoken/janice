package ui

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (u *UI) showAboutDialog() {
	info := u.app.Metadata()
	current := info.Version
	x, _ := url.Parse(websiteURL)
	c := container.NewVBox(
		widget.NewRichTextFromMarkdown(
			fmt.Sprintf("## %s\n\n"+
				"**Version:** %s\n\n"+
				"(c) 2024 Erik Kalkoken", info.Name, current),
		),
		widget.NewLabel("A desktop app for viewing large JSON files."),
		widget.NewHyperlink("Website", x),
	)
	d := dialog.NewCustom("About", "OK", c, u.window)
	d.Show()
}
