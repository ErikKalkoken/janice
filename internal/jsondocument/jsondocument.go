// Package jsondocument contains the logic for rendering a Fyne tree from a JSON document.
package jsondocument

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

const (
	// Update progress after x added nodes
	progressUpdateTick = 10_000
	// Total number of load steps
	totalLoadSteps = 4
	// Parent ID of root node
	rootNodeParentID = -1
	// Search target not found
	notFound = -1
)

var ErrLoadCanceled = errors.New("load canceled by caller")
var ErrNotFound = errors.New("key not found")

// JSONType represents the type of a JSON value.
type JSONType uint8

const (
	Undefined JSONType = iota
	Array
	Boolean
	Null
	Number
	Object
	String
	Unknown
)

var typeMap = map[JSONType]string{
	Array:     "array",
	Boolean:   "boolean",
	Null:      "null",
	Number:    "number",
	Object:    "object",
	String:    "string",
	Undefined: "undefined",
	Unknown:   "unknown",
}

func (t JSONType) String() string {
	s, ok := typeMap[t]
	if !ok {
		return typeMap[Undefined]
	}
	return s
}

// Node represents a node in the JSON data tree.
type Node struct {
	UID   string
	Key   string
	Value any
	Type  JSONType
}

// ProgressInfo represents the current progress while loading a document
// and is used to communicate the the UI.
type ProgressInfo struct {
	CurrentStep int
	Progress    float64
	Size        int
	TotalSteps  int
}

// This singleton represents an empty value in a Node.
var Empty = struct{}{}

// JSONDocument represents a JSON document which can be rendered by a Fyne tree widget.
type JSONDocument struct {
	// How often progress info is updated
	ProgressUpdateTick int

	progressInfo  binding.Untyped
	elementsCount int

	mu      sync.RWMutex
	ids     map[int][]int
	values  []Node // using a slice here instead of a map for better load time
	parents []int  // ditto
	n       int
}

// Returns a new JSONDocument object.
func New() *JSONDocument {
	j := &JSONDocument{
		progressInfo:       binding.NewUntyped(),
		ProgressUpdateTick: progressUpdateTick,
	}
	j.initialize(0)
	return j
}

// ChildUIDs returns the child UIDs for a given node.
// This can be used directly in the tree widget childUIDs() function.
func (j *JSONDocument) ChildUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	if !j.mu.TryRLock() {
		// This method can be called by another goroutine from the Fyne library while a new tree is loaded.
		// This can not block, or it would block the whole Fyne app.
		return []widget.TreeNodeID{}
	}
	defer j.mu.RUnlock()
	id := uid2id(uid)
	return ids2uids(j.ids[id])
}

// IsBranch reports wether a node is a branch.
// This can be used directly in the tree widget isBranch() function.
func (j *JSONDocument) IsBranch(uid widget.TreeNodeID) bool {
	if !j.mu.TryRLock() {
		return false
	}
	defer j.mu.RUnlock()
	id := uid2id(uid)
	_, found := j.ids[id]
	return found
}

// Value returns the value of a node
func (j *JSONDocument) Value(uid widget.TreeNodeID) Node {
	if !j.mu.TryRLock() {
		return Node{}
	}
	defer j.mu.RUnlock()
	id := uid2id(uid)
	return j.values[id]
}

// Load loads JSON data from a reader and builds a new JSON document from it.
// It reports it's current progress to the caller via updates to progressInfo.
func (j *JSONDocument) Load(ctx context.Context, reader fyne.URIReadCloser, progressInfo binding.Untyped) error {
	j.progressInfo = progressInfo
	byt, err := j.loadFile(reader)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ErrLoadCanceled
	default:
	}
	data, err := j.parseFile(byt)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ErrLoadCanceled
	default:
	}
	if err := j.setProgressInfo(ProgressInfo{CurrentStep: 3}); err != nil {
		return err
	}
	sizer := JSONTreeSizer{}
	size, err := sizer.Calculate(data)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ErrLoadCanceled
	default:
	}
	j.elementsCount = size
	slog.Info("Tree size calculated", "size", size)
	if err := j.setProgressInfo(ProgressInfo{CurrentStep: 4}); err != nil {
		return err
	}
	if err := j.render(ctx, data, size); err != nil {
		return err
	}
	slog.Info("Finished loading JSON document into tree", "size", j.n)
	return nil
}

