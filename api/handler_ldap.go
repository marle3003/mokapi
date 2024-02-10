package api

import (
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/http"
	"strings"
)

type ldapSummary struct {
	service
}

type ldapInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	Server      string   `json:"server"`
	Configs     []config `json:"configs,omitempty"`
}

func getLdapServices(services map[string]*runtime.LdapInfo, m *monitor.Monitor) []interface{} {
	result := make([]interface{}, 0, len(services))
	for _, hs := range services {
		s := service{
			Name:        hs.Info.Name,
			Description: hs.Info.Description,
			Version:     hs.Info.Version,
			Type:        ServiceLdap,
		}

		if m != nil {
			s.Metrics = m.FindAll(metrics.ByNamespace("ldap"), metrics.ByLabel("service", hs.Info.Name))
		}

		result = append(result, &ldapSummary{service: s})
	}
	return result
}

func (h *handler) handleLdapService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	s, ok := h.app.Ldap[name]
	if !ok {
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
