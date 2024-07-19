package ui

import (
	"fmt"
	"log/slog"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/ErikKalkoken/janice/internal/github"
	"github.com/ErikKalkoken/janice/internal/widgets"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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
			latest, isNewer, err := github.AvailableUpdate(githubOwner, githubRepo, current)
			if err != nil {
				slog.Error("Failed to fetch latest version from github", "err", err)
				return
			}
			if !isNewer {
				return
			}
			c.Add(layout.NewSpacer())
			l := widgets.NewTappableLabel("Update available", func() {
				f.showReleaseDialog(current, latest)
			})
			l.Importance = widget.HighImportance
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

func (f *statusBarFrame) showReleaseDialog(current, latest string) {
	x := websiteURL + "/releases"
	y, _ := url.Parse(x)
	link := widget.NewHyperlink(x, y)
	c := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("New release %s available. You have release %s.", latest, current)),
		widget.NewLabel("The link below will bring you to the releases page,\n"+
			"where you can download the latest version:"),
		link,
	)
	d := dialog.NewCustom("Update available", "OK", c, f.ui.window)
	d.Show()
}
