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
		assert.Equal(t, 5, j.Size())
	})
	t.Run("should return true when branch", func(t *testing.T) {
		assert.True(t, j.IsBranch(alphaID))
		assert.False(t, j.IsBranch(bravoID))
	})
	t.Run("should return value of parent node", func(t *testing.T) {
		got := j.Value(alphaID)
		want := jsondocument.Node{UID: alphaID, Key: "alpha", Value: jsondocument.Empty, Type: jsondocument.Object}
		assert.Equal(t, want, got)
	})
	t.Run("should return value of child node", func(t *testing.T) {
		got := j.Value(deltaID)
		want := jsondocument.Node{UID: deltaID, Key: "delta", Value: float64(1), Type: jsondocument.Number}
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
			assert.Equal(t, 10, j.Size())
			for i, id := range j.ChildUIDs("") {
				node := j.Value(id)
				switch i {
				case 0:
					assert.Equal(t, node, jsondocument.Node{UID: node.UID, Key: "alpha", Value: "abc", Type: jsondocument.String})
				case 1:
					assert.Equal(t, node, jsondocument.Node{UID: node.UID, Key: "bravo", Value: float64(5), Type: jsondocument.Number})
				case 2:
					assert.Equal(t, node, jsondocument.Node{UID: node.UID, Key: "charlie", Value: true, Type: jsondocument.Boolean})
				case 3:
					assert.Equal(t, node, jsondocument.Node{UID: node.UID, Key: "delta", Value: nil, Type: jsondocument.Null})
				case 4:
					assert.Equal(t, node, jsondocument.Node{UID: node.UID, Key: "echo", Value: jsondocument.Empty, Type: jsondocument.Array})
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
					assert.Equal(t, node, jsondocument.Node{UID: node.UID, Key: "foxtrot", Value: jsondocument.Empty, Type: jsondocument.Object})
					for n, childId := range j.ChildUIDs(id) {
						node := j.Value(childId)
						switch n {
						case 0:
							assert.Equal(t, node, jsondocument.Node{UID: node.UID, Key: "child", Value: float64(1), Type: jsondocument.Number})
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
			assert.Equal(t, 3, j.Size())
			for i, id := range j.ChildUIDs("") {
				node := j.Value(id)
				switch i {
				case 0:
					assert.Equal(t, node, jsondocument.Node{UID: "1", Key: "[0]", Value: "one", Type: jsondocument.String})
				case 1:
					assert.Equal(t, node, jsondocument.Node{UID: "2", Key: "[1]", Value: "two", Type: jsondocument.String})
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
			assert.Equal(t, 3, j.Size())
			x, err := info.Get()
			if assert.NoError(t, err) {
				p := x.(jsondocument.ProgressInfo)
				assert.Equal(t, 3, p.Size)
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

func TestJsonDocumentSearchKey(t *testing.T) {
	ctx := context.TODO()
	var dummy = binding.NewUntyped()
	j := jsondocument.New()
	data := map[string]any{
		"alpha": []any{1, 2, 3},
		"bravo": map[string]any{
			"charlie": 5,
			"delta":   map[string]any{"echo": 1, "foxtrot": 2},
		},
		"golf": []any{
			9,
			map[string]any{"echo": 5, "india": 9},
		},
	}
	if err := j.Load(ctx, makeDataReader(data), dummy); err != nil {
		t.Fatal(err)
	}
	ids := j.ChildUIDs("")
	alphaID, bravoID, golfID := ids[0], ids[1], ids[2]
	ids = j.ChildUIDs(bravoID)
	charlieID, deltaID := ids[0], ids[1]
	ids = j.ChildUIDs(deltaID)
	echo1ID, foxtrotID := ids[0], ids[1]
	ids = j.ChildUIDs(golfID)
	ids2 := j.ChildUIDs(ids[1])
	echo2ID, indiaID := ids2[0], ids2[1]
	cases := []struct {
		startUID   string
		key        string
		foundUID   string
		shouldFind bool
	}{
		{echo1ID, "echo", echo2ID, true},
		{"", "alpha", alphaID, true},
		{"", "bravo", bravoID, true},
		{"", "charlie", charlieID, true},
		{"", "delta", deltaID, true},
		{"", "echo", echo1ID, true},
		{"", "foxtrot", foxtrotID, true},
		{"", "golf", golfID, true},
		{"", "india", indiaID, true},
		{bravoID, "foxtrot", foxtrotID, true},
		{echo1ID, "india", indiaID, true},
		{echo1ID, "india", indiaID, true},
		{golfID, "echo", echo2ID, true},
		{indiaID, "alpha", "", false},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("can find %s from %v", tc.key, tc.startUID), func(t *testing.T) {
			got, err := j.SearchKey(tc.startUID, tc.key)
			if !tc.shouldFind {
				assert.ErrorIs(t, err, jsondocument.ErrNotFound)
			} else if assert.NoError(t, err) {
				assert.Equal(t, tc.foundUID, got)
			}
		})
	}

}

func makeDataReader(data any) fyne.URIReadCloser {
	x, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	r := bytes.NewReader(x)
	return jsondocument.MakeURIReadCloser(r, "test")
}
