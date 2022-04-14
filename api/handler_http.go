package api

import (
	"mokapi/runtime"
	"mokapi/runtime/logs"
	"mokapi/runtime/monitor"
	"net/http"
	"strconv"
	"strings"
)

type httpSummary struct {
	Name        string `json:"name"`
	LastRequest int64  `json:"lastRequest"`
	Requests    int64  `json:"requests"`
	Errors      int64  `json:"errors"`
}

func (h *handler) getHttpServices(w http.ResponseWriter, r *http.Request) {
	result := getHttpServices(h.app.Http, h.app.Monitor.Http)
	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func (h *handler) getHttpService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	if s, ok := h.app.Http[name]; ok {
		w.Header().Set("Content-Type", "application/json")

		writeJsonBody(w, s)
	} else {
		w.WriteHeader(404)
	}
}

func (h *handler) getHttpRequests(w http.ResponseWriter, r *http.Request) {
	limit := 10
	s := r.URL.Query().Get("limit")
	if n, err := strconv.Atoi(s); err == nil {
		limit = n
	}
	service := r.URL.Query().Get("service")

	w.Header().Set("Content-Type", "application/json")
	log := h.app.Monitor.Http.Log
	if len(log) == 0 {
		writeJsonBody(w, log)
	} else if len(service) == 0 {
		n := limit
		if len(log) < n {
			n = len(log)
		}
		writeJsonBody(w, h.app.Monitor.Http.Log[:n])
	} else {
		result := make([]*logs.HttpLog, 0, limit)
		for _, item := range h.app.Monitor.Http.Log {
			if item.Service == service {
				result = append(result, item)
			}
			if len(result) >= limit {
				break
			}
		}
		writeJsonBody(w, result)
	}

}

func getHttpServices(services map[string]*runtime.HttpInfo, m *monitor.Http) []*httpSummary {
	result := make([]*httpSummary, 0, len(services))
	for _, hs := range services {

		result = append(result, &httpSummary{
			Name:        hs.Info.Name,
			Requests:    int64(m.RequestCounter.WithLabel(hs.Info.Name).Value()),
			Errors:      int64(m.RequestErrorCounter.WithLabel(hs.Info.Name).Value()),
			LastRequest: int64(m.LastRequest.WithLabel(hs.Info.Name).Value()),
		})
	}
	return result
}
