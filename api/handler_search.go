package api

import (
	"mokapi/runtime"
	"mokapi/runtime/search"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

const queryText = "q"
const searchLimit = "limit"
const searchIndex = "index"

var searchFacetExcludeQueryParams = map[string]bool{
	queryText:   true,
	searchLimit: true,
	searchIndex: true,
}

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

	sr.Facets = getFacets(r.URL.Query())

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

func getFacets(query url.Values) map[string]string {
	facets := make(map[string]string)
	for key, values := range query {
		if slices.Contains(runtime.SupportedFacets, key) {
			if !searchFacetExcludeQueryParams[key] {
				facets[key] = values[0]
			}
		}
	}
	return facets
}
