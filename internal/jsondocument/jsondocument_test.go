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
			"alpha": map[string]any{"sub": "one"}}
		// when
		err := j.Load(makeDataReader(data), i1, i2)
		// then
		if assert.NoError(t, err) {
			ids := j.ChildUIDs("")
			want := jsondocument.Node{Key: "alpha", Value: jsondocument.Empty, Type: jsondocument.Object}
			got := j.Value(ids[0])
			assert.Equal(t, want, got)
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
