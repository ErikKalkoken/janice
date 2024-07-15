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
	x := u.app.Preferences().IntWithFallback(settingsRecentFileCount, settingsRecentFileCountDefault)
	recentEntry.SetText(strconv.Itoa(x))
	recentEntry.Validator = newPositiveNumberValidator()

	extFilterCheck := widget.NewCheck("enabled", func(bool) {})
	y := u.app.Preferences().BoolWithFallback(settingsExtensionFilter, settingsExtensionDefault)
	extFilterCheck.SetChecked(y)
	items := []*widget.FormItem{
		{Text: "Max recent files", Widget: recentEntry, HintText: "Maximum number of recent files remembered"},
		{Text: "JSON file filter", Widget: extFilterCheck, HintText: "Filter applied in file open dialog"},
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
			u.app.Preferences().SetInt(settingsRecentFileCount, x)
			u.app.Preferences().SetBool(settingsExtensionFilter, extFilterCheck.Checked)
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
