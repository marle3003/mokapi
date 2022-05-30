package api

import (
	"mokapi/runtime/events"
	"net/http"
	"strings"
)

func (h *handler) getEvents(w http.ResponseWriter, r *http.Request) {
	var data interface{}

	segments := strings.Split(r.URL.Path, "/")
	if len(segments) == 3 {
		traits := events.NewTraits()
		for k := range r.URL.Query() {
			traits.With(k, r.URL.Query().Get(k))
		}

		data = events.GetEvents(traits)
	} else {
		e := events.GetEvent(segments[3])
		if e.IsValid() {
			data = e
		} else {
			w.WriteHeader(404)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, data)
}
