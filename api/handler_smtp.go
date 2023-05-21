package api

import (
	"mokapi/config/dynamic/mail"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"net/http"
	"strings"
	"time"
)

type mailSummary struct {
	service
}

type mailInfo struct {
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	Version       string    `json:"version,omitempty"`
	Server        string    `json:"server"`
	Mailboxes     []mailbox `json:"mailboxes,omitempty"`
	MaxRecipients int       `json:"maxRecipients,omitempty"`
	Rules         []rule    `json:"rules,omitempty"`
}

type mailbox struct {
	Name     string     `json:"name"`
	Username string     `json:"username,omitempty"`
	Password string     `json:"password,omitempty"`
	Messages []*message `json:"mails,omitempty"`
}

type rule struct {
	Name           string          `json:"name"`
	Sender         string          `json:"sender"`
	Recipient      string          `json:"recipient"`
	Subject        string          `json:"subject"`
	Body           string          `json:"body"`
	Action         string          `json:"action"`
	RejectResponse *rejectResponse `json:"rejectResponse,omitempty"`
}

type rejectResponse struct {
	StatusCode         int     `json:"statusCode"`
	EnhancedStatusCode [3]int8 `json:"enhancedStatusCode"`
	Text               string  `json:"text"`
}

type message struct {
	Sender      *address     `json:"sender,omitempty"`
	From        []address    `json:"from"`
	To          []address    `json:"to"`
	ReplyTo     []address    `json:"replyTo,omitempty"`
	Cc          []address    `json:"cc,omitempty"`
	Bcc         []address    `json:"bbc,omitempty"`
	MessageId   string       `json:"messageId"`
	InReplyTo   string       `json:"inReplyTo,omitempty"`
	Date        time.Time    `json:"time"`
	Subject     string       `json:"subject"`
	ContentType string       `json:"contentType"`
	Encoding    string       `json:"encoding,omitempty"`
	Body        string       `json:"body"`
	Attachments []attachment `json:"attachments,omitempty"`
}

type address struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address"`
}

type attachment struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
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

func (h *handler) handleSmtpService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	n := len(segments)

	if n > 5 && segments[4] == "mails" {
		if n > 6 && segments[6] == "attachments" {
			h.getMailAttachment(w, segments[5], segments[7])
		} else {
			h.getMail(w, segments[5])
			return
		}
	}

	name := segments[4]

	s, ok := h.app.Smtp[name]
	if !ok {
		w.WriteHeader(404)
		return
	}

	if len(segments) > 6 && segments[5] == "mailboxes" {
		h.getMailbox(w, r, name, segments[6])
		return
	}

	result := &mailInfo{
		Name:          s.Info.Name,
		Description:   s.Info.Description,
		Version:       s.Info.Version,
		Server:        s.Server,
		MaxRecipients: s.MaxRecipients,
	}

	for _, m := range s.Store.Mailboxes {
		result.Mailboxes = append(result.Mailboxes, mailbox{
			Name:     m.Name,
			Username: m.Username,
			Password: m.Password,
		})
	}
	for _, r := range s.Rules {
		result.Rules = append(result.Rules, rule{
			Name:           r.Name,
			Sender:         r.Sender.String(),
			Recipient:      r.Recipient.String(),
			Subject:        r.Subject.String(),
			Body:           r.Body.String(),
			Action:         string(r.Action),
			RejectResponse: getRejectResponse(r),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func (h *handler) getMail(w http.ResponseWriter, messageId string) {
	var m *smtp.Message
	for _, s := range h.app.Smtp {
		m = s.Store.GetMail(messageId)
		if m != nil {
			break
		}
	}
	if m == nil {
		w.WriteHeader(404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, toMessage(m))
}

func (h *handler) getMailAttachment(w http.ResponseWriter, messageId, name string) {
	var m *smtp.Message
	for _, s := range h.app.Smtp {
		m = s.Store.GetMail(messageId)
		if m != nil {
			break
		}
	}
	if m == nil {
		w.WriteHeader(404)
		return
	}

	var att smtp.Attachment
	for _, a := range m.Attachments {
		if a.Name == name {
			att = a
		}
	}

	w.Header().Set("Content-Type", att.ContentType)
	w.Write(att.Data)
}

func (h *handler) getMailbox(w http.ResponseWriter, r *http.Request, service, name string) {
	s := h.app.Smtp[service]
	mb, ok := s.Store.Mailboxes[name]
	if !ok {
		w.WriteHeader(404)
		return
	}

	result := mailbox{
		Name:     mb.Name,
		Username: mb.Username,
		Password: mb.Password,
	}

	for _, m := range mb.Messages {
		result.Messages = append(result.Messages, toMessage(m))
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func getRejectResponse(r mail.Rule) *rejectResponse {
	if r.RejectResponse == nil {
		return nil
	}
	return &rejectResponse{
		StatusCode:         int(r.RejectResponse.StatusCode),
		EnhancedStatusCode: r.RejectResponse.EnhancedStatusCode,
		Text:               r.RejectResponse.Text,
	}
}

func toMessage(m *smtp.Message) *message {
	r := &message{
		From:        toAddress(m.From),
		To:          toAddress(m.To),
		ReplyTo:     toAddress(m.ReplyTo),
		Cc:          toAddress(m.Cc),
		Bcc:         toAddress(m.Bcc),
		MessageId:   m.MessageId,
		InReplyTo:   m.InReplyTo,
		Date:        m.Date,
		Subject:     m.Subject,
		ContentType: m.ContentType,
		Encoding:    m.Encoding,
		Body:        m.Body,
	}

	if m.Sender != nil {
		r.Sender = &address{
			Name:    m.Sender.Name,
			Address: m.Sender.Address,
		}
	}

	for _, a := range m.Attachments {
		r.Attachments = append(r.Attachments, attachment{
			Name:        a.Name,
			ContentType: a.ContentType,
			Size:        len(a.Data),
		})
	}

	return r
}

func toAddress(list []smtp.Address) []address {
	var r []address
	for _, a := range list {
		r = append(r, address{
			Name:    a.Name,
			Address: a.Address,
		})
	}
	return r
}
