package ui

import (
	"fmt"
	"log/slog"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/ErikKalkoken/janice/internal/github"
)

const (
	githubOwner = "ErikKalkoken"
	githubRepo  = "janice"
)

// statusBar represents the status bar frame in the UI.
type statusBar struct {
	widget.BaseWidget

	elementsCount *ttwidget.Label
	updateLink    *ttwidget.Hyperlink
	u             *UI
}

func newStatusBar(u *UI) *statusBar {
	x, _ := url.Parse(websiteURL + "/releases")
	w := &statusBar{
		elementsCount: ttwidget.NewLabel(""),
		updateLink:    ttwidget.NewHyperlink("Update available", x),
		u:             u,
	}
	w.ExtendBaseWidget(w)
	w.elementsCount.SetToolTip("Total count of elements in the JSON document")
	w.updateLink.Hide()
	notifyUpdates := w.u.app.Preferences().BoolWithFallback(settingNotifyUpdates, settingNotifyUpdatesDefault)
	if notifyUpdates {
		go func() {
			current := w.u.app.Metadata().Version
			latest, isNewer, err := github.AvailableUpdate(githubOwner, githubRepo, current)
			if err != nil {
				slog.Error("Failed to fetch latest version from github", "err", err)
				return
			}
			if !isNewer {
				return
			}
			w.updateLink.SetToolTip(fmt.Sprintf("Newer version %s available for download", latest))
			w.updateLink.Show()
		}()
	}
	return w
}

func (f *statusBar) reset() {
	f.elementsCount.SetText("")
}

func (f *statusBar) set(size int) {
	p := message.NewPrinter(language.English)
	f.elementsCount.SetText(p.Sprintf("%d elements", size))
}

func (w *statusBar) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewHBox(w.elementsCount, layout.NewSpacer(), w.updateLink)
	return widget.NewSimpleRenderer(c)
}
