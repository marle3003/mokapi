package api

import (
	"mokapi/runtime/events"
	"net/http"
	"strings"
)

func (h *handler) getEvents(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	traits := events.NewTraits()

	if len(segments) > 3 {
		traits.WithNamespace(segments[3])
	}

	list := events.Events(traits)

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, list)
}
