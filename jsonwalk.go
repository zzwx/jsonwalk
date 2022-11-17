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
	fmt.Stringer
	Child(pathAppend string, parentType NodeValueType) walkPath
	// Level returns path level, starting with 0 for the first element of the root JSON map.
	Level() int
}

// walkPath is a default WalkPath implementation for Walk function
// which can be overridden with the WalkWith call.
type walkPath struct {
	path  string
	level int
}

func (w walkPath) String() string {
	return w.path
}

func (w walkPath) Child(pathAppend string, parentType NodeValueType) walkPath {
	n := walkPath{
		path:  w.path,
		level: w.level + 1,
	}

	if parentType == Map {
		if n.path != "" {
			n.path += "."
		}
	}

	n.path = n.path + pathAppend

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
// The callback receives the path in a form of WalkPath, which, if not overridden by WalkWith, returns a full non-escaped
// string-like path, for every discovered Map in a simple `.`-separated form and `[int]` form to separate array elements.
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
	C(path WalkPath, key interface{}, value interface{}, vType NodeValueType)
}

type f struct {
	f func(path WalkPath, key interface{}, value interface{}, vType NodeValueType)
}

func (f f) C(path WalkPath, key interface{}, value interface{}, vType NodeValueType) {
	f.f(path, key, value, vType)
}

// Callback is a wrapper that accepts a callback function and returns a value that satisfies the WalkCallback interface.
func Callback(c func(path WalkPath, key interface{}, value interface{}, vType NodeValueType)) WalkCallback {
	return f{c}
}

// Print implements WalkCallback by printing JSON to the os.Stdout by utilizing the NewOutput(os.Stdout).
//
// Passing Print{} to the Walk function is enough to start printing the JSON structure.
type Print struct {
}

func (Print) C(path WalkPath, key interface{}, value interface{}, vType NodeValueType) {
	NewOutput(os.Stdout).C(path, key, value, vType)
}

// Output implements WalkCallback in a form of a simple tree-like structure
// good enough for debugging. NewOutput allows to specify the output.
type Output struct {
	w io.Writer
}

// NewOutput returns an Output object initialized with the io.Writer output.
// It satisfies the WalkCallback interface.
//
// The output is a printout of the JSON structure with the paths between two vertical lines and
// hints on the key and value types.
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
// of a (key:value) string:
//
//	"Born At":"New York City, NY" |Actors[1].Born At| (s:s)
//
// This way it hints on the types of both the key and value.
//
// For the array elements, the first will be represented as an integer, actual index:
//
//	1:"Isabella Jane" |Actors[0].children[1]| (1:s)
func NewOutput(w io.Writer) *Output {
	return &Output{w: w}
}

func (o *Output) C(path WalkPath, key interface{}, value interface{}, vType NodeValueType) {
	if vType == Array || vType == Map {
		// Not a leaf, only dealing with the key
		if path.Level() >= 0 {
			var lv string
			lv = strings.Repeat("  ", path.Level())
			_, _ = fmt.Fprintf(o.w, "%v%#v |%v| (%v)\n",
				lv, key, strings.TrimSpace(path.String()), strings.ToLower(vType.String()[:1]))
		} else {
			// For root map / array we simply output its type.
			_, _ = fmt.Fprintf(o.w, "(%v)\n",
				strings.ToLower(vType.String()[:1]))
		}
	} else {
		if path.Level() >= 0 {
			var lv = strings.Repeat("  ", path.Level())
			keyType := ""
			if k, ok := key.(int); ok {
				keyType = strconv.Itoa(k)
			} else {
				keyType = strings.ToLower(t(key).String()[:1])
			}

			_, _ = fmt.Fprintf(o.w, "%v%#v:%#v |%v| (%v:%v)\n",
				lv, key, value, strings.TrimSpace(path.String()), keyType, strings.ToLower(vType.String()[:1]))
		} else {
			// For root single value we simply output its value and type.
			_, _ = fmt.Fprintf(o.w, "%#v (%v)\n",
				value, strings.ToLower(vType.String()[:1]))
		}
		return
	}
}

// Walk walks unmarshalled arbitrary JSON with any of the root values.
//
// It calls valueCallback for every leaf of types Nil, Bool, String or Float64 as well as every non-leaf node of types Array or Map.
// Callback receives discovered type in a form of NodeValueType for any logic to be performed based on that.
//
// Map keys will arrive in unpredictable order.
//
//	var f interface{}
//	err := json.Unmarshal([]byte(src), &f)
//	if err != nil {
//		return // deal with error
//	}
//	if f == nil {
//		return // deal with nil if desired (Walk is a no-op in this case anyway)
//	}
//	jsonwalk.Walk(&f, jsonwalk.Print)
//
// For a custom callback, the easiest shortcut is a Callback which accepts a callback function.
//
//	jsonwalk.Callback(c func(path WalkPath, key interface{}, value interface{}, vType NodeValueType))
func Walk(m *interface{}, valueCallback WalkCallback) {
	w(walkPath{level: -1}, nil, m, valueCallback)
}

// WalkWith does the same as Walk except that it accepts starting WalkPath value
// which can be of a different type than the built-in walkPath,
// allowing for additional flexibility.
//
// If nil is passed for path, the function falls back to calling Walk.
func WalkWith(path WalkPath, m *interface{}, valueCallback WalkCallback) {
	if path == nil {
		Walk(m, valueCallback)
	} else {
		w(path, nil, m, valueCallback)
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

func w(path WalkPath, k interface{}, v *interface{}, valueCallback WalkCallback) {
	switch vt := (*v).(type) {
	case nil:
		if valueCallback != nil {
			valueCallback.C(path, k, vt, Nil)
		}
	case bool:
		if valueCallback != nil {
			valueCallback.C(path, k, vt, Bool)
		}
	case string:
		if valueCallback != nil {
			valueCallback.C(path, k, vt, String)
		}
	case float64:
		if valueCallback != nil {
			valueCallback.C(path, k, vt, Float64)
		}
	case []interface{}:
		if valueCallback != nil {
			valueCallback.C(path, k, vt, Array)
		}
		arrayWalk(path, &vt, valueCallback)
	case map[string]interface{}:
		if valueCallback != nil {
			valueCallback.C(path, k, vt, Map)
		}
		mapWalk(path, &vt, valueCallback)
	default:
		panic(fmt.Sprintf("%v=%v (unknown type %v)", k, vt, reflect.TypeOf(vt)))
	}
}

func mapWalk(path WalkPath, m *map[string]interface{}, valueCallback WalkCallback) {
	for k, v := range *m {
		w(path.Child(k, Map), k, &v, valueCallback)
	}
}

// arrayWalk is decoupled to support array-in-array recursion.
func arrayWalk(path WalkPath, a *[]interface{}, valueCallback WalkCallback) {
	for i, u := range *a {
		w(path.Child("["+strconv.Itoa(i)+"]", Array), i, &u, valueCallback)
	}
}
