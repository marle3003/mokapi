package api

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func (h *handler) getHttpService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	if s, ok := h.app.Http[name]; ok {
		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(s)
		if err != nil {
			log.Errorf("Error in writing service response: %v", err.Error())
		}
	} else {
		w.WriteHeader(404)
	}
}
