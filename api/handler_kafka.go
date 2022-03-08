package api

import (
	"net/http"
	"strings"
)

func (h *handler) getKafkaService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	if s, ok := h.app.Http[name]; ok {
		w.Header().Set("Content-Type", "application/json")
		writeJsonBody(w, s)
	} else {
		w.WriteHeader(404)
	}
}
