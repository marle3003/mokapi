package api

import (
	"mokapi/runtime/metrics"
	"net/http"
)

type dashboardInfo struct {
	Metrics []metrics.Metric `json:"metrics"`
}

func (h *handler) getDashboard(w http.ResponseWriter, _ *http.Request) {
	dashboard := dashboardInfo{
		Metrics: []metrics.Metric{
			h.app.Monitor.StartTime,
			h.app.Monitor.MemoryUsage,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, dashboard)
}
