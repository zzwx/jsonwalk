// An example of analyzing an embedded JSON data.
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
	jsonwalk.Walk(&f, jsonwalk.Print{})

	years := make(map[string]float64)

	// Collect data.<year> value
	jsonwalk.Walk(&f, jsonwalk.Callback(func(path jsonwalk.WalkPath, key interface{}, value interface{}, tp jsonwalk.NodeValueType) {
		if path.Level() == 2 && strings.HasPrefix(path.Path(), "data.") && tp == jsonwalk.String {
			f, err := strconv.ParseFloat(value.(string), 64)
			if err == nil {
				if k, ok := key.(string); ok {
					years[k] = f
				}
			}
		}
	}))

	keys := maps.Keys(years)
	slices.Sort(keys)

	cnt := 100
	if len(keys) < cnt {
		return
	}

	// average first twenty
	first := keys[0:cnt]
	sum := 0.0
	for _, k := range first {
		sum += toF(years[k])
	}
	av := sum / float64(len(first))

	fmt.Println("-year-|-temperature-")
	for _, k := range keys {
		v := toF(years[k])
		diff := fmt.Sprintf("%+6.2f%%", pDiff(av, v))
		if slices.Index(first, k) >= 0 {
			diff += " *"
		}

		fmt.Printf(" %v  %6.2f°F %v\n", k, v, diff)
	}

	fmt.Printf("Percentage shown is the difference between each line and the avarage of\n")
	fmt.Printf("the oldest available %d years [%v..%v] (%.2f) that are marked with *\n", len(first), keys[0], keys[len(first)-1], av)
}

func toF(c float64) float64 {
	return float64(c*9.0/5.0) + 32
}

func pDiff(old, new float64) float64 {
	return 100 * (new - old) / old
}
