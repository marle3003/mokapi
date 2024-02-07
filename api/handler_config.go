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
	Id       string      `json:"id"`
	Url      string      `json:"url"`
	Provider string      `json:"provider"`
	Time     time.Time   `json:"time"`
	Refs     []configRef `json:"refs,omitempty"`
}

type configRef struct {
	Id       string    `json:"id"`
	Url      string    `json:"url"`
	Provider string    `json:"provider"`
	Time     time.Time `json:"time"`
}

func (h *handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	if len(segments) == 5 {
		h.getConfigData(w, segments[3])
	} else if len(segments) == 4 {
		h.getConfigMetaData(w, segments[3])
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (h *handler) getConfigMetaData(w http.ResponseWriter, key string) {
	c, ok := h.app.Configs[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, toConfig(c))
}

func (h *handler) getConfigData(w http.ResponseWriter, key string) {
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
		dst = append(dst, toConfig(cfg))
	}
	return
}

func toConfig(cfg *dynamic.Config) config {
	var refs []configRef
	for _, ref := range cfg.Refs.List() {
		refs = append(refs, configRef{
			Id:       ref.Info.Key(),
			Url:      filepath.ToSlash(ref.Info.Path()),
			Provider: ref.Info.Provider,
			Time:     ref.Info.Time,
		})
	}

	return config{
		Id:       cfg.Info.Key(),
		Url:      filepath.ToSlash(cfg.Info.Path()),
		Time:     cfg.Info.Time,
		Provider: cfg.Info.Provider,
		Refs:     refs,
	}
}
