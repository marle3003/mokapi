package api

import (
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/http"
	"strings"
)

type mailSummary struct {
	service
}

type mailInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	Server      string `json:"server"`
}

func getMailServices(services map[string]*runtime.SmtpInfo, m *monitor.Monitor) []interface{} {
	result := make([]interface{}, 0, len(services))
	for _, hs := range services {
		s := service{
			Name:        hs.Info.Name,
			Description: hs.Info.Description,
			Version:     hs.Info.Version,
			Type:        ServiceSmtp,
		}

		if m != nil {
			s.Metrics = m.FindAll(metrics.ByNamespace("smtp"), metrics.ByLabel("service", hs.Info.Name))
		}

		result = append(result, &mailSummary{service: s})
	}
	return result
}

func (h *handler) getSmtpService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	s, ok := h.app.Smtp[name]
	if !ok {
		w.WriteHeader(404)
		return
	}

	result := &mailInfo{
		Name:        s.Info.Name,
		Description: s.Info.Description,
		Version:     s.Info.Version,
		Server:      s.Server,
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}
