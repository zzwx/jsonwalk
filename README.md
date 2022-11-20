[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/zzwx/jsonwalk)

`jsonwalk.Walk` walks arbitrary JSON nodes, unmarshalled with the standard library `json.Unmarshall` call, which can be any single value supported by JSON: null, bool, string, number, array or object. It aims a quick analysis of the input and extracting of the needed data.

> For a library that implements JSON searching & JSON modification, consider [gjson](https://github.com/tidwall/gjson) and [sjson](https://github.com/tidwall/sjson).

Internally the JSON types are mapped as following:

* `Nil` - `null` in JSON terminology
* `Bool` - `boolean`
* `String` - `string`
* `Float64` - `number`
* `Array` - `array`
* `Map` - `object`

For every discovered node it calls provided callback, which is accepted in a form of the `WalkCallback` interface.

The callback receives the discovered key, value and node type as `jsonwalk.NodeValueType` for any logic to be preformed based on the already known type assertion.

Map keys, as always, will be discovered in an unpredictable order so if any action depends on the order of such values, it should be made in a separate `Walk`.

Quick example of printing a JSON structure with values:

```go
var f interface{}
err := json.Unmarshal([]byte(src), &f)
if err != nil {
	return // deal with error
}
jsonwalk.Walk(&f, jsonwalk.Print{})
```

This built-in `Print{}` struct returns an implementation of the `WalkCallback`. To quickly provide a custom callback there's a `Callback` wrapper that accepts the callback function. 

Look into `examples` folder for inspiration.
