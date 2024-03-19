package api

import (
	"mokapi/json/generator"
	"net/http"
)

func (h *handler) handleFakerTree(w http.ResponseWriter, r *http.Request) {
	node := generator.FindByName("")

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, node)
}
