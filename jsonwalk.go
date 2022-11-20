package jsonwalk

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type NodeValueType int

const (
	Nil     NodeValueType = iota // "null" in JSON terminology.
	Bool                         // Can be type asserted as v.(bool)
	String                       // Can be type asserted as v.(string)
	Float64                      // "number" in JSON terminology. Can be type asserted as v.(float64)
	Array                        // Can be type asserted as v.([]interface{})
	Map                          // "object" in JSON terminology. Can be type asserted as v.(map[string]interface{})
)

type WalkPath interface {
	// Path returns a path leading to the node in a string representation.
	Path() string
	// Level returns path level, starting with 0 for the first element.
	Level() int
	// MapEl constructs a child element with the k key.
	MapEl(k string) walkPath
	// ArrayEl constructs an array child element with index i.
	ArrayEl(i int) walkPath
}

// walkPath is a default WalkPath implementation for Walk function
// which can be overridden with the WalkWith call.
type walkPath struct {
	parent      *walkPath // nil parent for the first node
	portion     string
	preSepErase bool
	level       int // origin has the level of 0
}

func newWalkPath() walkPath {
	return walkPath{} // all default
}

func (w walkPath) Path() string {
	var s string
	var v = &w
	for {
		if v != nil && v.parent != nil {
			var sep = "."
			if v.preSepErase {
				sep = ""
			}
			if v.parent.parent == nil {
				sep = ""
			}
			s = sep + v.portion + s
		} else {
			break
		}
		v = v.parent
	}
	return s
}

func (w walkPath) MapEl(k string) walkPath {
	n := walkPath{
		parent:      &w,
		portion:     k,
		preSepErase: false,
		level:       w.level + 1,
	}
	return n
}

func (w walkPath) ArrayEl(i int) walkPath {
	n := walkPath{
		parent:      &w,
		portion:     "[" + strconv.Itoa(i) + "]",
		preSepErase: true,
		level:       w.level + 1,
	}
	return n
}

func (w walkPath) Level() int {
	return w.level
}

// WalkCallback is an interface with a callback function that is called for both leaf nodes of types
// Nil ("null" in JSON terminology),
// Bool,
// String and
// Float64 ("number" in JSON terminology)
// as well as container nodes of types Array and Map ("object" in JSON terminology).
//
// WalkCallback.C receives the node path in a form of WalkPath, which, if not overridden by WalkWith,
// returns a full non-escaped string-like path, for every discovered Map in a simple "."-separated form and "[int]" form to separate array elements.
//
// No expectations as to the order in which paths based on Map keys are discovered should be made. Array elements arrive in the order
// they are found in JSON. The parent of a node is guaranteed to be discovered before its child.
//
// Print{} returns a simple WalkCallback implementation that prints result of walk.
// NewOutput(io.Writer) gives an option to accept the output destination.
// Callback(c func(path WalkPath, key interface{}, value interface{}, vType NodeValueType)) is a wrapper to pass
// a callback function that returns an object that implements WalkCallback with that function as a delegate.
//
// The key for array elements is of type int, for map it is depending on the key type.
type WalkCallback interface {
	C(path WalkPath, key interface{}, value interface{}, nodeValueType NodeValueType)
}

type f struct {
	f func(path WalkPath, key interface{}, value interface{}, nodeValueType NodeValueType)
}

func (f f) C(path WalkPath, key interface{}, value interface{}, nodeValueType NodeValueType) {
	f.f(path, key, value, nodeValueType)
}

// Callback is a wrapper that accepts a callback function and returns a value that satisfies the WalkCallback interface.
func Callback(c func(path WalkPath, key interface{}, value interface{}, nodeValueType NodeValueType)) WalkCallback {
	return f{c}
}

// Print implements WalkCallback by printing JSON to the os.Stdout by utilizing the NewOutput(os.Stdout).
//
// Passing Print{} to the Walk function is enough to start printing the JSON structure.
type Print struct {
}

func (Print) C(path WalkPath, key interface{}, value interface{}, nodeValueType NodeValueType) {
	NewOutput(os.Stdout).C(path, key, value, nodeValueType)
}

// output implements WalkCallback in a form of a simple tree-like structure
// good enough for debugging. NewOutput allows to specify the output.
type output struct {
	w io.Writer
}

// NewOutput returns an Output object initialized with the io.Writer output.
// It satisfies the WalkCallback interface.
//
// The output is a printout of the JSON structure with the paths between two vertical lines and
// hints on the key and value types. Indentation is squeezed to the left to save space. The actual
// level can be assumed from the amount of separate elements in the path.
//
// The first line of the output gives a hint on the root value type if it's a Map:
//
//	(m)
//	"Actors" |Actors| (a)
//
// or Array:
//
//	(a)
//	0:1 |[0]| (0:f)
//	1:2 |[1]| (1:f)
//
// For single values it's as simple as:
//
//	"abc" (s)
//
// All subsequent lines provide the type at the end of each line. It can also be in the form
// of a (keyType:valueType) for maps and values of maps, or (index:valueType) for arrays:
//
//	"Born At":"New York City, NY" |Actors[1].Born At| (s:s)
//
//	1:"Isabella Jane" |Actors[0].children[1]| (1:s)
//
//	"employees" |employees| (s:a)
func NewOutput(w io.Writer) *output {
	return &output{w: w}
}

