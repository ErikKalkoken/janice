package jsondocument

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestLoadFile(t *testing.T) {
// 	ctx := context.Background()
// 	j := New()
// 	t.Run("should return unmarshaled data from stream", func(t *testing.T) {
// 		// given
// 		data := map[string]any{"alpha": "two"}
// 		dat, err := json.Marshal(data)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		r := MakeURIReadCloser(bytes.NewReader(dat), "test")
// 		// when
// 		byt, err := j.loadFile(r)
// 		if assert.NoError(t, err) {
// 			got, err := j.parseFile(ctx, byt)
// 			// then
// 			if assert.NoError(t, err) {
// 				assert.Equal(t, data, got)
// 			}
// 		}
// 	})
// 	t.Run("should return error when stream can not be unmarshaled", func(t *testing.T) {
// 		// given
// 		r := MakeURIReadCloser(strings.NewReader("invalid JSON"), "test")
// 		// when
// 		byt, err := j.loadFile(r)
// 		if assert.NoError(t, err) {
// 			_, err := j.parseFile(ctx, byt)
// 			// then
// 			assert.Error(t, err)
// 		}
// 	})
// }

func TestAddNode(t *testing.T) {
	ctx := context.TODO()
	t.Run("can add root node", func(t *testing.T) {
		j := New()
		j.initialize(10)
		id, err := j.addNode(ctx, -1, "", Empty, Array)
		if assert.NoError(t, err) {
			assert.Equal(t, 1, j.Size())
			assert.Equal(t, Node{Key: "", Value: Empty, Type: Array}, j.values[id])
		}
	})
	t.Run("can add valid parent node", func(t *testing.T) {
		j := New()
		j.initialize(10)
		j.addNode(ctx, -1, "", Empty, Array)
		id, err := j.addNode(ctx, 0, "alpha", "one", String)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, j.Size())
			assert.Equal(t, Node{Key: "alpha", Value: "one", Type: String}, j.values[id])
		}
	})
	t.Run("can add valid child node", func(t *testing.T) {
		j := New()
		j.initialize(10)
		j.addNode(ctx, -1, "", Empty, Array)
		id1, _ := j.addNode(ctx, 0, "alpha", "one", String)
		id2, err := j.addNode(ctx, id1, "bravo", "two", String)
		if assert.NoError(t, err) {
			assert.Equal(t, 3, j.Size())
			assert.Equal(t, Node{Key: "bravo", Value: "two", Type: String}, j.values[id2])
		}
	})
	t.Run("should return error when parent UID does not exist", func(t *testing.T) {
		j := New()
		j.initialize(10)
		j.addNode(ctx, -1, "", Empty, Array)
		_, err := j.addNode(ctx, 5, "alpha", "one", String)
		assert.Error(t, err)
	})
}

func TestWildcard2Regex(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"test*", "^test.*$"},
		{"*test", "^.*test$"},
		{"test", "^test$"},
		{"first*second", "^first.*second$"},
	}
	for _, tc := range cases {
		got := wildCardToRegexp(tc.in)
		assert.Equal(t, tc.want, got)
	}
}

// func TestMemoryUsage(t *testing.T) {
// 	size := int32(1_000_000)
// 	ctx := context.Background()
// 	PrintMemUsage()
// 	j := New()
// 	j.initialize(size + 1)
// 	root, _ := j.addNode(ctx, -1, "", Empty, Array)
// 	for i := range size {
// 		k := strconv.Itoa(int(i))
// 		j.addNode(ctx, root, k, i, Number)
// 	}
// 	PrintMemUsage()
// 	t.Fail()
// }

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
