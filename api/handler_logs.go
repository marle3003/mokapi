package api

import "net/http"

func (h *handler) getLog(w http.ResponseWriter, _ *http.Request) {

	result := getKafkaServices(h.app.Kafka, h.app.Monitor)
	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}