func (o *output) C(path WalkPath, key interface{}, value interface{}, nodeValueType NodeValueType) {
	var level = path.Level()
	level = level - 1
	if level < 0 {
		level = 0
	}
	keyType := ""
	if k, ok := key.(int); ok { // Key is of type int for array indices
		keyType = strconv.Itoa(k)
	} else {
		keyType = strings.ToLower(t(key).String()[:1])
	}
	levelStr := ""
	//levelStr = fmt.Sprintf(" #%d", path.Level())
	if nodeValueType == Array || nodeValueType == Map {
		// Not a leaf, only dealing with the key
		if path.Level() == 0 {
			// For root map / array we simply output its type.
			_, _ = fmt.Fprintf(o.w, "(%v)%v\n",
				strings.ToLower(nodeValueType.String()[:1]), levelStr)
		} else {
			var lv = strings.Repeat("  ", level)
			_, _ = fmt.Fprintf(o.w, "%v%#v |%v| (%v:%v)%v\n",
				lv, key, strings.TrimSpace(path.Path()), keyType, strings.ToLower(nodeValueType.String()[:1]), levelStr)
		}
	} else {
		if path.Level() == 0 {
			// For root single value we simply output its value and type.
			_, _ = fmt.Fprintf(o.w, "%#v (%v)%v\n",
				value, strings.ToLower(nodeValueType.String()[:1]), levelStr)
		} else {
			var lv = strings.Repeat("  ", level)
			_, _ = fmt.Fprintf(o.w, "%v%#v:%#v |%v| (%v:%v)%v\n",
				lv, key, value, strings.TrimSpace(path.Path()), keyType, strings.ToLower(nodeValueType.String()[:1]), levelStr)
		}
		return
	}
}

// Walk walks unmarshalled arbitrary JSON with any of the root values.
//
// It calls walk.C for every leaf of types Nil, Bool, String or Float64 as well as every non-leaf node of types Array or Map.
// Callback receives discovered type in a form of NodeValueType for any logic to be performed based on that.
//
// Map keys will arrive in unpredictable order.
//
//	var f interface{}
//	err := json.Unmarshal([]byte(src), &f)
//	if err != nil {
//		return // deal with error
//	}
//	jsonwalk.Walk(&f, jsonwalk.Print{})
//
// For a custom callback, the easiest shortcut is a Callback which accepts a callback function.
//
//	jsonwalk.Walk(&f, jsonwalk.Callback(c func(path WalkPath, key interface{}, value interface{}, nodeValueType NodeValueType) {
//	  ...
//	}))
func Walk(m *interface{}, walk WalkCallback) {
	w(newWalkPath(), nil, m, walk)
}

// WalkWith does the same as Walk except that it accepts starting WalkPath value
// which can be of a different type than the built-in walkPath, allowing for an overriden
// behavior of path construction.
//
// If nil is passed for path, the function falls back to calling Walk.
func WalkWith(path WalkPath, m *interface{}, walk WalkCallback) {
	if path == nil {
		Walk(m, walk)
	} else {
		w(path, nil, m, walk)
	}
}

func t(k interface{}) NodeValueType {
	switch k.(type) {
	case nil:
		return Nil
	case bool:
		return Bool
	case string:
		return String
	case float64:
		return Float64
	case []interface{}:
		return Array
	case map[string]interface{}:
		return Map
	default:
		panic(fmt.Sprintf("unsupported type conversion for %v (%T)", k, k))
	}
}

func w(path WalkPath, k interface{}, v *interface{}, walk WalkCallback) {
	switch vt := (*v).(type) {
	case nil:
		if walk != nil {
			walk.C(path, k, vt, Nil)
		}
	case bool:
		if walk != nil {
			walk.C(path, k, vt, Bool)
		}
	case string:
		if walk != nil {
			walk.C(path, k, vt, String)
		}
	case float64:
		if walk != nil {
			walk.C(path, k, vt, Float64)
		}
	case []interface{}:
		if walk != nil {
			walk.C(path, k, vt, Array)
		}
		arrayWalk(path, &vt, walk)
	case map[string]interface{}:
		if walk != nil {
			walk.C(path, k, vt, Map)
		}
		mapWalk(path, &vt, walk)
	default:
		panic(fmt.Sprintf("%v=%v (unknown type %v)", k, vt, reflect.TypeOf(vt)))
	}
}

func mapWalk(path WalkPath, m *map[string]interface{}, walk WalkCallback) {
	for k, v := range *m {
		w(path.MapEl(k), k, &v, walk)
	}
}

func arrayWalk(path WalkPath, a *[]interface{}, walk WalkCallback) {
	for i, v := range *a {
		w(path.ArrayEl(i), i, &v, walk)
	}
}
