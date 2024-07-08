// Package jsondocument contains the logic for rendering a Fyne tree from a JSON document.
package jsondocument

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"sync"

	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

const (
	// Update progress after x added nodes
	progressUpdateTick = 10_000
	// Total number of load steps
	totalLoadSteps = 4
)

// JSONType represents the type of a JSON value.
type JSONType uint8

const (
	Unknown JSONType = iota
	Array
	Boolean
	Null
	Number
	Object
	String
)

var typeMap = map[JSONType]string{
	Unknown: "unknown",
	Array:   "array",
	Boolean: "boolean",
	Null:    "null",
	Number:  "number",
	Object:  "object",
	String:  "string",
}

func (t JSONType) String() string {
	s, ok := typeMap[t]
	if !ok {
		return "undefined"
	}
	return s
}

// Node represents a node in the JSON data tree.
type Node struct {
	Key   string
	Value any
	Type  JSONType
}

// ProgressInfo represents the current progress while loading a document
// and is used to communicate the the UI.
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

	mu      sync.RWMutex
	ids     map[int][]int
	values  map[int]Node
	parents map[int]int
	n       int
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
	id := uid2id(uid)
	return ids2uids(t.ids[id])
}

// IsBranch reports wether a node is a branch.
// This can be used directly in the tree widget isBranch() function.
func (t *JSONDocument) IsBranch(uid widget.TreeNodeID) bool {
	if !t.mu.TryRLock() {
		return false
	}
	defer t.mu.RUnlock()
	id := uid2id(uid)
	_, found := t.ids[id]
	return found
}

// Value returns the value of a node
func (t *JSONDocument) Value(uid widget.TreeNodeID) Node {
	if !t.mu.TryRLock() {
		return Node{}
	}
	defer t.mu.RUnlock()
	id := uid2id(uid)
	return t.values[id]
}

// Load loads JSON data from a reader and builds a new JSON document from it.
// It reports it's current progress to the caller via updates to progressInfo.
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
	s, err := c.Calculate(data)
	if err != nil {
		return err
	}
	t.elementsCount = s
	slog.Info("Tree size calculated", "size", s)
	if err := t.setProgressInfo(ProgressInfo{CurrentStep: 4}); err != nil {
		return err
	}
	if err := t.render(data); err != nil {
		return err
	}
	slog.Info("Finished loading JSON document into tree", "size", t.n)
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

// Path returns the path of a node in the tree.
func (t *JSONDocument) Path(uid widget.TreeNodeID) []widget.TreeNodeID {
	path := make([]int, 0)
	if !t.mu.TryRLock() {
		return []widget.TreeNodeID{}
	}
	defer t.mu.RUnlock()
	id := uid2id(uid)
	for {
		id = t.parents[id]
		if id == 0 {
			break
		}
		path = append(path, id)
	}
	slices.Reverse(path)
	return ids2uids(path)
}

// render is the main method for rendering the JSON data into a tree.
func (t *JSONDocument) render(data any) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.reset()
	switch v := data.(type) {
	case map[string]any:
		t.addObject(0, v)
	case []any:
		t.addArray(0, v)
	default:
		return fmt.Errorf("unrecognized format")
	}
	return nil
}

// addObject adds a JSON object to the tree.
func (t *JSONDocument) addObject(parentID int, data map[string]any) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		v := data[k]
		t.addValue(parentID, k, v)
	}
}

// addArray adds a JSON array to the tree.
func (t *JSONDocument) addArray(parentID int, a []any) {
	for i, v := range a {
		k := fmt.Sprintf("[%d]", i)
		t.addValue(parentID, k, v)
	}
}

// addValue adds a JSON value to the tree.
func (t *JSONDocument) addValue(parentID int, k string, v any) {
	switch v2 := v.(type) {
	case map[string]any:
		id := t.addNode(parentID, k, Empty, Object)
		t.addObject(id, v2)
	case []any:
		id := t.addNode(parentID, k, Empty, Array)
		t.addArray(id, v2)
	case string:
		t.addNode(parentID, k, v2, String)
	case float64:
		t.addNode(parentID, k, v2, Number)
	case bool:
		t.addNode(parentID, k, v2, Boolean)
	case nil:
		t.addNode(parentID, k, v2, Null)
	default:
		t.addNode(parentID, k, v2, Unknown)
	}
}

// addNode adds a node to the tree and returns the UID.
// Nodes will be rendered in the same order they are added.
// Use "" as parentUID for adding nodes at the top level.
// Returns the generated UID for this node and the incremented ID
func (t *JSONDocument) addNode(parentID int, key string, value any, typ JSONType) int {
	if parentID != 0 {
		_, found := t.values[parentID]
		if !found {
			panic(fmt.Sprintf("parent ID does not exist: %d", parentID))
		}
	}
	t.n++
	id := t.n
	_, found := t.values[id]
	if found {
		panic(fmt.Sprintf("UID for this node already exists: %v", id))
	}
	t.ids[parentID] = append(t.ids[parentID], id)
	t.values[id] = Node{Key: key, Value: value, Type: typ}
	t.parents[id] = parentID
	if t.n%progressUpdateTick == 0 {
		p := float64(t.n) / float64(t.elementsCount)
		if err := t.setProgressInfo(ProgressInfo{CurrentStep: 4, Progress: p}); err != nil {
			slog.Warn("Failed to set progress", "err", err)
		}
	}
	return id
}

// reset re-initializes the tree so a new tree can be build.
func (t *JSONDocument) reset() {
	t.values = make(map[int]Node)
	t.ids = make(map[int][]int)
	t.parents = make(map[int]int)
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

func uid2id(uid widget.TreeNodeID) int {
	if uid == "" {
		return 0
	}
	id, err := strconv.Atoi(uid)
	if err != nil {
		panic(err)
	}
	return id
}

func id2uid(id int) widget.TreeNodeID {
	if id == 0 {
		return ""
	}
	return strconv.Itoa(id)
}

func ids2uids(ids []int) []widget.TreeNodeID {
	uids := make([]widget.TreeNodeID, len(ids))
	for i, id := range ids {
		uids[i] = id2uid(id)
	}
	return uids
}
