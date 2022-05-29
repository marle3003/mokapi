package api

import (
	"mokapi/runtime/events"
	"net/http"
)

func (h *handler) getEvents(w http.ResponseWriter, r *http.Request) {
	traits := events.NewTraits()
	for k := range r.URL.Query() {
		traits.With(k, r.URL.Query().Get(k))
	}

	list := events.Events(traits)

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, list)
}
