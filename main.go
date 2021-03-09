// Package main serves up synthetic metrics.
// This is intended for integration tests of iter8-analytics service
// And for creating the code samples in Iter8 documentation at https://iter8.tools
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func hello(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "hello\n")
}

func headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

// HandlerFunc type is the type of function used as http request handler
type HandlerFunc func(w http.ResponseWriter, req *http.Request)

// PrometheusResponse struct captures a response from prometheus
/*
{
    "status": "success",
    "data": {
      "resultType": "vector",
      "result": [
        {
          "value": [1556823494.744, "21.7639"]
        }
      ]
    }
}
*/
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// func getHandlerFunc(m Matcher, provider string) HandlerFunc {
// 	switch provider {
// 	case "Prometheus":
// 		var f HandlerFunc = func(w http.ResponseWriter, req *http.Request) {
// 			if m.Match(req) {
// 				b, _ := json.Marshal(PrometheusResponse{
// 					Status: "success",
// 				})
// 				w.WriteHeader(http.StatusOK)
// 				w.Write(b)
// 			} else {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				w.Write([]byte("500 - non-matching request!"))
// 			}
// 		}
// 		return f
// 	default:
// 		panic("unknown provider: " + provider)
// 	}
// }

// Param is simply a name-value pair representing name and value of HTTP query param
type Param struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// MetricInfo provides information about the metric to be generated
type MetricInfo struct {
	Type       string   `json:"type"`
	Rate       *float64 `json:"rate"`
	Shift      *float64 `json:"shift"`
	Multiplier *float64 `json:"multiplier"`
	Alpha      *float64 `json:"alpha"`
	Beta       *float64 `json:"beta"`
}

// VersionInfo struct provides the param and metric information for a version
type VersionInfo struct {
	Params []Param    `json:"params"`
	Metric MetricInfo `json:"metric"`
}

// URIConf is the metrics gen configuration for a URI
type URIConf struct {
	Versions []VersionInfo     `json:"versions"`
	Headers  map[string]string `json:"headers"`
	URI      string            `json:"uri"`
	Provider string            `json:"provider"`
}

func main() {

	http.HandleFunc("/hello", hello)
	http.HandleFunc("/headers", headers)

	// find config url from env
	configURL := os.Getenv("CONFIG_URL")
	if len(configURL) == 0 {
		panic("No config URL supplied")
	}

	// read in config from url into config struct
	resp, err := http.Get(configURL)
	if err != nil {
		panic("HTTP GET with configured url did not succeed: " + configURL)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	var uriConfs []URIConf
	err = json.Unmarshal(body, &uriConfs)
	if err != nil {
		panic(err)
	}

	// for _, conf := range uriConfs {
	// 	http.HandleFunc(conf.URI, getHandlerFunc(conf.Matcher, conf.Provider))
	// }

	// http.ListenAndServe(":8090", nil)
}
