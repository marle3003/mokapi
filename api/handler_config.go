package api

import (
	"fmt"
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
	if len(segments) == 3 {
		h.getConfigs(w)
	} else if len(segments) == 4 {
		h.getConfigMetaData(w, segments[3])
	} else if len(segments) == 5 {
		h.getConfigData(w, r, segments[3])
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (h *handler) getConfigs(w http.ResponseWriter) {
	var configs []config
	for _, c := range h.app.Configs {
		configs = append(configs, toConfig(c))
	}
	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, configs)
}

func (h *handler) getConfigMetaData(w http.ResponseWriter, key string) {
	c := h.app.FindConfig(key)
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, toConfig(c))
}

func (h *handler) getConfigData(w http.ResponseWriter, r *http.Request, key string) {
	c := h.app.FindConfig(key)
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	token := r.Header.Get("If-None-Match")
	checksum := fmt.Sprintf("%x", c.Info.Checksum)
	if token != "" && token == checksum {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	path := c.Info.Kernel().Path()
	ext := filepath.Ext(path)
	mt := mime.TypeByExtension(filepath.Ext(ext))
	if mt == "" {
		mt = "text/plain"
	}
	w.Header().Set("Last-Modified", c.Info.Time.UTC().Format(http.TimeFormat))
	w.Header().Set("Content-Type", mt)
	w.Header().Set("Content-Disposition", "inline; filename=\""+filepath.Base(path)+"\"")
	w.Header().Set("ETag", checksum)
	w.Header().Set("Cache-Control", "no-cache")
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
	for _, ref := range cfg.Refs.List(false) {
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
