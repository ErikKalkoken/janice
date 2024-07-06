package jsondocument

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"fyne.io/fyne/v2/data/binding"
	"github.com/stretchr/testify/assert"
)

var dummy1 = binding.NewInt()

func TestLoadFile(t *testing.T) {
	t.Run("should return unmarshaled data from stream", func(t *testing.T) {
		// given
		data := map[string]any{"alpha": "two"}
		dat, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}
		r := bytes.NewReader(dat)
		// when
		got, err := loadFile(r, dummy1)
		// then
		if assert.NoError(t, err) {
			assert.Equal(t, data, got)
		}
	})
	t.Run("should return error when stream can not be unmarshaled", func(t *testing.T) {
		// given
		r := strings.NewReader("invalid JSON")
		// when
		_, err := loadFile(r, dummy1)
		// then
		assert.Error(t, err)
	})
}

func TestAddNode(t *testing.T) {
	t.Run("can add valid parent node", func(t *testing.T) {
		j := New()
		uid := j.addNode("", "alpha", "one", String)
		assert.Equal(t, 1, j.Size())
		assert.Equal(t, Node{Key: "alpha", Value: "one", Type: String}, j.Value(uid))
	})
	t.Run("can add valid child node", func(t *testing.T) {
		j := New()
		uid1 := j.addNode("", "alpha", "one", String)
		uid2 := j.addNode(uid1, "bravo", "two", String)
		assert.Equal(t, 2, j.Size())
		assert.Equal(t, Node{Key: "bravo", Value: "two", Type: String}, j.Value(uid2))
	})
	t.Run("should panic when parent UID does not exist", func(t *testing.T) {
		j := New()
		assert.Panics(t, func() {
			j.addNode("invalid", "alpha", "one", String)
		})
	})
}
