package jsonwalk_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/zzwx/jsonwalk"

	"golang.org/x/exp/slices"
)

func TestWalk(t *testing.T) {
	src := `{
		"Actors": [
			{
				"name": "Tom Cruise",
				"age": 56,
				"Born At": "Syracuse, NY",
				"Birthdate": "July 3, 1962",
				"wife": null,
				"weight": 67.5,
				"hasChildren": true,
				"hasGreyHair": false,
				"children": [
					"Suri",
					"Isabella Jane",
					"Connor"
				]
			},
			{
				"name": "Robert Downey Jr.",
				"age": 53,
				"Born At": "New York City, NY",
				"Birthdate": "April 4, 1965",
				"wife": "Susan Downey",
				"weight": 77.1,
				"hasChildren": true,
				"hasGreyHair": false,
				"children": [
					"Indio Falconer",
					"Avri Roel",
					"Exton Elias"
				]
			}
		]
	}`

	/*
	 Actors - Array - Actors
	 	0 - Map - Actors[0]
	 		name:"Tom Cruise" - String - Actors[0].name
	 		hasChildren:true - Bool - Actors[0].hasChildren
	 		hasGreyHair:false - Bool - Actors[0].hasGreyHair
	 		children - Array - Actors[0].children
	 			0:"Suri" - String - Actors[0].children[0]
	 			1:"Isabella Jane" - String - Actors[0].children[1]
	 			2:"Connor" - String - Actors[0].children[2]
	 		age:56 - Float64 - Actors[0].age
	 		Born At:"Syracuse, NY" - String - Actors[0].Born At
	 		Birthdate:"July 3, 1962" - String - Actors[0].Birthdate
	 		photo:"https://jsonformatter.org/img/tom-cruise.jpg" - String - Actors[0].photo
	 		wife:<nil> - Nil - Actors[0].wife
	 		weight:67.5 - Float64 - Actors[0].weight
	 	1 - Map - Actors[1]
	 		age:53 - Float64 - Actors[1].age
	 		Birthdate:"April 4, 1965" - String - Actors[1].Birthdate
	 		photo:"https://jsonformatter.org/img/Robert-Downey-Jr.jpg" - String - Actors[1].photo
	 		wife:"Susan Downey" - String - Actors[1].wife
	 		children - Array - Actors[1].children
	 			0:"Indio Falconer" - String - Actors[1].children[0]
	 			1:"Avri Roel" - String - Actors[1].children[1]
	 			2:"Exton Elias" - String - Actors[1].children[2]
	 		name:"Robert Downey Jr." - String - Actors[1].name
	 		Born At:"New York City, NY" - String - Actors[1].Born At
	 		weight:77.1 - Float64 - Actors[1].weight
	 		hasChildren:true - Bool - Actors[1].hasChildren
	 		hasGreyHair:false - Bool - Actors[1].hasGreyHair
	*/

	var f interface{}

	err := json.Unmarshal([]byte(src), &f)
	if err != nil {
		t.Errorf("error umarshalling json: %v", err)
		return
	}
	m := f.(map[string]interface{})

	foundCnt := 0
	jsonwalk.Walk(m, func(path jsonwalk.WalkPath, key string, value interface{}, vType jsonwalk.NodeValueType) (change bool, newValue interface{}) {
		if path.String() == "Actors" {
			if path.Level() != 0 {
				t.Errorf("expected level 0, got %v", path.Level())
			}
		}
		if path.String() == "Actors[0].name" {
			v := value.(string)
			if v == "Tom Cruise" {
				foundCnt++
			} else {
				t.Errorf("invalid value for %v: %v", path, v)
			}
			if path.Level() != 2 {
				t.Errorf("expected level 2, got %v", path.Level())
			}
		} else if path.String() == "Actors[0].wife" {
			if value == nil && vType == jsonwalk.Nil {
				foundCnt++
			} else {
				t.Errorf("invalid value for %v: %v", path, value)
			}
			if path.Level() != 2 {
				t.Errorf("expected level 2, got %v", path.Level())
			}
		} else if path.String() == "Actors[0].children" {
			v := value.([]interface{})
			var asStr []string

			for _, vv := range v {
				asStr = append(asStr, vv.(string))
			}

			if slices.Compare(asStr, []string{"Suri",
				"Isabella Jane",
				"Connor"}) == 0 {
				foundCnt++
			} else {
				t.Errorf("invalid value for %v: %v", path, v)
			}
			if path.Level() != 2 {
				t.Errorf("expected level 2, got %v", path.Level())
			}
		} else if path.String() == "Actors[1].name" {
			v := value.(string)
			if v == "Robert Downey Jr." {
				foundCnt++
			} else {
				t.Errorf("invalid value for %v: %v", path, v)
			}
			if path.Level() != 2 {
				t.Errorf("expected level 2, got %v", path.Level())
			}
		}
		return false, nil
	})

	expectedFoundCnt := 4

	if foundCnt != expectedFoundCnt {
		t.Errorf("supporsed to find %v paths, got %v", expectedFoundCnt, foundCnt)
	}

}

