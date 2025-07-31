package api

import (
	"mokapi/runtime/search"
	"net/http"
	"net/url"
	"strings"
)

const queryText = "q"
const searchLimit = "limit"
const searchIndex = "index"

func (h *handler) getSearchResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sr := search.Request{
		QueryText: getQueryParamInsensitive(r.URL.Query(), queryText),
		Limit:     10,
	}
	var err error
	sr.Index, sr.Limit, err = getPageInfo(r)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	results, err := h.app.Search(sr)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
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