// Size returns the number of nodes.
func (j *JSONDocument) Reset() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.initialize(0)
}

// Path returns the path of a node in the tree.
func (j *JSONDocument) Path(uid widget.TreeNodeID) []widget.TreeNodeID {
	path := make([]int, 0)
	if !j.mu.TryRLock() {
		return []widget.TreeNodeID{}
	}
	defer j.mu.RUnlock()
	id := uid2id(uid)
	for {
		id = j.parents[id]
		if id == 0 {
			break
		}
		path = append(path, id)
	}
	slices.Reverse(path)
	return ids2uids(path)
}

// Size returns the number of nodes.
func (j *JSONDocument) Size() int {
	if !j.mu.TryRLock() {
		return 0
	}
	defer j.mu.RUnlock()
	return j.n
}

func (j *JSONDocument) loadFile(reader fyne.URIReadCloser) ([]byte, error) {
	defer reader.Close()
	if err := j.setProgressInfo(ProgressInfo{CurrentStep: 1}); err != nil {
		return nil, err
	}
	dat, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %s", reader.URI(), err)
	}
	slog.Info("Read file", "uri", reader.URI())
	return dat, nil
}

func (j *JSONDocument) parseFile(dat []byte) (any, error) {
	if err := j.setProgressInfo(ProgressInfo{CurrentStep: 2}); err != nil {
		return nil, err
	}
	var data any
	if err := json.Unmarshal(dat, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %s", err)
	}
	slog.Info("Completed un-marshaling data")
	return data, nil
}

// render is the main method for rendering the JSON data into a tree.
func (j *JSONDocument) render(ctx context.Context, data any, size int) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.initialize(size)
	var err error
	switch v := data.(type) {
	case map[string]any:
		if _, err := j.addNode(ctx, -1, "", Empty, Object); err != nil {
			return err
		}
		err = j.addObject(ctx, 0, v)
	case []any:
		if _, err := j.addNode(ctx, -1, "", Empty, Array); err != nil {
			return err
		}
		err = j.addArray(ctx, 0, v)
	default:
		err = fmt.Errorf("unrecognized format")
	}
	return err
}

// addObject adds a JSON object to the tree.
func (j *JSONDocument) addObject(ctx context.Context, parentID int, data map[string]any) error {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		v := data[k]
		if err := j.addValue(ctx, parentID, k, v); err != nil {
			return err
		}
	}
	return nil
}

// addArray adds a JSON array to the tree.
func (j *JSONDocument) addArray(ctx context.Context, parentID int, a []any) error {
	for i, v := range a {
		k := fmt.Sprintf("[%d]", i)
		if err := j.addValue(ctx, parentID, k, v); err != nil {
			return err
		}
	}
	return nil
}

// addValue adds a JSON value to the tree.
func (j *JSONDocument) addValue(ctx context.Context, parentID int, k string, v any) error {
	switch v2 := v.(type) {
	case map[string]any:
		id, err := j.addNode(ctx, parentID, k, Empty, Object)
		if err != nil {
			return err
		}
		if err := j.addObject(ctx, id, v2); err != nil {
			return err
		}
	case []any:
		id, err := j.addNode(ctx, parentID, k, Empty, Array)
		if err != nil {
			return err
		}
		if err := j.addArray(ctx, id, v2); err != nil {
			return err
		}
	case string:
		_, err := j.addNode(ctx, parentID, k, v2, String)
		if err != nil {
			return err
		}
	case float64:
		_, err := j.addNode(ctx, parentID, k, v2, Number)
		if err != nil {
			return err
		}
	case bool:
		_, err := j.addNode(ctx, parentID, k, v2, Boolean)
		if err != nil {
			return err
		}
	case nil:
		_, err := j.addNode(ctx, parentID, k, v2, Null)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unrecognized JSON type %v", v)
	}
	return nil
}

