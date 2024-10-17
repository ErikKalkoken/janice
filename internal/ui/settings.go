package ui

import (
	"errors"
	"log/slog"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
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
	recentEntry := widget.NewEntry()
	recentEntry.OnChanged = func(s string) {
		x, err := strconv.Atoi(recentEntry.Text)
		if err != nil {
			slog.Error("Failed to convert", "err", err)
			return
		}
		u.app.Preferences().SetInt(settingRecentFileCount, x)
	}
	x := u.app.Preferences().IntWithFallback(settingRecentFileCount, settingRecentFileCountDefault)
	recentEntry.SetText(strconv.Itoa(x))
	recentEntry.Validator = newPositiveNumberValidator()

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

// newPositiveNumberValidator ensures entry is a positive number (incl. zero).
func newPositiveNumberValidator() fyne.StringValidator {
	myErr := errors.New("must be positive number")
	return func(text string) error {
		val, err := strconv.Atoi(text)
		if err != nil {
			return myErr
		}
		if val < 0 {
			return myErr
		}
		return nil
	}
}
