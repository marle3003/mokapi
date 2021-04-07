package api

import (
	"context"
	"encoding/json"
	"fmt"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"mokapi/models"
	"mokapi/server/api/asyncapi"
	"mokapi/server/api/openapi"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Binding struct {
	runtime    *models.Runtime
	server     *http.Server
	Addr       string
	fileServer http.Handler
}

type serviceSummary struct {
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Version     string    `json:"version,omitempty"`
	Type        string    `json:"type,omitempty"`
	BaseUrls    []baseUrl `json:"baseUrls"`
}

type baseUrl struct {
	Url         string `json:"url"`
	Description string `json:"description"`
}

func NewBinding(addr string, r *models.Runtime) *Binding {
	b := &Binding{runtime: r, Addr: addr}
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
	case strings.HasPrefix(p, "/api/services/openapi"):
		b.getService(w, r)
	case strings.HasPrefix(p, "/api/services/asyncapi"):
		b.getAsyncService(w, r)
	case p == "/api/dashboard":
		b.getDashboard(w, r)
	default:
		b.fileServer.ServeHTTP(w, r)
	}
}

func (b *Binding) getService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	if s, ok := b.runtime.OpenApi[name]; ok {
		w.Header().Set("Content-Type", "application/json")

		error := json.NewEncoder(w).Encode(openapi.NewService(s))
		if error != nil {
			log.Errorf("Error in writing service response: %v", error.Error())
		}
	} else {
		w.WriteHeader(404)
	}
}

func (b *Binding) getServices(w http.ResponseWriter, _ *http.Request) {
	services := make([]interface{}, 0)

	for k, c := range b.runtime.OpenApi {
		s := openapi.NewService(c)
		if len(s.Name) == 0 {
			s.Name = k
		}
		services = append(services, s)
	}
	for k, c := range b.runtime.AsyncApi {
		s := asyncapi.NewService(c)
		if len(s.Name) == 0 {
			s.Name = k
		}
		services = append(services, s)
	}
	for k, s := range b.runtime.Ldap {
		summary := serviceSummary{Name: s.Info.Name, Type: "LDAP"}
		summary.BaseUrls = append(summary.BaseUrls, baseUrl{Url: s.Address})
		if len(summary.Name) == 0 {
			summary.Name = k
		}
		services = append(services, summary)
	}

	w.Header().Set("Content-Type", "application/json")

	error := json.NewEncoder(w).Encode(services)
	if error != nil {
		log.Errorf("Error in writing service response: %v", error.Error())
	}
}

func (b *Binding) getDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard := newDashboard(b.runtime)

	w.Header().Set("Content-Type", "application/json")

	error := json.NewEncoder(w).Encode(dashboard)
	if error != nil {
		log.Errorf("Error in writing dashboard response: %v", error.Error())
	}
}

func (b *Binding) getAsyncService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	if c, ok := b.runtime.AsyncApi[name]; ok {
		w.Header().Set("Content-Type", "application/json")

		error := json.NewEncoder(w).Encode(asyncapi.NewService(c))
		if error != nil {
			log.Errorf("Error in writing service response: %v", error.Error())
		}
	} else {
		w.WriteHeader(404)
	}
}
