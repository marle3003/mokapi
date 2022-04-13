package api

import (
	"net/http"
)

type dashboardInfo struct {
	StartTime         int64          `json:"startTime"`
	MemoryUsage       int64          `json:"memoryUsage"`
	HttpRequests      int64          `json:"httpRequests"`
	HttpErrorRequests int64          `json:"httpErrorRequests"`
	KafkaMessages     int64          `json:"kafkaMessages"`
	Http              []*httpSummary `json:"httpServices"`
}

func (h *handler) getDashboard(w http.ResponseWriter, _ *http.Request) {
	dashboard := dashboardInfo{
		StartTime:         int64(h.app.Monitor.StartTime.Value()),
		MemoryUsage:       int64(h.app.Monitor.MemoryUsage.Value()),
		HttpRequests:      int64(h.app.Monitor.Http.RequestCounter.Value()),
		HttpErrorRequests: int64(h.app.Monitor.Http.RequestErrorCounter.Value()),
		KafkaMessages:     int64(h.app.Monitor.Kafka.Messages.Value()),
		Http:              getHttpServices(h.app.Http, h.app.Monitor.Http),
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, dashboard)
}
