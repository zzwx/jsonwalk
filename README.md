[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/zzwx/jsonwalk)

`jsonwalk.Walk` walks arbitrary JSON nodes, unmarshalled with the standard library `json.Unmarshall` call, asserted as `map[string]interface{}`.

For every discovered node it calls a callback. That includes every leaf node of JSON value types (`Nil` (`null` in JSON terminology), `Bool`, `Float64` (`number` in JSON terminology) and `String`) as well as every non-leaf node (`Array` and `Map` (`object` in JSON terminology)). The callback can return `true, <newValue>` in case a patch of the node is desired.

Callback receives the discovered node type in a form of `jsonvalue.NodeValueType` for any logic to be preformed based on the already known type assertion.

Map keys will be discovered in an unpredictable order so if any action is dependant on the order of such values, it should be made in a separate Walk.

Quick example of printing a JSON structure with values:

```go
var f interface{}
err := json.Unmarshal([]byte(src), &f)
if err != nil {
	return // deal with error
}
if f == nil {
	return // deal with nil if desired (Walk is a no-op in this case anyway)
}
jsonwalk.Walk(f.(map[string]interface{}), jsonwalk.Print())
```

`examples` folder contains a sample of how to modify the node values.
