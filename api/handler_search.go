package api

import (
	"github.com/pkg/errors"
	"mokapi/runtime/search"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const queryText = "querytext"
const searchLimit = "limit"
const searchIndex = "index"

func (h *handler) getSearchResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sr := search.Request{Limit: 10}

	q := getQueryParamInsensitive(r.URL.Query(), queryText)
	sr.Query, sr.Params = parseQuery(q)

	var err error
	sr.Index, sr.Limit, err = getPageInfo(r)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	results, err := h.app.Search(sr)
	if err != nil {
		if errors.Is(err, &search.ErrNotEnabled{}) {
			writeError(w, err, http.StatusBadRequest)
		} else {
			writeError(w, err, http.StatusInternalServerError)
		}
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

func parseQuery(query string) (string, map[string]string) {
	re := regexp.MustCompile(`([\w.]+)=("[^"]+"|\S+)`)

	params := make(map[string]string)
	matches := re.FindAllStringSubmatch(query, -1)

	s := query
	for _, m := range matches {
		key := m[1]
		value := strings.Trim(m[2], `"`)
		params[key] = value
		s = strings.Replace(s, m[0], "", 1)
	}

	s = strings.TrimSpace(s)
	return s, params
}
