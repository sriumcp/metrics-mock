// Package main serves up synthetic metrics.
// This is intended for integration tests of iter8-analytics service
// And for creating the code samples in Iter8 documentation at https://iter8.tools
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
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

func getHandlerFunc(conf URIConf) HandlerFunc {
	switch conf.Provider {
	case "Prometheus":
		// var f HandlerFunc = func(w http.ResponseWriter, req *http.Request) {
		// 	if m.Match(req) {
		// 		b, _ := json.Marshal(PrometheusResponse{
		// 			Status: "success",
		// 		})
		// 		w.WriteHeader(http.StatusOK)
		// 		w.Write(b)
		// 	} else {
		// 		w.WriteHeader(http.StatusInternalServerError)
		// 		w.Write([]byte("500 - non-matching request!"))
		// 	}
		// }
		// return f
		return hello
	default:
		panic("unknown provider: " + conf.Provider)
	}
}

// Param is simply a name-value pair representing name and value of HTTP query param
type Param struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// MetricInfo provides information about the metric to be generated
type MetricInfo struct {
	Type       string   `yaml:"type"`
	Rate       *float64 `yaml:"rate"`
	Shift      *float64 `yaml:"shift"`
	Multiplier *float64 `yaml:"multiplier"`
	Alpha      *float64 `yaml:"alpha"`
	Beta       *float64 `yaml:"beta"`
}

// VersionInfo struct provides the param and metric information for a version
type VersionInfo struct {
	Params []Param    `yaml:"params"`
	Metric MetricInfo `yaml:"metric"`
}

// URIConf is the metrics gen configuration for a URI
type URIConf struct {
	Versions []VersionInfo     `yaml:"versions"`
	Headers  map[string]string `yaml:"headers"`
	URI      string            `yaml:"uri"`
	Provider string            `yaml:"provider"`
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
	err = yaml.Unmarshal(body, &uriConfs)
	if err != nil {
		panic(err)
	}

	// check of URIs are unique
	uriset := make(map[string]struct{})
	for _, conf := range uriConfs {
		if _, ok := uriset[conf.URI]; ok {
			log.Error(uriset)
			log.Error(conf.URI)
			panic("URIs are not unique")
		}
		uriset[conf.URI] = struct{}{}
	}

	for _, conf := range uriConfs {
		http.HandleFunc(conf.URI, getHandlerFunc(conf))
	}

	http.ListenAndServe(":8080", nil)
}
