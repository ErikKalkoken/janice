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

func TestJsonDocument(t *testing.T) {
	i1 := binding.NewInt()
	i2 := binding.NewInt()
	t.Run("can load object", func(t *testing.T) {
		// given
		j := jsondocument.NewJSONDocument()
		data := map[string]any{
			"master": map[string]any{"alpha": "abc", "bravo": 5, "charlie": true, "delta": nil}}
		// when
		err := j.Load(makeDataReader(data), i1, i2)
		// then
		if assert.NoError(t, err) {
			ids := j.ChildUIDs("")
			want := jsondocument.Node{Key: "master", Value: jsondocument.Empty, Type: jsondocument.Object}
			got := j.Value(ids[0])
			assert.Equal(t, want, got)
			for i, id := range j.ChildUIDs(ids[0]) {
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
