package api

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/webui"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type handler struct {
	config     static.Api
	path       string
	base       string
	app        *runtime.App
	fileServer http.Handler
	index      string
}

type info struct {
	Version        string     `json:"version"`
	BuildTime      string     `json:"buildTime"`
	ActiveServices []string   `json:"activeServices,omitempty"`
	Search         searchInfo `json:"search"`
}

type searchInfo struct {
	Enabled bool `json:"enabled"`
}

type serviceType string

var (
	ServiceHttp  serviceType = "http"
	ServiceKafka serviceType = "kafka"
	ServiceMail  serviceType = "mail"
	ServiceLdap  serviceType = "ldap"
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
		config: config,
		path:   config.Path,
		base:   config.Base,
		app:    app,
	}

	if config.Dashboard {
		webapp := webui.App
		b, err := webapp.ReadFile("dist/index.html")
		if err != nil {
			panic(err)
		}
		h.index = string(b)

		dist, err := fs.Sub(webapp, "dist")
		if err != nil {
			panic(err)
		}

		h.fileServer = http.FileServer(http.FS(dist))
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
	case strings.HasPrefix(p, "/api/services/mail/"):
		h.handleMailService(w, r)
	case strings.HasPrefix(p, "/api/services/ldap/"):
		h.handleLdapService(w, r)
	case p == "/api/dashboard":
		h.getDashboard(w, r)
	case strings.HasPrefix(p, "/api/metrics"):
		h.getMetrics(w, r)
	case strings.HasPrefix(p, "/api/events"):
		h.getEvents(w, r)
	case p == "/api/schema/example":
		h.getExampleData(w, r)
	case p == "/api/schema/validate":
		h.validate(w, r)
	case strings.HasPrefix(p, "/api/system/"):
		h.serveSystem(w, r)
	case strings.HasPrefix(p, "/api/configs"):
		h.handleConfig(w, r)
	case strings.HasPrefix(p, "/api/faker/tree"):
		h.handleFakerTree(w, r)
	case strings.HasPrefix(p, "/api/search"):
		h.getSearchResults(w, r)
	case h.fileServer != nil:
		if r.Method != "GET" {
			http.Error(w, fmt.Sprintf("method %v is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}
		if isAsset(r.URL.Path) {
			r.URL.Path = "/assets/" + filepath.Base(r.URL.Path)
		} else if isImage(r.URL.Path) {
			// don't change url
		} else {
			if len(h.path) > 0 || len(h.base) > 0 {
				base := h.path
				if len(h.base) > 0 {
					base = h.base
				}
				html := strings.Replace(h.index, "<base href=\"/\" />", fmt.Sprintf("<base href=\"%v/\" />", base), 1)
				html = h.replaceMeta(r.URL, html)

				_, err := w.Write([]byte(html))
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
	services = append(services, getHttpServices(h.app.ListHttp(), h.app.Monitor)...)
	services = append(services, getKafkaServices(h.app.Kafka, h.app.Monitor)...)
	services = append(services, getMailServices(h.app.Mail, h.app.Monitor)...)
	services = append(services, getLdapServices(h.app.Ldap, h.app.Monitor)...)
	slices.SortFunc(services, func(a interface{}, b interface{}) int {
		return compareService(a, b)
	})
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

	i := info{Version: h.app.Version, BuildTime: h.app.BuildTime, Search: searchInfo{Enabled: h.config.Search.Enabled}}
	if len(h.app.ListHttp()) > 0 {
		i.ActiveServices = append(i.ActiveServices, "http")
	}
	if len(h.app.Kafka.List()) > 0 {
		i.ActiveServices = append(i.ActiveServices, "kafka")
	}
	if len(h.app.Mail.List()) > 0 {
		i.ActiveServices = append(i.ActiveServices, "mail")
	}
	if len(h.app.Ldap.List()) > 0 {
		i.ActiveServices = append(i.ActiveServices, "ldap")
	}

	writeJsonBody(w, i)
}

func writeJsonBody(w http.ResponseWriter, i interface{}) {
	b, err := json.Marshal(i)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		log.Errorf("write response body failed: %v", err)
	}
}

func isAsset(path string) bool {
	return strings.Contains(path, "/assets/")
}

func isImage(path string) bool {
	str := filepath.Ext(path)
	switch str {
	case ".jpg", ".jpeg", ".png", ".svg":
		return true
	default:
		return false
	}
}

func compareService(a, b interface{}) int {
	return strings.Compare(getServiceName(a), getServiceName(b))
}

func getServiceName(a interface{}) string {
	switch v := a.(type) {
	case *httpSummary:
		return v.Name
	case *kafkaSummary:
		return v.Name
	case *ldapSummary:
		return v.Name
	case *mailSummary:
		return v.Name
	}
	return ""
}

func getPageInfo(r *http.Request) (index int, limit int, err error) {
	limit = 10

	sIndex := getQueryParamInsensitive(r.URL.Query(), searchIndex)
	if sIndex != "" {
		index, err = strconv.Atoi(sIndex)
		if err != nil {
			err = fmt.Errorf("invalid index value: %s", err)
		}
	}
	sLimit := getQueryParamInsensitive(r.URL.Query(), searchLimit)
	if sLimit != "" {
		limit, err = strconv.Atoi(sLimit)
		if err != nil {
			err = fmt.Errorf("invalid limit value: %s", err)
		}
	}
	return
}
