package api

import (
	"encoding/json"
	"mokapi/models"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	application *models.Application
}

type serviceSummary struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
}

func New(a *models.Application) *Handler {
	return &Handler{application: a}
}

func (h Handler) CreateRoutes(router *mux.Router) {
	router.Methods(http.MethodGet).Path("/api/services/{name}").HandlerFunc(h.getService)
	router.Methods(http.MethodGet).Path("/api/services").HandlerFunc(h.getServices)
	router.Methods(http.MethodGet).Path("/api/dashboard").HandlerFunc(h.getDashboard)
}

func newServiceSummary(s *models.WebServiceInfo) serviceSummary {
	return serviceSummary{Name: s.Data.Name, Description: s.Data.Description, Version: s.Data.Version}
}

func (h Handler) getServices(rw http.ResponseWriter, request *http.Request) {
	services := make([]serviceSummary, 0)

	for _, s := range h.application.WebServices {
		services = append(services, newServiceSummary(s))
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	error := json.NewEncoder(rw).Encode(services)
	if error != nil {
		log.Errorf("Error in writing service response: %v", error.Error())
	}
}

func (h Handler) getDashboard(rw http.ResponseWriter, request *http.Request) {
	dashboard := newDashboard(h.application)

	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	error := json.NewEncoder(rw).Encode(dashboard)
	if error != nil {
		log.Errorf("Error in writing dashboard response: %v", error.Error())
	}
}
