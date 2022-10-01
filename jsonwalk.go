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
	Child(sub string, parentType NodeValueType) walkPath
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

func (w walkPath) Child(sub string, parentType NodeValueType) walkPath {
	n := walkPath{
		path:  w.path,
		level: w.level + 1,
	}

	if parentType == Map {
		if n.path != "" {
			n.path += "."
		}
	}

	n.path = n.path + sub

	return n
}

func (w walkPath) Level() int {
	return w.level // strings.Count(w.path, ".") + strings.Count(w.path, "[")
}

// WalkCallback is a type of the callback function that is called for both leaf nodes of types
// Nil ("null" in JSON terminology),
// Bool,
// String and
// Float64 ("number" in JSON terminology)
// as well as container nodes of types Array and Map ("object" in JSON terminology).
//
// The callback receives the path in a form of WalkPath, which with its basic String() method returns a full non-escaped
// string-like path, for every discovered Map in a simple `.`-separated form and `[int]` form to separate array elements.
//
// The callback returns (false, nil) when no changes should be made to the value, or (true, <newValue>) when a change should be made to the value.
// The value should be of the same structure as how Go sees JSON.
//
// No expectations as to the order in which paths based on Map keys are discovered should be made. Array elements arrive in the order
// they are found in JSON. The parent of a node is guaranteed to be discovered before its child.
//
// Because of that it might be necessary to start with a separate Walk to come to needed conclusions before changing the nodes.
//
// Print() returns a simple WalkCallback that prints result of walk. Output(io.Writer) gives an option to accept the output destination.
type WalkCallback func(path WalkPath, key string, value interface{},
	vType NodeValueType) (change bool, newValue interface{})

// Print is a wrapper around Output(os.Stdout).
func Print() WalkCallback {
	return Output(os.Stdout)
}

// Output returns a WalkCallback that outputs to w the results of Walk in a form of a simple tree-like structure
// good enough for debugging.
func Output(w io.Writer) WalkCallback {
	return func(path WalkPath, key string, value interface{}, vType NodeValueType) (change bool, newValue interface{}) {
		if vType == Array || vType == Map {
			// Not a leaf, only dealing with the key
			_, _ = fmt.Fprintf(w, "%v%v - %v - %v\n",
				strings.Repeat("  ", path.Level()), key, vType, path)
			return false, 0
		} else {
			q := ""
			if vType == String {
				q = "\""
			}
			_, _ = fmt.Fprintf(w, "%v%v:%v%v%v - %v - %v\n",
				strings.Repeat("  ", path.Level()), key, q, value, q, vType, path)
			return false, 0
		}
	}
}

// Walk walks unmarshalled arbitrary JSON, asserted as map[string]interface{}.
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
//	jsonwalk.Walk(f.(map[string]interface{}), jsonwalk.Print())
func Walk(m map[string]interface{}, valueCallback WalkCallback) {
	mapWalk(walkPath{}, m, valueCallback)
}

// WalkWith does the same as Walk except that it accepts starting WalkPath value
// which can be of a different type than the built-in walkPath,
// allowing for additional flexibility.
//
// If nil is passed for path, the function falls back to calling Walk.
func WalkWith(path WalkPath, m map[string]interface{}, valueCallback WalkCallback) {
	if path == nil {
		Walk(m, valueCallback)
	} else {
		mapWalk(path, m, valueCallback)
	}
}

func mapWalk(path WalkPath, m map[string]interface{}, valueCallback WalkCallback) {
	for k, v := range m {
		child := path.Child(k, Map)
		switch vt := v.(type) {
		case nil:
			if valueCallback != nil {
				if ch, new_ := valueCallback(child, k, vt, Nil); ch {
					m[k] = new_
				}
			}
		case bool:
			if valueCallback != nil {
				if ch, new_ := valueCallback(child, k, vt, Bool); ch {
					m[k] = new_
				}
			}
		case string:
			if valueCallback != nil {
				if ch, new_ := valueCallback(child, k, vt, String); ch {
					m[k] = new_
				}
			}
		case float64:
			if valueCallback != nil {
				if ch, new_ := valueCallback(child, k, vt, Float64); ch {
					m[k] = new_
				}
			}
		case []interface{}:
			if valueCallback != nil {
				if ch, new_ := valueCallback(child, k, vt, Array); ch {
					m[k] = new_
				}
			}
			arrayWalk(child, vt, valueCallback)
		case map[string]interface{}:
			if valueCallback != nil {
				if ch, new_ := valueCallback(child, k, vt, Map); ch {
					m[k] = new_
				}
			}
			mapWalk(child, vt, valueCallback)
		default:
			panic(fmt.Sprintf("%v=%v (unknown type %v) (%v)", k, vt, reflect.TypeOf(vt), child))
		}
	}
}

// arrayWalk is decoupled to support array-in-array recursion.
func arrayWalk(path WalkPath, a []interface{}, valueCallback WalkCallback) {
	for i, u := range a {
		child := path.Child("["+strconv.Itoa(i)+"]", Array)
		switch vt := u.(type) {
		case nil:
			if valueCallback != nil {
				if ch, new_ := valueCallback(child, strconv.Itoa(i), vt, Nil); ch {
					a[i] = new_
				}
			}
		case bool:
			if valueCallback != nil {
				if ch, new_ := valueCallback(child, strconv.Itoa(i), vt, Bool); ch {
					a[i] = new_
				}
			}
		case string:
			if valueCallback != nil {
				if ch, new_ := valueCallback(child, strconv.Itoa(i), vt, String); ch {
					a[i] = new_
				}
			}
		case float64:
			if valueCallback != nil {
				if ch, new_ := valueCallback(child, strconv.Itoa(i), vt, Float64); ch {
					a[i] = new_
				}
			}
		case []interface{}:
			if valueCallback != nil {
				if ch, new_ := valueCallback(child, strconv.Itoa(i), vt, Array); ch {
					a[i] = new_
				}
			}
			arrayWalk(child, vt, valueCallback)
		case map[string]interface{}:
			if valueCallback != nil {
				if ch, new_ := valueCallback(child, strconv.Itoa(i), vt, Map); ch {
					a[i] = new_
				}
			}
			mapWalk(child, vt, valueCallback)
		default:
			panic(fmt.Sprintf("%v (unknown type %v) (%v)", vt, reflect.TypeOf(vt), child))
		}
	}

}
