package main

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

/*
Example new as relic response
{
  "results": [
    {
      "sum": 1
    }
  ],
  "performanceStats": {
    "inspectedCount": 591101,
    "omittedCount": 0,
    "matchCount": 4156,
    "wallClockTime": 129
  },
  "metadata": {
    "eventTypes": [
      "Metric"
    ],
    "eventType": "Metric",
    "timeAggregations": [
      "raw metrics"
    ],
    "openEnded": true,
    "beginTime": "2021-03-10T17:45:21Z",
    "endTime": "2021-03-10T18:15:21Z",
    "beginTimeMillis": 1615398321253,
    "endTimeMillis": 1615400121253,
    "rawSince": "30 MINUTES AGO",
    "rawUntil": "NOW",
    "rawCompareWith": "",
    "guid": "1873f2d2-2ac1-8827-2e87-7487064fb369",
    "routerGuid": "0ab7d878-2aa1-fe72-644f-8443cddadb0d",
    "messages": [],
    "contents": [
      {
        "function": "sum",
        "attribute": "istio_requests_total",
        "simple": true
      }
    ]
  }
}
*/

// NewRelicResult is the result section of NewRelicResponse
type NewRelicResult struct {
	Sum float64 `json:"sum"`
}

// NewRelicResponse struct captures a response from new relic
type NewRelicResponse struct {
	Results []NewRelicResult `json:"results"`
}

// NewRelicHandlerFunc mimics responses for requests to NewRelic backend
func (newrelicHandler *RequestHandler) NewRelicHandlerFunc(w http.ResponseWriter, req *http.Request) {
	if !newrelicHandler.conf.MatchHeaders(req) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("headers are not matching"))
	} else {
		if version := newrelicHandler.conf.GetVersion(req); version != nil {
			b, _ := json.Marshal(NewRelicResponse{
				Results: []NewRelicResult{
					{
						Sum: getValue(version),
					},
				},
			})
			w.WriteHeader(http.StatusOK)
			w.Write(b)
			log.Info(version)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - cannot find any matching version in request!"))
		}
	}
}
