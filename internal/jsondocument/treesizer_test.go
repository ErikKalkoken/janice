package jsondocument_test

import (
	"testing"

	"github.com/ErikKalkoken/jsonviewer/internal/jsondocument"
	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	t.Run("can size object based tree", func(t *testing.T) {
		data := map[string]any{
			"alpha":   "abc",
			"bravo":   5,
			"charlie": true,
			"delta":   nil,
			"echo":    []any{1, 2},
			"foxtrot": map[string]any{"child": 1},
		}
		c := jsondocument.JSONTreeSizer{}
		x, err := c.Run(data)
		if assert.NoError(t, err) {
			assert.Equal(t, 9, x)
		}
	})
	t.Run("can size array based tree", func(t *testing.T) {
		data := []any{
			"alpha",
			"bravo",
			"charlie",
			[]any{1, 2},
			map[string]any{"child": 1},
		}
		c := jsondocument.JSONTreeSizer{}
		x, err := c.Run(data)
		if assert.NoError(t, err) {
			assert.Equal(t, 8, x)
		}
	})
}
