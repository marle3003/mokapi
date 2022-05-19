package api

import (
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/http"
	"strings"
)

type httpSummary struct {
	service
}

func (h *handler) getHttpServices(w http.ResponseWriter, _ *http.Request) {
	result := getHttpServices(h.app.Http, h.app.Monitor)
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

func getHttpServices(services map[string]*runtime.HttpInfo, m *monitor.Monitor) []interface{} {
	result := make([]interface{}, 0, len(services))
	for _, hs := range services {

		result = append(result, &httpSummary{
			service: service{
				Name:        hs.Info.Name,
				Description: hs.Info.Description,
				Version:     hs.Info.Version,
				Type:        ServiceHttp,
				Metrics:     m.FindAll(metrics.ByNamespace("http"), metrics.ByLabel("service", hs.Info.Name)),
			},
		})
	}
	return result
}
