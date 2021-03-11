// Package main serves up synthetic metrics.
// This is intended for integration tests of iter8-analytics service
// And for creating the code samples in Iter8 documentation at https://iter8.tools
package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
	distuv "gonum.org/v1/gonum/stat/distuv"
	yaml "gopkg.in/yaml.v2"
)

var start time.Time

func init() {
	start = time.Now()
}

// RequestHandler handles requests based on confs
type RequestHandler struct {
	conf URIConf
}

// HandlerFunc type is the type of function used as http request handler
type HandlerFunc func(w http.ResponseWriter, req *http.Request)

func getHandlerFunc(conf URIConf) HandlerFunc {
	switch conf.Provider {
	case "prometheus":
		return (&RequestHandler{
			conf: conf,
		}).PrometheusHandlerFunc
	case "newrelic":
		return (&RequestHandler{
			conf: conf,
		}).NewRelicHandlerFunc
	default:
		panic("unknown provider: " + conf.Provider)
	}
}

func getValue(version *VersionInfo) float64 {
	elapsed := time.Now().Sub(start)
	if version.Metric.Type == "counter" {
		return elapsed.Seconds() * version.Metric.Rate
	}
	if version.Metric.Type == "gauge" {
		log.Info("metricinfo...", version.Metric)
		a, b := (elapsed.Seconds()+1.0)*version.Metric.Alpha, (elapsed.Seconds()+1.0)*version.Metric.Beta
		beta := distuv.Beta{
			Alpha: a,
			Beta:  b,
		}.Rand()
		log.Info("elapsed seconds, a, b, beta ...")
		log.Info(elapsed.Seconds(), a, b, beta)
		return version.Metric.Shift + beta*version.Metric.Multiplier
	}
	return 21.7639
}

// Param is simply a name-value pair representing name and value of HTTP query param
type Param struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// MetricInfo provides information about the metric to be generated
type MetricInfo struct {
	Type       string  `yaml:"type"`
	Rate       float64 `yaml:"rate"`
	Shift      float64 `yaml:"shift"`
	Multiplier float64 `yaml:"multiplier"`
	Alpha      float64 `yaml:"alpha"`
	Beta       float64 `yaml:"beta"`
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

// MatchHeaders ensures that the headers in URIConf match the headers in the request
func (u *URIConf) MatchHeaders(req *http.Request) bool {
	for key, val := range u.Headers {
		if req.Header.Get(key) != val {
			return false
		}
	}
	return true
}

// GetVersion finds the correct version in URIConf based on params in the request or returns nil if no matching version is found
func (u *URIConf) GetVersion(req *http.Request) *VersionInfo {
	for _, version := range u.Versions {
		found := true
		for _, param := range version.Params {
			val, ok := req.URL.Query()[param.Name]
			// query has this parameter
			if ok && len(val[0]) > 0 {
				matched, err := regexp.Match(param.Value, []byte(val[0]))
				if err != nil || !matched { // query parameter does not match value
					log.Warn("found no match for ... " + param.Name)
					log.Warn(param.Value)
					log.Warn(val[0])
					found = false
					break
				} else { // query parameter matches value
					log.Info("found match for ... " + param.Name)
					log.Info(param.Value)
					log.Info(val[0])
				}
			} else { // query doesn't have this parameter
				found = false
			}
		}
		if found { // return the first version found
			return &version
		}
	}
	return nil
}

func main() {
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

	// check if URIs are unique
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
