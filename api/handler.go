package api

import (
	"encoding/json"
	"fmt"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	log "github.com/sirupsen/logrus"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/version"
	"net/http"
	"net/url"
	"strings"
)

type handler struct {
	path       string
	app        *runtime.App
	fileServer http.Handler
}

type info struct {
	Version string `json:"version"`
}

type apiError struct {
	Message string `json:"message"`
}

func New(app *runtime.App, config static.Api) http.Handler {
	h := &handler{
		path: config.Path,
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
	if r.Method != "GET" {
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
		h.getHttpService(w, r)
	case strings.HasPrefix(p, "/api/services/kafka/"):
		h.getKafkaService(w, r)
	case strings.HasPrefix(p, "/api/services/smtp/"):
		h.getSmtpService(w, r)
	case strings.HasPrefix(p, "/api/http/requests"):
		h.getHttpRequests(w, r)
	case p == "/api/dashboard":
		h.getDashboard(w, r)
	case strings.HasPrefix(p, "/api/metrics"):
		h.getMetrics(w, r)
	case strings.HasPrefix(p, "/api/events"):
		h.getEvents(w, r)
	case h.fileServer != nil:
		h.fileServer.ServeHTTP(w, r)
	default:
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
	writeJsonBody(w, info{Version: version.BuildVersion})
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
