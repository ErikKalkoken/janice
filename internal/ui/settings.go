package ui

import (
	"errors"
	"log/slog"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (u *UI) showSettingsDialog() {
	recentEntry := widget.NewEntry()
	x := u.app.Preferences().IntWithFallback(settingRecentFileCount, settingRecentFileCountDefault)
	recentEntry.SetText(strconv.Itoa(x))
	recentEntry.Validator = newPositiveNumberValidator()

	extFilter := widget.NewCheck("enabled", func(bool) {})
	y := u.app.Preferences().BoolWithFallback(settingExtensionFilter, settingExtensionDefault)
	extFilter.SetChecked(y)

	notifyUpdates := widget.NewCheck("enabled", func(bool) {})
	z := u.app.Preferences().BoolWithFallback(settingNotifyUpdates, settingNotifyUpdatesDefault)
	notifyUpdates.SetChecked(z)

	items := []*widget.FormItem{
		{Text: "Max recent files", Widget: recentEntry, HintText: "Maximum number of recent files remembered"},
		{Text: "JSON file filter", Widget: extFilter, HintText: "Wether to apply the JSON extension filter in file lists"},
		{Text: "Notify about updates", Widget: notifyUpdates, HintText: "Wether to notify when a new version is available (requires restart)"},
	}
	d := dialog.NewForm(
		"Preferences", "Apply", "Cancel", items, func(applied bool) {
			if !applied {
				return
			}
			x, err := strconv.Atoi(recentEntry.Text)
			if err != nil {
				slog.Error("Failed to convert", "err", err)
				return
			}
			u.app.Preferences().SetInt(settingRecentFileCount, x)
			u.app.Preferences().SetBool(settingExtensionFilter, extFilter.Checked)
			u.app.Preferences().SetBool(settingNotifyUpdates, notifyUpdates.Checked)
		}, u.window)
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
