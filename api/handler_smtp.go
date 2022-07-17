package api

import (
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/http"
	"strings"
)

type smtpSummary struct {
	service
}

func (h *handler) getSmtpService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	if s, ok := h.app.Smtp[name]; ok {
		w.Header().Set("Content-Type", "application/json")
		writeJsonBody(w, s)
	} else {
		w.WriteHeader(404)
	}
}

func getSmtpServices(services map[string]*runtime.SmtpInfo, m *monitor.Monitor) []interface{} {
	result := make([]interface{}, 0, len(services))
	for _, s := range services {
		summary := &smtpSummary{
			service: service{
				Name:        s.Info.Name,
				Description: s.Info.Description,
				Version:     s.Info.Version,
				Type:        ServiceSmtp,
				Metrics:     m.FindAll(metrics.ByNamespace("smtp"), metrics.ByLabel("service", s.Info.Name)),
			},
		}

		result = append(result, summary)
	}
	return result
}
