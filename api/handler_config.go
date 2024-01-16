package api

import (
	log "github.com/sirupsen/logrus"
	"mime"
	"mokapi/config/dynamic"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type config struct {
	Id       string    `json:"id"`
	Url      string    `json:"url"`
	Provider string    `json:"provider"`
	Time     time.Time `json:"time"`
}

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

func getConfigs(src []*dynamic.Config) (dst []config) {
	for _, cfg := range src {
		path := cfg.Info.Path()
		path = filepath.ToSlash(path)
		dst = append(dst, config{
			Id:       cfg.Info.Key(),
			Url:      path,
			Time:     cfg.Info.Time,
			Provider: cfg.Info.Provider,
		})
	}
	return
}
