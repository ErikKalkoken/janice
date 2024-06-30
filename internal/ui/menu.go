package ui

import (
	"encoding/json"
	"io"
	"log"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

func makeMenu(u *UI) *fyne.MainMenu {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open File...", func() {
			d1 := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					dialog.ShowError(err, u.window)
					return
				}
				if reader == nil {
					return
				}
				defer reader.Close()
				progress := binding.NewFloat()
				c := container.NewVBox(
					widget.NewLabel("Loading file. Please wait..."),
					widget.NewProgressBarWithData(progress),
				)
				d2 := dialog.NewCustomWithoutButtons("Loading", c, u.window)
				d2.Show()
				data := loadFile(reader)
				u.setData(data, progress)
				d2.Hide()
				u.setTitle(reader.URI().Name())
			}, u.window)
			f := storage.NewExtensionFileFilter([]string{".json"})
			d1.SetFilter(f)
			d1.Show()
		}))
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Documentation", func() {
			url, _ := url.Parse("https://developer.fyne.io")
			_ = u.app.OpenURL(url)
		}))

	main := fyne.NewMainMenu(fileMenu, helpMenu)
	return main
}

func loadFile(reader fyne.URIReadCloser) any {
	dat, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}
	var data any
	if err := json.Unmarshal(dat, &data); err != nil {
		log.Fatalf("failed to unmarshal JSON: %s", err)
	}
	log.Printf("Read and unmarshaled JSON file")
	return data
}
