package ui

import (
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	kxwidget "github.com/ErikKalkoken/fyne-kx/widget"
)

const (
	themeAuto  = "auto"
	themeDark  = "dark"
	themeLight = "light"
)

// setting keys and defaults
const (
	settingExtensionDefault       = true
	settingExtensionFilter        = "extension-filter"
	settingNotifyUpdates          = "notify-updates"
	settingNotifyUpdatesDefault   = true
	settingRecentFileCount        = "recent-file-count"
	settingRecentFileCountDefault = 5
	settingTheme                  = "theme"
	settingThemeDefault           = themeAuto
)

func (u *UI) showSettingsDialog() {
	// recent files
	recentEntry := kxwidget.NewSlider(3, 20, 1)
	recentEntry.SetValue(
		u.app.Preferences().IntWithFallback(settingRecentFileCount, settingRecentFileCountDefault),
	)
	recentEntry.OnChangeEnded = func(v int) {
		u.app.Preferences().SetInt(settingRecentFileCount, v)
	}

	// apply file filter
	extFilter := widget.NewCheck("enabled", func(v bool) {
		u.app.Preferences().SetBool(settingExtensionFilter, v)
	})
	y := u.app.Preferences().BoolWithFallback(settingExtensionFilter, settingExtensionDefault)
	extFilter.SetChecked(y)

	notifyUpdates := widget.NewCheck("enabled", func(v bool) {
		u.app.Preferences().SetBool(settingNotifyUpdates, v)
	})
	z := u.app.Preferences().BoolWithFallback(settingNotifyUpdates, settingNotifyUpdatesDefault)
	notifyUpdates.SetChecked(z)

	themeChoice := widget.NewRadioGroup(
		[]string{themeAuto, themeDark, themeLight}, func(v string) {
			u.app.Preferences().SetString(settingTheme, v)
			u.setTheme(v)
		},
	)
	initialTheme := u.app.Preferences().StringWithFallback(settingTheme, settingThemeDefault)
	themeChoice.SetSelected(initialTheme)

	items := []*widget.FormItem{
		{Text: "Max recent files", Widget: recentEntry, HintText: "Maximum number of recent files remembered"},
		{Text: "JSON file filter", Widget: extFilter, HintText: "Wether to show files with .json extension only"},
		{Text: "Notify about updates", Widget: notifyUpdates, HintText: "Wether to notify when an update is available (requires restart)"},
		{Text: "Theme", Widget: themeChoice, HintText: "Choose the preferred theme"},
	}
	d := dialog.NewCustom("Settings", "Close", widget.NewForm(items...), u.window)
	d.Show()
}
