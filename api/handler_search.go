package api

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/runtime/search"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

const queryText = "querytext"
const searchLimit = "limit"
const searchIndex = "index"

var skipTerms = []string{queryText, searchLimit, searchIndex}

func (h *handler) getSearchResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	values := r.URL.Query()
	terms := make(map[string]string)
	for key, vals := range values {
		if len(vals) > 0 && !slices.ContainsFunc(skipTerms, func(s string) bool {
			return strings.EqualFold(key, s)
		}) {
			terms[key] = vals[0]
		}
	}

	sr := search.Request{
		Query: getQueryParamInsensitive(r.URL.Query(), queryText),
		Terms: terms,
		Limit: 10,
	}

	sIndex := getQueryParamInsensitive(r.URL.Query(), "index")
	if sIndex != "" {
		var err error
		sr.Index, err = strconv.Atoi(sIndex)
		if err != nil {
			writeError(w, fmt.Errorf("invalid index value: %s", err), http.StatusBadRequest)
		}
	}
	sLimit := getQueryParamInsensitive(r.URL.Query(), "limit")
	if sLimit != "" {
		var err error
		sr.Limit, err = strconv.Atoi(sLimit)
		if err != nil {
			writeError(w, fmt.Errorf("invalid limit value: %s", err), http.StatusBadRequest)
		}
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
