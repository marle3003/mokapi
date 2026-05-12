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
	NumRequests float64 `json:"ldap_requests_total"`
	LastRequest float64 `json:"ldap_request_timestamp"`
}

func getLdapServices(store *runtime.LdapStore, m *monitor.Monitor) []service {
	list := store.List()
	result := make([]service, 0, len(list))
	for _, li := range list {
		s := service{
			Name:        li.Info.Name,
			Description: li.Info.Description,
			Version:     li.Info.Version,
			Type:        ServiceLdap,
		}

		s.Metrics = ldapMetrics{
			NumRequests: m.Ldap.RequestCounter.Sum(metrics.NewQuery(metrics.ByLabel("service", li.Info.Name))),
			LastRequest: m.Ldap.LastRequest.Max(metrics.NewQuery(metrics.ByLabel("service", li.Info.Name))),
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
	result := &ldapInfo{
		Name:        s.Info.Name,
		Description: s.Info.Description,
		Version:     s.Info.Version,
		Server:      s.Address,
		Configs:     getConfigs(s.Configs()),
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}
