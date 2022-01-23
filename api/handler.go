package api

import (
	"fmt"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"mokapi/config/static"
	"mokapi/runtime"
	"net/http"
	"net/url"
)

type handler struct {
	dashboard       bool
	dashboardAssets *assetfs.AssetFS
	app             *runtime.App
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
}
