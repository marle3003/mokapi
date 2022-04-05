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
	dashboard       bool
	dashboardAssets *assetfs.AssetFS
	app             *runtime.App
}

type info struct {
	Version string `json:"version"`
}

type apiError struct {
	Message string `json:"message"`
}

func New(app *runtime.App, dashboard bool) http.Handler {
	return &handler{
		dashboard:       dashboard,
		dashboardAssets: nil,
		app:             app,
	}
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
	case p == "/api/info":
		h.getInfo(w, r)
	case strings.HasPrefix(p, "/api/services/http/"):
		h.getHttpService(w, r)
	case strings.HasPrefix(p, "/api/services/kafka/"):
		h.getKafkaService(w, r)
	}
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
