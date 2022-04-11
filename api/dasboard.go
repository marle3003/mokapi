package api

import (
	"mokapi/runtime/monitor"
	"net/http"
)

type dashboardInfo struct {
	*monitor.Monitor
	Http []*httpSummary
}

func (h *handler) getDashboard(w http.ResponseWriter, _ *http.Request) {
	dashboard := dashboardInfo{
		Monitor: h.app.Monitor,
		Http:    getHttpServices(h.app.Http, h.app.Monitor.Http),
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, dashboard)
}
