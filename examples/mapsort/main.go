package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/zzwx/jsonwalk"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var src = `
{"description":{"title":"Global Land and Ocean Temperature Anomalies, January-December","units":"Degrees Celsius","base_period":"1901-2000","missing":-999},"data":{"1880":"-0.11","1881":"-0.07","1882":"-0.09","1883":"-0.17","1884":"-0.25","1885":"-0.24","1886":"-0.23","1887":"-0.28","1888":"-0.12","1889":"-0.08","1890":"-0.33","1891":"-0.25","1892":"-0.30","1893":"-0.32","1894":"-0.31","1895":"-0.24","1896":"-0.09","1897":"-0.09","1898":"-0.27","1899":"-0.16","1900":"-0.07","1901":"-0.15","1902":"-0.26","1903":"-0.38","1904":"-0.45","1905":"-0.28","1906":"-0.21","1907":"-0.38","1908":"-0.43","1909":"-0.44","1910":"-0.40","1911":"-0.44","1912":"-0.34","1913":"-0.32","1914":"-0.14","1915":"-0.09","1916":"-0.32","1917":"-0.40","1918":"-0.30","1919":"-0.25","1920":"-0.23","1921":"-0.16","1922":"-0.25","1923":"-0.25","1924":"-0.24","1925":"-0.18","1926":"-0.08","1927":"-0.17","1928":"-0.18","1929":"-0.33","1930":"-0.11","1931":"-0.06","1932":"-0.13","1933":"-0.26","1934":"-0.11","1935":"-0.16","1936":"-0.12","1937":"-0.01","1938":"-0.02","1939":"0.01","1940":"0.16","1941":"0.27","1942":"0.11","1943":"0.11","1944":"0.28","1945":"0.18","1946":"-0.01","1947":"-0.03","1948":"-0.05","1949":"-0.07","1950":"-0.15","1951":"0.00","1952":"0.05","1953":"0.13","1954":"-0.10","1955":"-0.13","1956":"-0.18","1957":"0.07","1958":"0.12","1959":"0.08","1960":"0.05","1961":"0.09","1962":"0.11","1963":"0.12","1964":"-0.14","1965":"-0.07","1966":"-0.01","1967":"0.00","1968":"-0.03","1969":"0.11","1970":"0.06","1971":"-0.07","1972":"0.04","1973":"0.19","1974":"-0.06","1975":"0.01","1976":"-0.07","1977":"0.21","1978":"0.12","1979":"0.23","1980":"0.28","1981":"0.32","1982":"0.19","1983":"0.36","1984":"0.17","1985":"0.16","1986":"0.24","1987":"0.38","1988":"0.39","1989":"0.29","1990":"0.45","1991":"0.39","1992":"0.24","1993":"0.28","1994":"0.34","1995":"0.47","1996":"0.32","1997":"0.51","1998":"0.65","1999":"0.44","2000":"0.42","2001":"0.57","2002":"0.62","2003":"0.64","2004":"0.58","2005":"0.67","2006":"0.63","2007":"0.62","2008":"0.54","2009":"0.64","2010":"0.72","2011":"0.57","2012":"0.63","2013":"0.67","2014":"0.74","2015":"0.93","2016":"0.99","2017":"0.90","2018":"0.82","2019":"0.94","2020":"0.97","2021":"0.84"}}`

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
		if strings.HasPrefix(path.String(), "data.") && vType == jsonwalk.String {
			f, err := strconv.ParseFloat(value.(string), 10)
			if err == nil {
				// In fact we'll return it back as Float64 right away
				return true, float64(f*9.0/5.0) + 32
			}
		}

		return false, nil
	})

	fmt.Println("---")

	// We know the structure of the "data" path, so we can sort the map as incoming value.
	jsonwalk.Walk(f.(map[string]interface{}), func(path jsonwalk.WalkPath, key string, value interface{}, vType jsonwalk.NodeValueType) (change bool, newValue interface{}) {
		if path.String() == "data" && vType == jsonwalk.Map {
			if v, ok := value.(map[string]interface{}); ok {
				keys := maps.Keys(v)
				slices.Sort(keys)
				for _, k := range keys {
					if f, ok := v[k].(float64); ok { // It's already float64 due to previous modification
						fmt.Printf("%v %6.2f℉\n", k, f)
					}
				}

			}
		}
		return false, nil
	})
}