package api

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func (h *handler) getConfig(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	key := segments[3]

	c, ok := h.app.Configs[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Last-Modified", c.Info.Time.UTC().Format(http.TimeFormat))
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write(c.Raw)
	if err != nil {
		log.Errorf("http write file %v failed: %v", c.Info.Url.String(), err)
	}
}
