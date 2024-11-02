package ui

import (
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	kxdialog "github.com/ErikKalkoken/fyne-kx/dialog"
	kxwidget "github.com/ErikKalkoken/fyne-kx/widget"
)

// setting keys and defaults
const (
	settingExtensionDefault       = true
	settingExtensionFilter        = "extension-filter"
	settingNotifyUpdates          = "notify-updates"
	settingNotifyUpdatesDefault   = true
	settingRecentFileCount        = "recent-file-count"
	settingRecentFileCountDefault = 5
)

func (u *UI) showSettingsDialog() {
	// recent files
	recentEntry := kxwidget.NewSlider(3, 20)
	x := u.app.Preferences().IntWithFallback(settingRecentFileCount, settingRecentFileCountDefault)
	recentEntry.SetValue(float64(x))
	recentEntry.OnChangeEnded = func(v float64) {
		u.app.Preferences().SetInt(settingRecentFileCount, int(v))
	}

	// apply file filter
	extFilter := kxwidget.NewSwitch(func(v bool) {
		u.app.Preferences().SetBool(settingExtensionFilter, v)
	})
	y := u.app.Preferences().BoolWithFallback(settingExtensionFilter, settingExtensionDefault)
	extFilter.SetState(y)

	notifyUpdates := kxwidget.NewSwitch(func(v bool) {
		u.app.Preferences().SetBool(settingNotifyUpdates, v)
	})
	z := u.app.Preferences().BoolWithFallback(settingNotifyUpdates, settingNotifyUpdatesDefault)
	notifyUpdates.SetState(z)

	items := []*widget.FormItem{
		{Text: "Max recent files", Widget: recentEntry, HintText: "Maximum number of recent files remembered"},
		{Text: "JSON file filter", Widget: extFilter, HintText: "Wether to show files with .json extension only"},
		{Text: "Notify about updates", Widget: notifyUpdates, HintText: "Wether to notify when an update is available (requires restart)"},
	}
	d := dialog.NewCustom("Settings", "Close", widget.NewForm(items...), u.window)
	kxdialog.AddDialogKeyHandler(d, u.window)
	d.Show()
}
