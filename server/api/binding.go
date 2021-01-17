package api

import (
	"context"
	"encoding/json"
	"fmt"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"mokapi/models"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Binding struct {
	appl       *models.Application
	server     *http.Server
	Addr       string
	fileServer http.Handler
}

type serviceSummary struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
}

func NewBinding(addr string, a *models.Application) *Binding {
	b := &Binding{appl: a, Addr: addr}
	b.server = &http.Server{Addr: addr, Handler: b}
	b.fileServer = http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir})
	return b
}

func (b *Binding) Start() {
	go func() {
		log.Infof("Starting api on %v", b.Addr)
		b.server.ListenAndServe()
	}()
}

func (b *Binding) Stop() {
	go func() {
		log.Infof("Stopping api on %v", b.Addr)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		b.server.SetKeepAlivesEnabled(false)
		if error := b.server.Shutdown(ctx); error != nil {
			log.Errorf("Could not gracefully shutdown server %v", b.Addr)
		}
	}()
}

func (b *Binding) Apply(data interface{}) error {
	return nil
}

func (b *Binding) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, fmt.Sprintf("method %v is not allowed", r.Method), http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	switch p := r.URL.Path; {
	case p == "/api/services":
		b.getServices(w, r)
	case strings.HasPrefix(p, "/api/services/"):
		b.getService(w, r)
	case p == "/api/dashboard":
		b.getDashboard(w, r)
	default:
		b.fileServer.ServeHTTP(w, r)
	}
}

func (b *Binding) getService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[3]

	if s, ok := b.appl.WebServices[name]; ok {
		service := newService(s)

		w.Header().Set("Content-Type", "application/json")

		error := json.NewEncoder(w).Encode(service)
		if error != nil {
			log.Errorf("Error in writing service response: %v", error.Error())
		}
	} else {
		w.WriteHeader(404)
	}
}

func (b *Binding) getServices(w http.ResponseWriter, r *http.Request) {
	services := make([]serviceSummary, 0)

	for _, s := range b.appl.WebServices {
		services = append(services, newServiceSummary(s))
	}

	w.Header().Set("Content-Type", "application/json")

	error := json.NewEncoder(w).Encode(services)
	if error != nil {
		log.Errorf("Error in writing service response: %v", error.Error())
	}
}

func (b *Binding) getDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard := newDashboard(b.appl)

	w.Header().Set("Content-Type", "application/json")

	error := json.NewEncoder(w).Encode(dashboard)
	if error != nil {
		log.Errorf("Error in writing dashboard response: %v", error.Error())
	}
}
