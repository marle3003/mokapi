package api

import (
	log "github.com/sirupsen/logrus"
	"mime"
	"net/http"
	"path/filepath"
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

	path := c.Info.Path()
	ext := filepath.Ext(path)
	mt := mime.TypeByExtension(filepath.Ext(ext))
	if mt == "" {
		mt = "text/plain"
	}
	w.Header().Set("Last-Modified", c.Info.Time.UTC().Format(http.TimeFormat))
	w.Header().Set("Content-Type", mt)
	w.Header().Set("Content-Disposition", "inline; filename=\""+filepath.Base(path)+"\"")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write(c.Raw)
	if err != nil {
		log.Errorf("http write file %v failed: %v", c.Info.Url.String(), err)
	}
}
