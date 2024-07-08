package jsondocument

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFile(t *testing.T) {
	j := New()
	t.Run("should return unmarshaled data from stream", func(t *testing.T) {
		// given
		data := map[string]any{"alpha": "two"}
		dat, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}
		r := bytes.NewReader(dat)
		// when
		byt, err := j.loadFile(r)
		if assert.NoError(t, err) {
			got, err := j.parseFile(byt)
			// then
			if assert.NoError(t, err) {
				assert.Equal(t, data, got)
			}
		}
	})
	t.Run("should return error when stream can not be unmarshaled", func(t *testing.T) {
		// given
		r := strings.NewReader("invalid JSON")
		// when
		byt, err := j.loadFile(r)
		if assert.NoError(t, err) {
			_, err := j.parseFile(byt)
			// then
			assert.Error(t, err)
		}
	})
}

func TestAddNode(t *testing.T) {
	t.Run("can add valid parent node", func(t *testing.T) {
		j := New()
		j.initialize(10)
		id := j.addNode(0, "alpha", "one", String)
		assert.Equal(t, 1, j.Size())
		assert.Equal(t, Node{Key: "alpha", Value: "one", Type: String}, j.values[id])
	})
	t.Run("can add valid child node", func(t *testing.T) {
		j := New()
		j.initialize(10)
		id1 := j.addNode(0, "alpha", "one", String)
		id2 := j.addNode(id1, "bravo", "two", String)
		assert.Equal(t, 2, j.Size())
		assert.Equal(t, Node{Key: "bravo", Value: "two", Type: String}, j.values[id2])
	})
	t.Run("should panic when parent UID does not exist", func(t *testing.T) {
		j := New()
		j.initialize(10)
		assert.Panics(t, func() {
			j.addNode(1000, "alpha", "one", String)
		})
	})
}
