package api

import (
	"net/http"
	"net/url"
	"strings"
)

func (h *handler) getSearchResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	queryText := getQueryParamInsensitive(r.URL.Query(), "querytext")
	results, err := h.app.Search(queryText)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	} else {
		writeJsonBody(w, results)
	}
}

func getQueryParamInsensitive(values url.Values, key string) string {
	key = strings.ToLower(key)
	for k, v := range values {
		if strings.ToLower(k) == key && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}
