package ui

import (
	"log/slog"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/ErikKalkoken/janice/internal/github"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	githubOwner = "ErikKalkoken"
	githubRepo  = "janice"
)

// statusBarFrame represents the status bar frame in the UI.
type statusBarFrame struct {
	content *fyne.Container
	ui      *UI

	statusTreeSize *widget.Label
}

func (u *UI) newStatusBarFrame() *statusBarFrame {
	f := &statusBarFrame{
		ui:             u,
		statusTreeSize: widget.NewLabel(""),
	}
	// status bar frame
	c := container.NewHBox(f.statusTreeSize)
	notifyUpdates := u.app.Preferences().BoolWithFallback(settingNotifyUpdates, settingNotifyUpdatesDefault)
	if notifyUpdates {
		go func() {
			current := u.app.Metadata().Version
			_, isNewer, err := github.AvailableUpdate(githubOwner, githubRepo, current)
			if err != nil {
				slog.Error("Failed to fetch latest version from github", "err", err)
				return
			}
			if !isNewer {
				return
			}
			c.Add(layout.NewSpacer())
			x, _ := url.Parse(websiteURL + "/releases")
			l := widget.NewHyperlink("Update available", x)
			c.Add(l)
		}()
	}
	f.content = c
	return f
}

func (f *statusBarFrame) reset() {
	f.statusTreeSize.SetText("")
}
func (f *statusBarFrame) set(size int) {
	p := message.NewPrinter(language.English)
	f.statusTreeSize.SetText(p.Sprintf("%d elements", size))
}