// addNode adds a node to the tree and returns the UID.
// Nodes will be rendered in the same order they are added.
// parentID == -1 denotes the root node
// Returns the generated UID for this node and the incremented ID
func (j *JSONDocument) addNode(ctx context.Context, parentID int, key string, value any, typ JSONType) (int, error) {
	if parentID != rootNodeParentID {
		n := j.values[parentID]
		if n.Type == Undefined {
			return 0, fmt.Errorf("parent ID does not exist: %d", parentID)
		}
	}
	id := j.n
	n := j.values[id]
	if n.Type != Undefined {
		return 0, fmt.Errorf("ID for this node already exists: %v", id)
	}
	j.values[id] = Node{UID: id2uid(id), Key: key, Value: value, Type: typ}
	j.parents[id] = parentID
	if parentID != rootNodeParentID {
		j.ids[parentID] = append(j.ids[parentID], id)
	}
	if j.n%j.ProgressUpdateTick == 0 {
		select {
		case <-ctx.Done():
			return 0, ErrLoadCanceled
		default:
		}
		p := float64(j.n) / float64(j.elementsCount)
		if err := j.setProgressInfo(ProgressInfo{CurrentStep: 4, Progress: p}); err != nil {
			slog.Warn("Failed to set progress", "err", err)
		}
	}
	j.n++
	return id, nil
}

// initialize initializes the tree and allocates needed memory.
//
// A valid tree includes a root node (ID=0) and at least one normal node.
func (j *JSONDocument) initialize(size int) {
	j.ids = make(map[int][]int)
	j.values = make([]Node, size)
	j.parents = make([]int, size)
	j.n = 0
}

func (j *JSONDocument) setProgressInfo(info ProgressInfo) error {
	info.TotalSteps = totalLoadSteps
	info.Size = j.elementsCount
	if err := j.progressInfo.Set(info); err != nil {
		return err
	}
	return nil
}

// SearchKey returns the next node with a matching key or an error if not found.
// The starting node will be ignored, so that is is possible to find successive nodes with the same key.
// The search direction is from top to bottom.
func (j *JSONDocument) SearchKey(uid widget.TreeNodeID, key string) (widget.TreeNodeID, error) {
	var foundID int
	j.mu.RLock()
	defer j.mu.RUnlock()
	id := uid2id(uid)
	startID := id
	n := j.values[id]
	if n.Type != Array && n.Type != Object {
		id = j.parents[id]
	}
	for {
		switch n := j.values[id]; n.Type {
		case Array, Object:
			foundID = j.searchKey(id, key)
		}
		if foundID != startID && foundID != notFound {
			return id2uid(foundID), nil
		}
		for {
			parentID := j.parents[id]
			childIDs := j.ids[parentID]
			idx := slices.Index(childIDs, id)
			if idx < len(childIDs)-1 {
				id = childIDs[idx+1]
				break
			}
			if parentID == rootNodeParentID || j.parents[parentID] == rootNodeParentID {
				return "", ErrNotFound
			}
			id = parentID
		}
	}
}

func (j *JSONDocument) searchKey(id int, key string) int {
	for _, childID := range j.ids[id] {
		n := j.values[childID]
		if n.Key == key {
			return childID
		}
	}
	for _, childID := range j.ids[id] {
		n := j.values[childID]
		switch n.Type {
		case Array, Object:
			if foundID := j.searchKey(childID, key); foundID != notFound {
				return foundID
			}
		}
	}
	return notFound
}

// Extract returns a segment of the JSON document, with the given UID as new root container.
// Note that only arrays and objects can be extracted
func (j *JSONDocument) Extract(uid widget.TreeNodeID) ([]byte, error) {
	var data any
	j.mu.RLock()
	defer j.mu.RUnlock()
	id := uid2id(uid)
	n := j.values[id]
	switch n.Type {
	case Array:
		data = j.extractArray(id)
	case Object:
		data = j.extractObject(id)
	default:
		return nil, fmt.Errorf("can only extract objects and arrays")
	}
	return json.Marshal(data)
}

func (j *JSONDocument) extractArray(id int) []any {
	data := make([]any, len(j.ids[id]))
	for i, childID := range j.ids[id] {
		n := j.values[childID]
		var v any
		switch n.Type {
		case Array:
			v = j.extractArray(childID)
		case Object:
			v = j.extractObject(childID)
		default:
			v = n.Value
		}
		data[i] = v
	}
	return data
}

func (j *JSONDocument) extractObject(id int) map[string]any {
	data := make(map[string]any)
	for _, childID := range j.ids[id] {
		n := j.values[childID]
		var v any
		switch n.Type {
		case Array:
			v = j.extractArray(childID)
		case Object:
			v = j.extractObject(childID)
		default:
			v = n.Value
		}
		data[n.Key] = v
	}
	return data
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
