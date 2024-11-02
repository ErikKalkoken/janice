// Package jsondocument contains the logic for rendering a Fyne tree from a JSON document.
package jsondocument

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	// Update progress after x added nodes
	progressUpdateTick = 10_000
	// Total number of load steps
	totalLoadSteps = 3
	// Parent ID of root node
	rootNodeParentID = -1
	// Search target not found
	notFound = -1
)

var ErrCallerCanceled = errors.New("process canceled by caller")
var ErrNotFound = errors.New("not found")

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

// SearchType represents the type of search to perform.
type SearchType uint

const (
	SearchKey SearchType = iota
	SearchString
	SearchNumber
	SearchKeyword
)

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

// Node represents a node in the JSON data tree.
type Node struct {
	Key   string
	Value any
	Type  JSONType
}

// JSONDocument represents a JSON document which can be rendered by a Fyne tree widget.
type JSONDocument struct {
	// How often progress info is updated
	ProgressUpdateTick int32

	progressInfo  binding.Untyped
	elementsCount int

	// ids are stored as int32 to save memory. The API converts them to and from UID strings.
	ids     map[int32][]int32
	values  []Node  // using a slice here instead of a map for better load time
	parents []int32 // ditto
	n       int32
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
	id := uid2id(uid)
	return ids2uids(j.ids[id])
}

// IsBranch reports wether a node is a branch.
// This can be used directly in the tree widget isBranch() function.
func (j *JSONDocument) IsBranch(uid widget.TreeNodeID) bool {
	id := uid2id(uid)
	_, found := j.ids[id]
	return found
}

// Value returns the value of a node
func (j *JSONDocument) Value(uid widget.TreeNodeID) Node {
	id := uid2id(uid)
	return j.values[id]
}

// Load loads JSON data from a reader and builds a new JSON document from it.
// It reports it's current progress to the caller via updates to progressInfo.
// Closes the reader.
func (j *JSONDocument) Load(ctx context.Context, reader fyne.URIReadCloser, progressInfo binding.Untyped) error {
	j.progressInfo = progressInfo
	data, err := j.load(ctx, reader)
	if errors.Is(err, context.Canceled) {
		err = ErrCallerCanceled
	}
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ErrCallerCanceled
	default:
	}
	if err := j.setProgressInfo(ProgressInfo{CurrentStep: 2}); err != nil {
		return err
	}
	sizer := JSONTreeSizer{}
	size, err := sizer.Calculate(data)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ErrCallerCanceled
	default:
	}
	j.elementsCount = size
	slog.Info("Tree size calculated", "size", size)
	if err := j.setProgressInfo(ProgressInfo{CurrentStep: 3}); err != nil {
		return err
	}
	if err := j.render(ctx, data, int32(size)); err != nil {
		return err
	}
	data = nil // GC can free this memory
	slog.Info("Finished loading JSON document into tree", "size", j.n)
	return nil
}

// Size returns the number of nodes.
func (j *JSONDocument) Reset() {
	j.initialize(0)
}

// Parent returns the UID of the parent node.
func (j *JSONDocument) Parent(uid widget.TreeNodeID) widget.TreeNodeID {
	id := uid2id(uid)
	return id2uid(j.parents[id])
}

// Path returns the path of a node in the tree.
func (j *JSONDocument) Path(uid widget.TreeNodeID) []widget.TreeNodeID {
	path := make([]int32, 0)
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
	return int(j.n)
}

// readCloserCtx adds context to a ReadCloser and allows a stream to be canceled.
type readCloserCtx struct {
	ctx context.Context
	r   io.ReadCloser
}

