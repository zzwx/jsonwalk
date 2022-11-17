[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/zzwx/jsonwalk)

`jsonwalk.Walk` walks arbitrary JSON nodes, unmarshalled with the standard library `json.Unmarshall` call, which can be any single value supported by JSON: null, bool, string, number, array or object. It aims a quick analysis of the input and extracting of the needed data.

Internally these types are mapped as following:

* `Nil` - `null` in JSON terminology
* `Bool` - `boolean`
* `String` - `string`
* `Float64` - `number`
* `Array` - `array`
* `Map` - `object`

For every discovered node it calls a provided in the `Walk` callback for every leaf or non-leaf node (`Array` and `Map`).

The callback receives the discovered node type in a form of `jsonvalue.NodeValueType` for any logic to be preformed based on the already known type assertion.

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
jsonwalk.Walk(&f, jsonwalk.Print())
```

This buil-in `Print` method, returning an implementation of `WalkCallback`. To quickly provide your own, utilized the `Callback` wrapper that accepts the callback function. 

Look into `examples` folder for full examples.