func TestWalkPrint(t *testing.T) {
	// https://awesomeopensource.com/project/jdorfman/awesome-json-datasets

	fake := []string{`{
		"Actors": [
			{
				"name": "Tom Cruise",
				"age": 56,
				"Born At": "Syracuse, NY",
				"Birthdate": "July 3, 1962",
				"wife": null,
				"weight": 67.5,
				"hasChildren": true,
				"hasGreyHair": false,
				"children": [
					"Suri",
					"Isabella Jane",
					"Connor"
				]
			},
			{
				"name": "Robert Downey Jr.",
				"age": 53,
				"Born At": "New York City, NY",
				"Birthdate": "April 4, 1965",
				"wife": "Susan Downey",
				"weight": 77.1,
				"hasChildren": true,
				"hasGreyHair": false,
				"children": [
					"Indio Falconer",
					"Avri Roel",
					"Exton Elias"
				]
			}
		]
	}`,
		`{
		"name" : "blogger",
		"users" : [
			[ "admins", "1", "2" , 3],
			[ "editors", "4", "5" , "6"]
		]}`, `{
		"name" : "Admin",
		"age" : 36,
		"rights" : [ "admin", "editor", "contributor" ],
		"and": true}`, `{
	"employees": [
		{
			"id": 1,
			"name": "Admin",
			"location": "India"
		},
		{
			"id": 2,
			"name": "Author",
			"location": "USA"
		},
		{
			"id": 3,
			"name": "Visitor",
			"location": "USA"
		}
	]
}`,
		`{"provider":"https://www.exchangerate-api.com","WARNING_UPGRADE_TO_V6":"https://www.exchangerate-api.com/docs/free","terms":"https://www.exchangerate-api.com/terms","base":"USD","date":"2022-09-20","time_last_updated":1663632002,"rates":{"USD":1,"AED":3.67,"AFN":87.36,"ALL":116.86,"AMD":417.16,"ANG":1.79,"AOA":431.56,"ARS":143.34,"AUD":1.49,"AWG":1.79,"AZN":1.7,"BAM":1.95,"BBD":2,"BDT":103.14,"BGN":1.95,"BHD":0.376,"BIF":2031.09,"BMD":1,"BND":1.41,"BOB":6.89,"BRL":5.25,"BSD":1,"BTN":79.6,"BWP":13.18,"BYN":2.54,"BZD":2,"CAD":1.33,"CDF":2050.28,"CHF":0.965,"CLP":922.43,"CNY":7,"COP":4431.05,"CRC":630.27,"CUP":24,"CVE":110.15,"CZK":24.51,"DJF":177.72,"DKK":7.45,"DOP":52.45,"DZD":140.6,"EGP":19.4,"ERN":15,"ETB":52.67,"EUR":0.999,"FJD":2.24,"FKP":0.876,"FOK":7.45,"GBP":0.876,"GEL":2.84,"GGP":0.876,"GHS":10.42,"GIP":0.876,"GMD":55.41,"GNF":8596.13,"GTQ":7.77,"GYD":207.83,"HKD":7.85,"HNL":24.35,"HRK":7.53,"HTG":117.38,"HUF":399.61,"IDR":14947.74,"ILS":3.45,"IMP":0.876,"INR":79.61,"IQD":1449.95,"IRR":41908.53,"ISK":140.02,"JEP":0.876,"JMD":151.94,"JOD":0.709,"JPY":143.27,"KES":120.88,"KGS":81.68,"KHR":4081.29,"KID":1.49,"KMF":491.44,"KRW":1389.75,"KWD":0.3,"KYD":0.833,"KZT":477.82,"LAK":17898.64,"LBP":1507.5,"LKR":357.12,"LRD":153.85,"LSL":17.7,"LYD":4.89,"MAD":10.65,"MDL":19.36,"MGA":4126.46,"MKD":61.67,"MMK":2737.05,"MNT":3219.13,"MOP":8.09,"MRU":37.55,"MUR":44.72,"MVR":15.41,"MWK":1035.73,"MXN":19.98,"MYR":4.53,"MZN":64.59,"NAD":17.7,"NGN":426.91,"NIO":35.52,"NOK":10.24,"NPR":127.37,"NZD":1.68,"OMR":0.384,"PAB":1,"PEN":3.87,"PGK":3.52,"PHP":57.34,"PKR":237.34,"PLN":4.7,"PYG":6973.2,"QAR":3.64,"RON":4.92,"RSD":117.42,"RUB":60.32,"RWF":1073.74,"SAR":3.75,"SBD":8,"SCR":12.86,"SDG":566.69,"SEK":10.79,"SGD":1.41,"SHP":0.876,"SLE":14.35,"SLL":14353.17,"SOS":564.73,"SRD":26.84,"SSP":645.38,"STN":24.47,"SYP":2507.93,"SZL":17.7,"THB":36.97,"TJS":10.2,"TMT":3.5,"TND":2.94,"TOP":2.37,"TRY":18.3,"TTD":6.77,"TVD":1.49,"TWD":31.29,"TZS":2332.28,"UAH":37.16,"UGX":3818.26,"UYU":40.67,"UZS":10980.6,"VES":8.03,"VND":23678.92,"VUV":117.43,"WST":2.65,"XAF":655.25,"XCD":2.7,"XDR":0.772,"XOF":655.25,"XPF":119.2,"YER":250.13,"ZAR":17.7,"ZMW":15.69,"ZWL":603.91}}`,
		`{
  "meta": {
    "disclaimer": "Do not rely on openFDA to make decisions regarding medical care. While we make every effort to ensure that data is accurate, you should assume all results are unvalidated. We may limit or otherwise restrict your access to the API in line with our Terms of Service.",
    "terms": "https://open.fda.gov/terms/",
    "license": "https://open.fda.gov/license/",
    "last_updated": "2022-09-14",
    "results": {
      "skip": 0,
      "limit": 1,
      "total": 22969
    }
  },
  "results": [
    {
      "country": "United States",
      "city": "Davie",
      "address_1": "4131 SW 47th Ave Ste 1403",
      "reason_for_recall": "Recall initiated as a precautionary measure due to potential risk of product contamination with Burkholderia cepacia.",
      "address_2": "",
      "product_quantity": "1,990 bottles",
      "code_info": "UPC No. 632687615989; Lot No. 30661601, Exp. Date 05/2018.",
      "center_classification_date": "20161025",
      "distribution_pattern": "FL, MI, MS, and OH.",
      "state": "FL",
      "product_description": "CytoDetox, Hydrolyzed Clinoptilolite Fragments, 1 oz./30 mL, OTC Non-Sterile.  Dietary supplement.",
      "report_date": "20161102",
      "classification": "Class II",
      "openfda": {},
      "recalling_firm": "Pharmatech LLC",
      "recall_number": "F-0276-2017",
      "initial_firm_notification": "Letter",
      "product_type": "Food",
      "event_id": "75272",
      "more_code_info": "",
      "recall_initiation_date": "20160808",
      "postal_code": "33314-4036",
      "voluntary_mandated": "Voluntary: Firm initiated",
      "status": "Ongoing"
    }
  ]
}`,
		`{"type":"FeatureCollection","metadata":{"generated":1663717668000,"url":"https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_hour.geojson","title":"USGS All Earthquakes, Past Hour","status":200,"api":"1.10.3","count":4},"features":[{"type":"Feature","properties":{"mag":0.4,"place":"Southern Alaska","time":1663716691293,"updated":1663716795644,"tz":null,"url":"https://earthquake.usgs.gov/earthquakes/eventpage/ak022c3c7yeh","detail":"https://earthquake.usgs.gov/earthquakes/feed/v1.0/detail/ak022c3c7yeh.geojson","felt":null,"cdi":null,"mmi":null,"alert":null,"status":"automatic","tsunami":0,"sig":2,"net":"ak","code":"022c3c7yeh","ids":",ak022c3c7yeh,","sources":",ak,","types":",origin,phase-data,","nst":null,"dmin":null,"rms":0.17,"gap":null,"magType":"ml","type":"earthquake","title":"M 0.4 - Southern Alaska"},"geometry":{"type":"Point","coordinates":[-151.684,61.2888,71]},"id":"ak022c3c7yeh"},
{"type":"Feature","properties":{"mag":1.36,"place":"20km ESE of Anza, CA","time":1663716685670,"updated":1663717324870,"tz":null,"url":"https://earthquake.usgs.gov/earthquakes/eventpage/ci40103511","detail":"https://earthquake.usgs.gov/earthquakes/feed/v1.0/detail/ci40103511.geojson","felt":null,"cdi":null,"mmi":null,"alert":null,"status":"automatic","tsunami":0,"sig":28,"net":"ci","code":"40103511","ids":",ci40103511,","sources":",ci,","types":",focal-mechanism,nearby-cities,origin,phase-data,scitech-link,","nst":52,"dmin":0.03067,"rms":0.23,"gap":35,"magType":"ml","type":"earthquake","title":"M 1.4 - 20km ESE of Anza, CA"},"geometry":{"type":"Point","coordinates":[-116.4588333,33.5101667,5.24]},"id":"ci40103511"},
{"type":"Feature","properties":{"mag":1.7,"place":"Central Alaska","time":1663716207812,"updated":1663716870671,"tz":null,"url":"https://earthquake.usgs.gov/earthquakes/eventpage/ak022c3c68el","detail":"https://earthquake.usgs.gov/earthquakes/feed/v1.0/detail/ak022c3c68el.geojson","felt":null,"cdi":null,"mmi":null,"alert":null,"status":"automatic","tsunami":0,"sig":44,"net":"ak","code":"022c3c68el","ids":",ak022c3c68el,","sources":",ak,","types":",origin,phase-data,","nst":null,"dmin":null,"rms":0.98,"gap":null,"magType":"ml","type":"earthquake","title":"M 1.7 - Central Alaska"},"geometry":{"type":"Point","coordinates":[-149.3796,63.1907,79.5]},"id":"ak022c3c68el"},
{"type":"Feature","properties":{"mag":1.6,"place":"Alaska Peninsula","time":1663714766535,"updated":1663714852597,"tz":null,"url":"https://earthquake.usgs.gov/earthquakes/eventpage/ak022c3bsih9","detail":"https://earthquake.usgs.gov/earthquakes/feed/v1.0/detail/ak022c3bsih9.geojson","felt":null,"cdi":null,"mmi":null,"alert":null,"status":"automatic","tsunami":0,"sig":39,"net":"ak","code":"022c3bsih9","ids":",ak022c3bsih9,","sources":",ak,","types":",origin,phase-data,","nst":null,"dmin":null,"rms":0.76,"gap":null,"magType":"ml","type":"earthquake","title":"M 1.6 - Alaska Peninsula"},"geometry":{"type":"Point","coordinates":[-155.4719,58.2744,0]},"id":"ak022c3bsih9"}],"bbox":[-155.4719,33.5101667,0,-116.4588333,63.1907,79.5]}`,
	}
	// b, err := src.MarshalJSON()
	//err = json.Unmarshal(b, &f)
	// printStruct("", "", m)
	for i, v := range fake {
		testPrint(i, v, t)
	}
}

func testPrint(idx int, jsonString string, t *testing.T) {
	fmt.Printf("\nTest %d\n", idx)
	var f interface{}

	err := json.Unmarshal([]byte(jsonString), &f)
	if err != nil {
		t.Errorf("err:%v\n", err)
		return
	}
	if f == nil {
		t.Errorf("empty json\n")
		return
	}
	m := f.(map[string]interface{})

	jsonwalk.Walk(m, jsonwalk.Print())

	_, err = json.Marshal(m)
	if err != nil {
		t.Errorf("error marshalling json: %v", err)
	}
	// fmt.Printf("\nmarshalled=\n%v\n", string(mm))
}
