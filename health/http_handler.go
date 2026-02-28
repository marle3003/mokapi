package health

import (
	"fmt"
	"mokapi/config/static"
	"mokapi/lib"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type handler struct {
	path string
	log  bool
}

func New(cfg static.Health) http.Handler {
	return &handler{path: healthPath(cfg), log: cfg.Log}
}

func BuildUrl(cfg static.Health) (*url.URL, error) {
	s := fmt.Sprintf("http://:%v%v", cfg.Port, healthPath(cfg))
	return url.Parse(s)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, fmt.Sprintf("method %v is not allowed", r.Method), http.StatusMethodNotAllowed)
		if h.log {
			log.Warnf("healthcheck: method not allowed: %v %v", r.Method, lib.GetUrl(r))
		}
		return
	}

	switch r.URL.Path {
	case h.path:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
		if h.log {
			log.Infof("healthcheck: %v %v: healthy", r.Method, lib.GetUrl(r))
		}
	default:
		http.NotFound(w, r)
		if h.log {
			log.Debugf("healthcheck: not found: %v %v", r.Method, lib.GetUrl(r))
		}
	}
}

func healthPath(cfg static.Health) string {
	path := "/health"
	if cfg.Path != "" {
		path = cfg.Path
	}
	return path
}
