package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

/*
Example prometheus response
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

// PrometheusResult is the result section of PrometheusResponseData
type PrometheusResult []struct {
	Value []interface{} `json:"value"`
}

// PrometheusResponseData is the data section of Prometheus response
type PrometheusResponseData struct {
	ResultType string           `json:"resultType"`
	Result     PrometheusResult `json:"result"`
}

// PrometheusResponse struct captures a response from prometheus
type PrometheusResponse struct {
	Status string                 `json:"status"`
	Data   PrometheusResponseData `json:"data"`
}

// PrometheusHandlerFunc mimics responses for requests to Prometheus backend
func (promHandler *RequestHandler) PrometheusHandlerFunc(w http.ResponseWriter, req *http.Request) {
	if !promHandler.conf.MatchHeaders(req) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("headers are not matching"))
	} else {
		if version := promHandler.conf.GetVersion(req); version != nil {
			b, _ := json.Marshal(PrometheusResponse{
				Status: "success",
				Data: PrometheusResponseData{
					ResultType: "vector",
					Result: PrometheusResult{
						{
							Value: []interface{}{1556823494.744, fmt.Sprint(getValue(version))},
						},
					},
				},
			})
			w.WriteHeader(http.StatusOK)
			w.Write(b)
			log.Info(version)
		} else {
			log.Error("cannot find any matching version in request")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - cannot find any matching version in request!"))
		}
	}
}
