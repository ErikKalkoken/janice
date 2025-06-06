package ui

import (
	"fmt"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/ErikKalkoken/janice/internal/jsondocument"
	"github.com/stretchr/testify/assert"
)

func TestAddToListWithRotation(t *testing.T) {
	t.Run("can add to empty list", func(t *testing.T) {
		var l []string
		l2 := addToListWithRotation(l, "alpha", 5)
		assert.Equal(t, []string{"alpha"}, l2)
	})
	t.Run("should insert new items on top", func(t *testing.T) {
		var l = []string{"alpha"}
		l2 := addToListWithRotation(l, "bravo", 5)
		assert.Equal(t, []string{"bravo", "alpha"}, l2)
	})
	t.Run("should throw away bottom item to keep max length", func(t *testing.T) {
		var l = []string{"alpha", "bravo", "charlie"}
		l2 := addToListWithRotation(l, "delta", 3)
		assert.Equal(t, []string{"delta", "alpha", "bravo"}, l2)
	})
	t.Run("should insert new on top and remove duplicates", func(t *testing.T) {
		var l = []string{"alpha", "bravo"}
		l2 := addToListWithRotation(l, "bravo", 5)
		assert.Equal(t, []string{"bravo", "alpha"}, l2)
	})
}

func TestMakeShortCut(t *testing.T) {
	const macOS = "darwin"
	cases := []struct {
		name         string
		goos         string
		wantKey      fyne.KeyName
		wantModifier fyne.KeyModifier
		wantIsError  bool
	}{
		{"fileNew", "", fyne.KeyN, fyne.KeyModifierControl, false},
		{"fileOpen", "", fyne.KeyO, fyne.KeyModifierControl, false},
		{"fileReload", "", fyne.KeyR, fyne.KeyModifierAlt, false},
		{"fileQuit", "", fyne.KeyQ, fyne.KeyModifierControl, false},
		{"fileSettings", "", fyne.KeyComma, fyne.KeyModifierControl, false},
		{"goBottom", "", fyne.KeyEnd, fyne.KeyModifierControl, false},
		{"goTop", "", fyne.KeyHome, fyne.KeyModifierControl, false},

		{"fileNew", macOS, fyne.KeyN, fyne.KeyModifierSuper, false},
		{"fileOpen", macOS, fyne.KeyO, fyne.KeyModifierSuper, false},
		{"fileReload", macOS, fyne.KeyR, fyne.KeyModifierAlt, false},
		{"fileSettings", macOS, fyne.KeyComma, fyne.KeyModifierSuper, false},
		{"goBottom", macOS, fyne.KeyDown, fyne.KeyModifierSuper, false},
		{"goTop", macOS, fyne.KeyUp, fyne.KeyModifierSuper, false},

		{"invalid", "", fyne.KeyN, fyne.KeyModifierControl, true},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s %s %v", tc.name, tc.goos, tc.wantIsError), func(t *testing.T) {
			gotShortCut, gotErr := makeShortCut(tc.name, tc.goos)
			if tc.wantIsError {
				assert.Error(t, gotErr)
			} else {
				if assert.NoError(t, gotErr) {
					assert.Equal(t, tc.wantKey, gotShortCut.KeyName)
					assert.Equal(t, tc.wantModifier, gotShortCut.Modifier)
				}
			}
		})
	}
}

func TestCanLoadDocument(t *testing.T) {
	a := test.NewTempApp(t)
	u, err := NewUI(a)
	assert.NoError(t, err)
	u.window.Show()
	x := jsondocument.MakeURIReadCloser(strings.NewReader(`{"alpha": 1}`), "dummy")
	ch := make(chan struct{})
	u.loadDocument(x, func() {
		close(ch)
	})
	<-ch
	assert.Equal(t, 2, u.document.Size())
}
