package jsondocument_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"fyne.io/fyne/v2/data/binding"
	"github.com/ErikKalkoken/jsonviewer/internal/jsondocument"
	"github.com/stretchr/testify/assert"
)

var dummy1 = binding.NewInt()
var dummy2 = binding.NewInt()

func TestJsonDocument(t *testing.T) {
	j := jsondocument.New()
	data := map[string]any{
		"alpha": map[string]any{"second": 1},
		"bravo": 2,
	}
	if err := j.Load(makeDataReader(data), dummy1, dummy2); err != nil {
		t.Fatal(err)
	}
	ids := j.ChildUIDs("")
	alphaID := ids[0]
	bravoID := ids[1]
	t.Run("should return tree size", func(t *testing.T) {
		assert.Equal(t, 3, j.Size())
	})
	t.Run("should return true when branch", func(t *testing.T) {
		assert.True(t, j.IsBranch(alphaID))
		assert.False(t, j.IsBranch(bravoID))
	})
	t.Run("should return node value", func(t *testing.T) {
		got := j.Value(alphaID)
		want := jsondocument.Node{Key: "alpha", Value: jsondocument.Empty, Type: jsondocument.Object}
		assert.Equal(t, want, got)
	})
}
func TestJsonDocumentLoad(t *testing.T) {
	t.Run("can load object with values of all types", func(t *testing.T) {
		// given
		j := jsondocument.New()
		data := map[string]any{
			"alpha":   "abc",
			"bravo":   5,
			"charlie": true,
			"delta":   nil,
			"echo":    []int{1, 2},
			"foxtrot": map[string]any{"child": 1},
		}
		// when
		err := j.Load(makeDataReader(data), dummy1, dummy2)
		// then
		if assert.NoError(t, err) {
			assert.Equal(t, 9, j.Size())
			for i, id := range j.ChildUIDs("") {
				node := j.Value(id)
				switch i {
				case 0:
					assert.Equal(t, node, jsondocument.Node{Key: "alpha", Value: "abc", Type: jsondocument.String})
				case 1:
					assert.Equal(t, node, jsondocument.Node{Key: "bravo", Value: float64(5), Type: jsondocument.Number})
				case 2:
					assert.Equal(t, node, jsondocument.Node{Key: "charlie", Value: true, Type: jsondocument.Boolean})
				case 3:
					assert.Equal(t, node, jsondocument.Node{Key: "delta", Value: nil, Type: jsondocument.Null})
				case 4:
					assert.Equal(t, node, jsondocument.Node{Key: "echo", Value: jsondocument.Empty, Type: jsondocument.Array})
					for n, childId := range j.ChildUIDs(id) {
						node := j.Value(childId)
						switch n {
						case 0:
							assert.Equal(t, node.Value, float64(1))
						case 1:
							assert.Equal(t, node.Value, float64(2))
						}
					}
				case 5:
					assert.Equal(t, node, jsondocument.Node{Key: "foxtrot", Value: jsondocument.Empty, Type: jsondocument.Object})
					for n, childId := range j.ChildUIDs(id) {
						node := j.Value(childId)
						switch n {
						case 0:
							assert.Equal(t, node, jsondocument.Node{Key: "child", Value: float64(1), Type: jsondocument.Number})
						}
					}
				}
			}
		}
	})
	t.Run("can load array", func(t *testing.T) {
		// given
		j := jsondocument.New()
		data := []any{"one", "two"}
		// when
		err := j.Load(makeDataReader(data), dummy1, dummy2)
		// then
		if assert.NoError(t, err) {
			assert.Equal(t, 2, j.Size())
			for i, id := range j.ChildUIDs("") {
				node := j.Value(id)
				switch i {
				case 0:
					assert.Equal(t, node, jsondocument.Node{Key: "[0]", Value: "one", Type: jsondocument.String})
				case 1:
					assert.Equal(t, node, jsondocument.Node{Key: "[1]", Value: "two", Type: jsondocument.String})
				}
			}
		}
	})
}

func makeDataReader(data any) io.Reader {
	x, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	r := bytes.NewReader(x)
	return r

}
