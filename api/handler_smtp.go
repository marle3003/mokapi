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
	Name          string      `json:"name"`
	Description   string      `json:"description,omitempty"`
	Version       string      `json:"version,omitempty"`
	Server        string      `json:"server"`
	Mailboxes     []mailboxes `json:"mailboxes,omitempty"`
	MaxRecipients int         `json:"maxRecipients,omitempty"`
	Rules         []rule      `json:"rules,omitempty"`
}

type mailboxes struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type rule struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Action    string `json:"action"`
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
		Name:          s.Info.Name,
		Description:   s.Info.Description,
		Version:       s.Info.Version,
		Server:        s.Server,
		MaxRecipients: s.MaxRecipients,
	}

	for _, m := range s.Mailboxes {
		result.Mailboxes = append(result.Mailboxes, mailboxes{
			Name:     m.Name,
			Username: m.Username,
			Password: m.Password,
		})
	}
	for _, r := range s.Rules {
		result.Rules = append(result.Rules, rule{
			Sender:    r.Sender,
			Recipient: r.Recipient,
			Subject:   r.Subject,
			Body:      r.Body,
			Action:    string(r.Action),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}
