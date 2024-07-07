// Package jsondocument contains the logic for building a Fyne tree from a JSON document.
package jsondocument

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"sync"

	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

const (
	// Update progress after x added nodes
	progressUpdateTick = 10_000
	totalLoadSteps     = 4
)

type JSONType uint

const (
	Unknown JSONType = iota
	Array
	Boolean
	Null
	Number
	Object
	String
)

type Node struct {
	Key   string
	Value any
	Type  JSONType
}

type ProgressInfo struct {
	Progress    float64
	Size        int
	CurrentStep int
	TotalSteps  int
}

// This singleton represents an empty value in a Node.
var Empty = struct{}{}

// JSONDocument represents a JSON document which can be rendered by a Fyne tree widget.
type JSONDocument struct {
	progressInfo  binding.Untyped
	elementsCount int

	mu     sync.RWMutex
	ids    map[widget.TreeNodeID][]widget.TreeNodeID
	values map[widget.TreeNodeID]Node
	n      int
}

// Returns a new JSONDocument object.
func New() *JSONDocument {
	t := &JSONDocument{progressInfo: binding.NewUntyped()}
	t.reset()
	return t
}

// ChildUIDs returns the child UIDs for a given node.
// This can be used directly in the tree widget childUIDs() function.
func (t *JSONDocument) ChildUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	if !t.mu.TryRLock() {
		// This method can be called by another goroutine from the Fyne library while a new tree is loaded.
		// This can not block, or it would block the whole Fyne app.
		return []widget.TreeNodeID{}
	}
	defer t.mu.RUnlock()
	return t.ids[uid]
}

// IsBranch reports wether a node is a branch.
// This can be used directly in the tree widget isBranch() function.
func (t *JSONDocument) IsBranch(uid widget.TreeNodeID) bool {
	if !t.mu.TryRLock() {
		return false
	}
	defer t.mu.RUnlock()
	_, found := t.ids[uid]
	return found
}

// Value returns the value of a node
func (t *JSONDocument) Value(uid widget.TreeNodeID) Node {
	if !t.mu.TryRLock() {
		return Node{}
	}
	defer t.mu.RUnlock()
	return t.values[uid]
}

// Load loads a new tree from a reader.
func (t *JSONDocument) Load(reader io.Reader, progressInfo binding.Untyped) error {
	t.progressInfo = progressInfo
	data, err := t.loadFile(reader)
	if err != nil {
		return err
	}
	if err := t.setProgressInfo(ProgressInfo{CurrentStep: 3}); err != nil {
		return err
	}
	c := JSONTreeSizer{}
	s, err := c.Run(data)
	if err != nil {
		return err
	}
	t.elementsCount = s
	slog.Info("Tree size calculated", "size", s)
	if err := t.setProgressInfo(ProgressInfo{CurrentStep: 4}); err != nil {
		return err
	}
	if err := t.load(data); err != nil {
		return err
	}
	slog.Info("Finished loading JSON document into tree", "size", t.n)
	return nil
}

func (t *JSONDocument) load(data any) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.reset()
	switch v := data.(type) {
	case map[string]any:
		t.addObject("", v)
	case []any:
		t.addArray("", v)
	default:
		return fmt.Errorf("unrecognized format")
	}
	return nil
}

// Size returns the number of nodes.
func (t *JSONDocument) Size() int {
	if !t.mu.TryRLock() {
		return 0
	}
	defer t.mu.RUnlock()
	return t.n
}

// addObject adds a JSON object to the tree.
func (t *JSONDocument) addObject(parentUID widget.TreeNodeID, data map[string]any) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		v := data[k]
		t.addValue(parentUID, k, v)
	}
}

// addArray adds a JSON array to the tree.
func (t *JSONDocument) addArray(parentUID string, a []any) {
	for i, v := range a {
		k := fmt.Sprintf("[%d]", i)
		t.addValue(parentUID, k, v)
	}
}

// addValue adds a JSON value to the tree.
func (t *JSONDocument) addValue(parentUID widget.TreeNodeID, k string, v any) {
	switch v2 := v.(type) {
	case map[string]any:
		uid := t.addNode(parentUID, k, Empty, Object)
		t.addObject(uid, v2)
	case []any:
		uid := t.addNode(parentUID, k, Empty, Array)
		t.addArray(uid, v2)
	case string:
		t.addNode(parentUID, k, v2, String)
	case float64:
		t.addNode(parentUID, k, v2, Number)
	case bool:
		t.addNode(parentUID, k, v2, Boolean)
	case nil:
		t.addNode(parentUID, k, v2, Null)
	default:
		t.addNode(parentUID, k, v2, Unknown)
	}
}

// addNode adds a node to the tree and returns the UID.
// Nodes will be rendered in the same order they are added.
// Use "" as parentUID for adding nodes at the top level.
// Returns the generated UID for this node and the incremented ID
func (t *JSONDocument) addNode(parentUID widget.TreeNodeID, key string, value any, typ JSONType) widget.TreeNodeID {
	if parentUID != "" {
		_, found := t.values[parentUID]
		if !found {
			panic(fmt.Sprintf("parent UID does not exist: %s", parentUID))
		}
	}
	s := parentUID
	if parentUID == "" {
		s = "ID"
	}
	uid := fmt.Sprintf("%s-%d", s, t.n)
	_, found := t.values[uid]
	if found {
		panic(fmt.Sprintf("UID for this node already exists: %v", uid))
	}
	t.ids[parentUID] = append(t.ids[parentUID], uid)
	t.values[uid] = Node{Key: key, Value: value, Type: typ}
	t.n++
	if t.n%progressUpdateTick == 0 {
		p := float64(t.n) / float64(t.elementsCount)
		if err := t.setProgressInfo(ProgressInfo{CurrentStep: 4, Progress: p}); err != nil {
			slog.Warn("Failed to set progress", "err", err)
		}
	}
	return uid
}

// reset re-initializes the tree so a new tree can be build.
func (t *JSONDocument) reset() {
	t.values = make(map[widget.TreeNodeID]Node)
	t.ids = make(map[widget.TreeNodeID][]widget.TreeNodeID)
	t.n = 0
}

func (t *JSONDocument) setProgressInfo(info ProgressInfo) error {
	info.TotalSteps = totalLoadSteps
	info.Size = t.elementsCount
	if err := t.progressInfo.Set(info); err != nil {
		return err
	}
	return nil
}

func (t *JSONDocument) loadFile(reader io.Reader) (any, error) {
	if err := t.setProgressInfo(ProgressInfo{CurrentStep: 1}); err != nil {
		return nil, err
	}
	dat, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err)
	}
	if err := t.setProgressInfo(ProgressInfo{CurrentStep: 2}); err != nil {
		return nil, err
	}
	var data any
	if err := json.Unmarshal(dat, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %s", err)
	}
	slog.Info("Read and unmarshaled JSON file")
	return data, nil
}