func (r *readCloserCtx) Read(p []byte) (n int, err error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

func (r *readCloserCtx) Close() error {
	return r.r.Close()
}

// newReaderContext returns a new readCloser with a context.
func newReaderContext(ctx context.Context, r io.ReadCloser) io.ReadCloser {
	return &readCloserCtx{ctx: ctx, r: r}
}

func (j *JSONDocument) load(ctx context.Context, reader io.ReadCloser) (any, error) {
	defer reader.Close()
	if err := j.setProgressInfo(ProgressInfo{CurrentStep: 1}); err != nil {
		return nil, err
	}
	var data any
	reader2 := newReaderContext(ctx, reader)
	dec := json.NewDecoder(reader2)
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

// render is the main method for rendering the JSON data into a tree.
func (j *JSONDocument) render(ctx context.Context, data any, size int32) error {
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
func (j *JSONDocument) addObject(ctx context.Context, parentID int32, data map[string]any) error {
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
func (j *JSONDocument) addArray(ctx context.Context, parentID int32, a []any) error {
	var sb strings.Builder
	for i, v := range a {
		sb.Reset()
		sb.WriteByte('[')
		sb.WriteString(strconv.Itoa(i))
		sb.WriteByte(']')
		if err := j.addValue(ctx, parentID, sb.String(), v); err != nil {
			return err
		}
	}
	return nil
}

// addValue adds a JSON value to the tree.
func (j *JSONDocument) addValue(ctx context.Context, parentID int32, k string, v any) error {
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
func (j *JSONDocument) addNode(ctx context.Context, parentID int32, key string, value any, typ JSONType) (int32, error) {
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
	j.values[id] = Node{Key: key, Value: value, Type: typ}
	j.parents[id] = parentID
	if parentID != rootNodeParentID {
		j.ids[parentID] = append(j.ids[parentID], id)
	}
	if j.n%j.ProgressUpdateTick == 0 {
		select {
		case <-ctx.Done():
			return 0, ErrCallerCanceled
		default:
		}
		p := float64(j.n) / float64(j.elementsCount)
		if err := j.setProgressInfo(ProgressInfo{CurrentStep: 3, Progress: p}); err != nil {
			slog.Warn("Failed to set progress", "err", err)
		}
	}
	j.n++
	return id, nil
}

// initialize initializes the tree and allocates needed memory.
//
// A valid tree includes a root node (ID=0) and at least one normal node.
func (j *JSONDocument) initialize(size int32) {
	j.ids = make(map[int32][]int32)
	j.values = make([]Node, size)
	j.parents = make([]int32, size)
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

// Search returns the next node with a matching key or an error if not found or canceled.
// The starting node will be ignored, so that is is possible to find successive nodes with the same key.
// The search direction is from top to bottom.
func (j *JSONDocument) Search(ctx context.Context, uid widget.TreeNodeID, search string, typ SearchType) (widget.TreeNodeID, error) {
	if search == "" {
		return "", ErrNotFound
	}
	id := uid2id(uid)
	startID := id
	pattern, err := regexp.Compile(wildCardToRegexp(search))
	if err != nil {
		return "", err
	}

	for {
		foundID, err := j.searchNode(ctx, id, pattern, typ)
		if err != nil {
			return "", err
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
			select {
			case <-ctx.Done():
				return "", ErrCallerCanceled
			default:
			}
		}
	}
}

func (j *JSONDocument) searchNode(ctx context.Context, id int32, pattern *regexp.Regexp, typ SearchType) (int32, error) {
	n := j.values[id]
	if n.Type == Array || n.Type == Object {
		foundID, err := j.searchContainer(ctx, id, pattern, typ)
		if err != nil {
			return 0, err
		}
		if foundID != notFound {
			return foundID, nil
		}
	}
	switch typ {
	case SearchKey:
		if pattern.MatchString(n.Key) {
			return id, nil
		}
	case SearchKeyword:
		switch n.Type {
		case Boolean:
			if pattern.MatchString(fmt.Sprint(n.Value)) {
				return id, nil
			}
		case Null:
			if pattern.MatchString("null") {
				return id, nil
			}
		default:
			return notFound, nil
		}
	case SearchNumber:
		if n.Type != Number {
			return notFound, nil
		}
		v := n.Value.(float64)
		if pattern.MatchString(strconv.FormatFloat(v, 'f', -1, 64)) {
			return id, nil
		}
	case SearchString:
		if n.Type != String {
			return notFound, nil
		}
		v := n.Value.(string)
		if pattern.MatchString(v) {
			return id, nil
		}
	default:
		panic("Undefined search type")
	}
	return notFound, nil
}

func (j *JSONDocument) searchContainer(ctx context.Context, id int32, pattern *regexp.Regexp, typ SearchType) (int32, error) {
	select {
	case <-ctx.Done():
		return 0, ErrCallerCanceled
	default:
	}
	for _, childID := range j.ids[id] {
		foundID, err := j.searchNode(ctx, childID, pattern, typ)
		if err != nil {
			return 0, err
		}
		if foundID != notFound {
			return foundID, nil
		}
	}
	return notFound, nil
}

func wildCardToRegexp(pattern string) string {
	components := strings.Split(pattern, "*")
	if len(components) == 1 {
		// if len is 1, there are no *'s, return exact match pattern
		return "^" + pattern + "$"
	}
	var result strings.Builder
	for i, literal := range components {

		// Replace * with .*
		if i > 0 {
			result.WriteString(".*")
		}

		// Quote any regular expression meta characters in the
		// literal text.
		result.WriteString(regexp.QuoteMeta(literal))
	}
	return "^" + result.String() + "$"
}

// Extract returns a segment of the JSON document, with the given UID as new root container.
// Note that only arrays and objects can be extracted
func (j *JSONDocument) Extract(uid widget.TreeNodeID) ([]byte, error) {
	var data any
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

func (j *JSONDocument) extractArray(id int32) []any {
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

func (j *JSONDocument) extractObject(id int32) map[string]any {
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

func uid2id(uid widget.TreeNodeID) int32 {
	if uid == "" {
		return 0
	}
	id, err := strconv.Atoi(uid)
	if err != nil {
		panic(err)
	}
	return int32(id)
}

func id2uid(id int32) widget.TreeNodeID {
	if id == 0 {
		return ""
	}
	return strconv.Itoa(int(id))
}

func ids2uids(ids []int32) []widget.TreeNodeID {
	uids := make([]widget.TreeNodeID, len(ids))
	for i, id := range ids {
		uids[i] = id2uid(id)
	}
	return uids
}
