package jsondocument_test

import (
	"testing"

	"github.com/ErikKalkoken/jsonviewer/internal/jsondocument"
	"github.com/stretchr/testify/assert"
)

func TestJsonDocument(t *testing.T) {
	t.Run("can load object", func(t *testing.T) {
		j := jsondocument.NewJSONDocument()
		data := map[string]any{
			"alpha": map[string]any{"sub": "one"}}
		err := j.Load(data, nil)
		if assert.NoError(t, err) {
			ids := j.ChildUIDs("")
			assert.Equal(t, "alpha", j.Value(ids[0]))
		}
	})
}
