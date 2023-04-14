package api

import (
	"encoding/json"
	"fmt"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	log "github.com/sirupsen/logrus"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

type handler struct {
	path       string
	base       string
	app        *runtime.App
	fileServer http.Handler
}

type info struct {
	Version        string   `json:"version"`
	ActiveServices []string `json:"activeServices,omitempty"`
}

type serviceType string

var (
	ServiceHttp  serviceType = "http"
	ServiceKafka serviceType = "kafka"
)

type service struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Contact     *contact         `json:"contact,omitempty"`
	Version     string           `json:"version,omitempty"`
	Type        serviceType      `json:"type"`
	Metrics     []metrics.Metric `json:"metrics,omitempty"`
}

type contact struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Email string `json:"email"`
}

type apiError struct {
	Message string `json:"message"`
}

func New(app *runtime.App, config static.Api) http.Handler {
	h := &handler{
		path: config.Path,
		base: config.Base,
		app:  app,
	}

	if config.Dashboard {
		h.fileServer = http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir})
	}

	return h
}

func BuildUrl(cfg static.Api) (*url.URL, error) {
	s := fmt.Sprintf("http://:%v%v", cfg.Port, cfg.Path)
	return url.Parse(s)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, fmt.Sprintf("method %v is not allowed", r.Method), http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch p := r.URL.Path; {
	case len(h.path) > 0 && strings.HasPrefix(p, h.path):
		r.URL.Path = r.URL.Path[len(h.path):]
		h.ServeHTTP(w, r)
	case p == "/api/info":
		h.getInfo(w, r)
	case p == "/api/services":
		h.getServices(w, r)
	case strings.HasPrefix(p, "/api/services/http/"):
		h.getHttpService(w, r, h.app.Monitor)
	case strings.HasPrefix(p, "/api/services/kafka/"):
		h.getKafkaService(w, r)
	case strings.HasPrefix(p, "/api/services/smtp/"):
		h.getSmtpService(w, r)
	case p == "/api/dashboard":
		h.getDashboard(w, r)
	case strings.HasPrefix(p, "/api/metrics"):
		h.getMetrics(w, r)
	case strings.HasPrefix(p, "/api/events"):
		h.getEvents(w, r)
	case p == "/api/schema/example":
		h.getExampleData(w, r)
	case h.fileServer != nil:
		if isAsset(r.URL.Path) {
			r.URL.Path = "/assets/" + filepath.Base(r.URL.Path)
		} else if filepath.Ext(r.URL.Path) == ".svg" {
			r.URL.Path = "/" + filepath.Base(r.URL.Path)
		} else {
			if len(h.path) > 0 || len(h.base) > 0 {
				base := h.path
				if len(h.base) > 0 {
					base = h.base
				}
				data, err := Asset("index.html")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				html := strings.Replace(string(data), "<base href=\"/\" />", fmt.Sprintf("<base href=\"%v/\" />", base), 1)
				_, err = w.Write([]byte(html))
				if err != nil {
					log.Errorf("unable to write index.html: %v", err)
				}
				return
			} else {
				r.URL.Path = "/"
			}
		}
		h.fileServer.ServeHTTP(w, r)
	default:
		log.Errorf("dashboard file not found: %v", r.URL)
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (h *handler) getServices(w http.ResponseWriter, _ *http.Request) {
	services := make([]interface{}, 0)
	services = append(services, getHttpServices(h.app.Http, h.app.Monitor)...)
	services = append(services, getKafkaServices(h.app.Kafka, h.app.Monitor)...)
	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, services)
}

func writeError(w http.ResponseWriter, err error, status int) {
	log.Error(err)
	data, err := json.Marshal(apiError{Message: err.Error()})
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	http.Error(w, string(data), status)
}

func (h *handler) getInfo(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	i := info{Version: h.app.Version}
	if len(h.app.Http) > 0 {
		i.ActiveServices = append(i.ActiveServices, "http")
	}
	if len(h.app.Kafka) > 0 {
		i.ActiveServices = append(i.ActiveServices, "kafka")
	}
	if len(h.app.Smtp) > 0 {
		i.ActiveServices = append(i.ActiveServices, "smtp")
	}
	if len(h.app.Ldap) > 0 {
		i.ActiveServices = append(i.ActiveServices, "ldap")
	}

	writeJsonBody(w, i)
}

func writeJsonBody(w http.ResponseWriter, i interface{}) {
	b, err := json.Marshal(i)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	}
	_, err = w.Write(b)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	}
}

func isAsset(path string) bool {
	return strings.Contains(path, "/assets/")
}
