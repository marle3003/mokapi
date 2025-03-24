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

func getLdapServices(store *runtime.LdapStore, m *monitor.Monitor) []interface{} {
	list := store.List()
	result := make([]interface{}, 0, len(list))
	for _, hs := range list {
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
