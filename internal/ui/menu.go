package ui

import (
	"bytes"
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
				infoText := binding.NewString()
				c := container.NewVBox(
					widget.NewLabelWithData(infoText),
					widget.NewProgressBarWithData(u.document.Progress),
				)
				d2 := dialog.NewCustomWithoutButtons("Loading", c, u.window)
				d2.Show()
				infoText.Set("Loading file... Please Wait.")
				data, n := loadFile(reader)
				infoText.Set("Parsing file... Please Wait.")
				u.setData(data, n)
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

func loadFile(reader fyne.URIReadCloser) (any, int) {
	dat, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}
	var data any
	if err := json.Unmarshal(dat, &data); err != nil {
		log.Fatalf("failed to unmarshal JSON: %s", err)
	}
	log.Printf("Read and unmarshaled JSON file")
	n := bytes.Count(dat, []byte{'\n'})
	log.Printf("File has %d LOC", n)
	return data, n
}
