package api

import (
	"mokapi/runtime/metrics"
	"net/http"
	"strings"
)

func (h *handler) getMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := make([]metrics.QueryOptions, 0)
	result := make([]metrics.Metric, 0)

	segments := strings.Split(r.URL.Path, "/")
	if len(segments) > 3 {
		query = append(query, metrics.ByNamespace(segments[3]))
	}

	if !r.URL.Query().Has("names") {
		result = h.app.Monitor.FindAll(query...)
	} else {
		names := strings.Split(r.URL.Query().Get("names"), ",")
		for _, name := range names {
			result = append(
				result,
				h.app.Monitor.FindAll(append(query, metrics.ByFQName(name))...)...,
			)
		}
	}

	writeJsonBody(w, result)
}
