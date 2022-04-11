package api

import (
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"net/http"
	"strings"
	"time"
)

type httpSummary struct {
	Name        string    `json:"name"`
	LastRequest time.Time `json:"lastRequest"`
	Requests    int64     `json:"requests"`
	Errors      int64     `json:"errors"`
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

func getHttpServices(services map[string]*runtime.HttpInfo, m *monitor.Http) []*httpSummary {
	result := make([]*httpSummary, 0, len(services))
	for _, hs := range services {

		result = append(result, &httpSummary{
			Name:     hs.Info.Name,
			Requests: int64(m.RequestCounter.WithLabel(hs.Info.Name).Value()),
			Errors:   int64(m.RequestErrorCounter.WithLabel(hs.Info.Name).Value()),
		})
	}
	return result
}
