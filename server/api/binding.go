package api

import (
	"context"
	"encoding/json"
	"fmt"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"mokapi/models"
	"mokapi/server/api/asyncapi"
	"mokapi/server/api/openapi"
	"mokapi/version"
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
	path       string
}

type info struct {
	Version string `json:"version"`
}

// todo move to runtime module
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

func NewBinding(addr string, r *models.Runtime, path string) *Binding {
	b := &Binding{runtime: r, Addr: addr, path: path}
	b.server = &http.Server{Addr: addr, Handler: b}
	b.fileServer = http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir})
	return b
}

func (b *Binding) Start() {
	go func() {
		log.Infof("starting api on %v", b.Addr)
		err := b.server.ListenAndServe()
		if err != nil {
			log.Errorf("unable to start api on %v", b.Addr)
		}
	}()
}

func (b *Binding) Stop() {
	go func() {
		log.Infof("stopping api on %v", b.Addr)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		b.server.SetKeepAlivesEnabled(false)
		if err := b.server.Shutdown(ctx); err != nil {
			log.Errorf("could not gracefully shutdown server %v: %v", b.Addr, err.Error())
		}
	}()
}

func (b *Binding) Apply(_ interface{}) error {
	return nil
}

func (b *Binding) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, fmt.Sprintf("method %v is not allowed", r.Method), http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	switch p := r.URL.Path; {
	case len(b.path) > 0 && strings.HasPrefix(p, b.path):
		r.URL.Path = r.URL.Path[len(b.path):]
		b.ServeHTTP(w, r)
	case p == "/api/info":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info{Version: version.BuildVersion})
	case p == "/api/services":
		b.getServices(w, r)
	case strings.HasPrefix(p, "/api/services/openapi"):
		b.getService(w, r)
	case strings.HasPrefix(p, "/api/services/asyncapi"):
		b.getAsyncService(w, r)
	case strings.HasPrefix(p, "/api/services/smtp"):
		b.getSmtpService(w, r)
	case p == "/api/dashboard":
		b.getDashboard(w, r)
	case strings.HasPrefix(p, "/api/dashboard/http/requests/"):
		b.getHttpRequest(w, r)
	case strings.HasPrefix(p, "/api/dashboard/kafka"):
		b.handleKafka(w, r)
	case strings.HasPrefix(p, "/api/dashboard/smtp/mails/"):
		b.getSmtpMail(w, r)
	default:
		b.fileServer.ServeHTTP(w, r)
	}
}

func (b *Binding) getService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	if s, ok := b.runtime.OpenApi[name]; ok {
		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(openapi.NewService(s))
		if err != nil {
			log.Errorf("Error in writing service response: %v", err.Error())
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
		summary := serviceSummary{Name: s.Info.Name, Description: s.Info.Description, Type: "LDAP"}
		summary.BaseUrls = append(summary.BaseUrls, baseUrl{Url: s.Address})
		if len(summary.Name) == 0 {
			summary.Name = k
		}
		services = append(services, summary)
	}
	for k, s := range b.runtime.Smtp {
		summary := serviceSummary{Name: s.Name, Type: "SMTP"}
		summary.BaseUrls = append(summary.BaseUrls, baseUrl{Url: s.Address})
		if len(summary.Name) == 0 {
			summary.Name = k
		}
		services = append(services, summary)
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(services)
	if err != nil {
		log.Errorf("error in writing service response: %v", err.Error())
	}
}

func (b *Binding) getDashboard(w http.ResponseWriter, _ *http.Request) {
	dashboard := newDashboard(b.runtime)

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(dashboard)
	if err != nil {
		log.Errorf("Error in writing dashboard response: %v", err.Error())
	}
}

func (b *Binding) getAsyncService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	if c, ok := b.runtime.AsyncApi[name]; ok {
		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(asyncapi.NewService(c))
		if err != nil {
			log.Errorf("Error in writing service response: %v", err.Error())
		}
	} else {
		w.WriteHeader(404)
	}
}

func (b *Binding) getHttpRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	segments := strings.Split(r.URL.Path, "/")
	id := segments[len(segments)-1]
	for _, o := range b.runtime.Metrics.LastRequests {
		if o.Id == id {
			err := json.NewEncoder(w).Encode(newRequest(o))
			if err != nil {
				log.Errorf("Error in writing service response: %v", err.Error())
			}
			return
		}
	}

	w.WriteHeader(404)
}

func (b *Binding) getSmtpMail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	segments := strings.Split(r.URL.Path, "/")
	id := segments[len(segments)-1]
	for _, o := range b.runtime.Metrics.LastMails {
		if o.Mail.Id == id {
			err := json.NewEncoder(w).Encode(newMail(o))
			if err != nil {
				log.Errorf("Error in writing smtp mail response: %v", err.Error())
			}
			return
		}
	}

	w.WriteHeader(404)
}
