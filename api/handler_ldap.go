package api

import (
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/http"
	"strings"
)

type ldapInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	Server      string   `json:"server"`
	Configs     []config `json:"configs,omitempty"`
}

type ldapMetrics struct {
	NumRequests float64 `json:"requests_total"`
	LastRequest float64 `json:"request_timestamp"`
}

func getLdapServices(store *runtime.LdapStore, m *monitor.Monitor) []service {
	list := store.List()
	result := make([]service, 0, len(list))
	for _, li := range list {
		cfg := li.Config()
		if cfg == nil {
			continue
		}
		s := service{
			Name:        cfg.Info.Name,
			Description: cfg.Info.Description,
			Version:     cfg.Info.Version,
			Type:        ServiceLdap,
		}

		s.Metrics = ldapMetrics{
			NumRequests: m.Ldap.RequestCounter.Sum(metrics.NewQuery(metrics.ByLabel("service", cfg.Info.Name))),
			LastRequest: m.Ldap.LastRequest.Max(metrics.NewQuery(metrics.ByLabel("service", cfg.Info.Name))),
		}

		result = append(result, s)
	}
	return result
}

func (h *handler) handleLdapService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	s := h.app.Ldap.Get(name)
	if s == nil {
		w.WriteHeader(404)
		return
	}
	cfg := s.Config()
	if cfg == nil {
		w.WriteHeader(404)
		return
	}
	result := &ldapInfo{
		Name:        cfg.Info.Name,
		Description: cfg.Info.Description,
		Version:     cfg.Info.Version,
		Server:      cfg.Address,
		Configs:     getConfigs(s.Configs()),
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}
