package jsondocument_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/ErikKalkoken/jsonviewer/internal/jsondocument"
	"github.com/stretchr/testify/assert"
)

func TestJsonDocument(t *testing.T) {
	ctx := context.TODO()
	var dummy = binding.NewUntyped()
	j := jsondocument.New()
	data := map[string]any{
		"alpha": map[string]any{"charlie": map[string]any{"delta": 1}},
		"bravo": 2,
	}
	if err := j.Load(ctx, makeDataReader(data), dummy); err != nil {
		t.Fatal(err)
	}
	ids := j.ChildUIDs("")
	alphaID := ids[0]
	bravoID := ids[1]
	ids = j.ChildUIDs(alphaID)
	charlieID := ids[0]
	ids = j.ChildUIDs(charlieID)
	deltaID := ids[0]
	t.Run("should return tree size", func(t *testing.T) {
		assert.Equal(t, 4, j.Size())
	})
	t.Run("should return true when branch", func(t *testing.T) {
		assert.True(t, j.IsBranch(alphaID))
		assert.False(t, j.IsBranch(bravoID))
	})
	t.Run("should return value of parent node", func(t *testing.T) {
		got := j.Value(alphaID)
		want := jsondocument.Node{Key: "alpha", Value: jsondocument.Empty, Type: jsondocument.Object}
		assert.Equal(t, want, got)
	})
	t.Run("should return value of child node", func(t *testing.T) {
		got := j.Value(deltaID)
		want := jsondocument.Node{Key: "delta", Value: float64(1), Type: jsondocument.Number}
		assert.Equal(t, want, got)
	})
	t.Run("should return empty path for parent node", func(t *testing.T) {
		got := j.Path(alphaID)
		assert.Len(t, got, 0)
	})
	t.Run("should return path for child node", func(t *testing.T) {
		got := j.Path(deltaID)
		want := []widget.TreeNodeID{alphaID, charlieID}
		assert.Equal(t, want, got)
	})
}
func TestJsonDocumentLoad(t *testing.T) {
	ctx := context.TODO()
	var dummy = binding.NewUntyped()
	t.Run("can load object with values of all types and sort keys", func(t *testing.T) {
		// given
		j := jsondocument.New()
		data := map[string]any{
			"bravo":   5,
			"alpha":   "abc",
			"echo":    []any{1, 2},
			"charlie": true,
			"delta":   nil,
			"foxtrot": map[string]any{"child": 1},
		}
		// when
		err := j.Load(ctx, makeDataReader(data), dummy)
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
		err := j.Load(ctx, makeDataReader(data), dummy)
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
	t.Run("can load JSON and update progress", func(t *testing.T) {
		// given
		info := binding.NewUntyped()
		j := jsondocument.New()
		j.ProgressUpdateTick = 1
		data := []any{"one", "two"}
		// when
		err := j.Load(ctx, makeDataReader(data), info)
		// then
		if assert.NoError(t, err) {
			assert.Equal(t, 2, j.Size())
			x, err := info.Get()
			if assert.NoError(t, err) {
				p := x.(jsondocument.ProgressInfo)
				assert.Equal(t, 2, p.Size)
				assert.Equal(t, 4, p.CurrentStep)
				assert.Equal(t, 4, p.TotalSteps)
			}
		}
	})
}

func TestJsonDocumentExtract(t *testing.T) {
	ctx := context.TODO()
	var dummy = binding.NewUntyped()
	j := jsondocument.New()
	data := map[string]any{
		"alpha": map[string]any{"charlie": map[string]any{"delta": 1}},
		"bravo": []any{1, 2, 3},
	}
	if err := j.Load(ctx, makeDataReader(data), dummy); err != nil {
		t.Fatal(err)
	}
	ids := j.ChildUIDs("")
	alphaID := ids[0]
	bravoID := ids[1]
	ids = j.ChildUIDs(alphaID)
	charlieID := ids[0]
	ids = j.ChildUIDs(charlieID)
	deltaID := ids[0]
	t.Run("can extract object", func(t *testing.T) {
		got, err := j.Extract(alphaID)
		if assert.NoError(t, err) {
			want, _ := json.Marshal(map[string]any{"charlie": map[string]any{"delta": float64(1)}})
			assert.Equal(t, want, got)
		}
	})
	t.Run("can extract array", func(t *testing.T) {
		got, err := j.Extract(bravoID)
		if assert.NoError(t, err) {
			want, _ := json.Marshal([]any{float64(1), float64(2), float64(3)})
			assert.Equal(t, want, got)
		}
	})
	t.Run("should return error when trying to extract normal value", func(t *testing.T) {
		_, err := j.Extract(deltaID)
		assert.Error(t, err)
	})
}

func makeDataReader(data any) fyne.URIReadCloser {
	x, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	r := bytes.NewReader(x)
	return jsondocument.MakeURIReadCloser(r, "test")
}

func TestJSONType(t *testing.T) {
	cases := []struct {
		typ  jsondocument.JSONType
		name string
	}{
		{jsondocument.Array, "array"},
		{jsondocument.Boolean, "boolean"},
		{jsondocument.Null, "null"},
		{jsondocument.Number, "number"},
		{jsondocument.Object, "object"},
		{jsondocument.String, "string"},
		{jsondocument.Undefined, "undefined"},
		{jsondocument.Unknown, "unknown"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("can return name of type %T as string", tc.typ), func(t *testing.T) {
			got := fmt.Sprint(tc.typ)
			assert.Equal(t, tc.name, got)
		})
	}
}
