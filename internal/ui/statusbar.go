package ui

import (
	"fmt"
	"log/slog"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/ErikKalkoken/janice/internal/github"
)

const (
	githubOwner = "ErikKalkoken"
	githubRepo  = "janice"
)

// statusBarFrame represents the status bar frame in the UI.
type statusBarFrame struct {
	content *fyne.Container
	u       *UI

	elementsCount *ttwidget.Label
}

func (u *UI) newStatusBarFrame() *statusBarFrame {
	f := &statusBarFrame{
		u:             u,
		elementsCount: ttwidget.NewLabel(""),
	}
	f.elementsCount.SetToolTip("Total count of elements in the JSON document")
	// status bar frame
	c := container.NewHBox(f.elementsCount)
	notifyUpdates := u.app.Preferences().BoolWithFallback(settingNotifyUpdates, settingNotifyUpdatesDefault)
	if notifyUpdates {
		go func() {
			current := u.app.Metadata().Version
			latest, isNewer, err := github.AvailableUpdate(githubOwner, githubRepo, current)
			if err != nil {
				slog.Error("Failed to fetch latest version from github", "err", err)
				return
			}
			if !isNewer {
				return
			}
			c.Add(layout.NewSpacer())
			x, _ := url.Parse(websiteURL + "/releases")
			l := ttwidget.NewHyperlink("Update available", x)
			l.SetToolTip(fmt.Sprintf("Newer version %s available for download", latest))
			c.Add(l)
		}()
	}
	f.content = c
	return f
}

func (f *statusBarFrame) reset() {
	f.elementsCount.SetText("")
}
func (f *statusBarFrame) set(size int) {
	p := message.NewPrinter(language.English)
	f.elementsCount.SetText(p.Sprintf("%d elements", size))
}
