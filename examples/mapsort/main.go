package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/zzwx/jsonwalk"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

//go:embed data.json
var src []byte

func main() {
	var f interface{}
	err := json.Unmarshal([]byte(src), &f)
	if err != nil {
		return // deal with error
	}
	if f == nil {
		return // deal with nil if desired (Walk is a no-op in this case anyway)
	}
	jsonwalk.Walk(f.(map[string]interface{}), jsonwalk.Print())

	// Let's modify each data.<year> value to Fahrenheit.
	jsonwalk.Walk(f.(map[string]interface{}), func(path jsonwalk.WalkPath, key string, value interface{}, vType jsonwalk.NodeValueType) (change bool, newValue interface{}) {
		if path.Level() == 1 && strings.HasPrefix(path.String(), "data.") && vType == jsonwalk.String {
			f, err := strconv.ParseFloat(value.(string), 64)
			if err == nil {
				// In fact we'll return it back as Float64 right away
				return true, float64(f*9.0/5.0) + 32
			}
		}

		return false, nil
	})

	fmt.Println()

	// We know the structure of the "data" path, so we can sort the map as one incoming value of the "data" node.
	jsonwalk.Walk(f.(map[string]interface{}), func(path jsonwalk.WalkPath, key string, value interface{}, vType jsonwalk.NodeValueType) (change bool, newValue interface{}) {
		if path.Level() == 0 && path.String() == "data" && vType == jsonwalk.Map {
			if v, ok := value.(map[string]interface{}); ok {
				keys := maps.Keys(v)
				slices.Sort(keys)
				for _, k := range keys {
					if f, ok := v[k].(float64); ok { // It's already float64 due to previous modification
						fmt.Printf("%v %6.2f°F\n", k, f)
					}
				}

			}
		}
		return false, nil
	})
}
