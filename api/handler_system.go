package api

import (
	"mokapi/runtime/events"
	"net/http"
	"strings"
)

func (h *handler) serveSystem(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	if len(segments) == 4 && segments[3] == "events" {
		traits := events.NewTraits()
		for k := range r.URL.Query() {
			traits.With(k, r.URL.Query().Get(k))
		}

		data := events.GetStores(traits)
		if data == nil {
			w.WriteHeader(404)
		} else {
			w.Header().Set("Content-Type", "application/json")
			writeJsonBody(w, data)
		}
	} else {
		w.WriteHeader(404)
		return
	}
}
