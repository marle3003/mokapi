package api

import (
	"fmt"
	"mokapi/media"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"net/http"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
)

type mailSummary struct {
	service
}

type mailInfo struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Version     string       `json:"version,omitempty"`
	Contact     *contact     `json:"contact,omitempty"`
	Servers     []mailServer `json:"servers,omitempty"`
	Mailboxes   []mailbox    `json:"mailboxes,omitempty"`
	Rules       []rule       `json:"rules,omitempty"`
	Configs     []config     `json:"configs,omitempty"`
	Settings    settings     `json:"settings,omitempty"`
}

type mailServer struct {
	Host        string `json:"host"`
	Protocol    string `json:"protocol"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type settings struct {
	MaxRecipients     int  `json:"maxRecipients"`
	AutoCreateMailbox bool `json:"autoCreateMailbox"`
}

type mailbox struct {
	Name        string `json:"name"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	Description string `json:"description,omitempty"`
	NumMessages int    `json:"numMessages"`
}

type mailboxDetails struct {
	mailbox
	Folders []string `json:"folders,omitempty"`
}

type rule struct {
	Name           string          `json:"name"`
	Description    string          `json:"description,omitempty"`
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
	Message            string  `json:"message"`
}

type messageInfo struct {
	MessageId string    `json:"messageId"`
	From      []address `json:"from"`
	To        []address `json:"to"`
	Subject   string    `json:"subject"`
	Date      time.Time `json:"date"`
}

type message struct {
	Server                  string       `json:"server,omitempty"`
	Sender                  *address     `json:"sender,omitempty"`
	From                    []address    `json:"from"`
	To                      []address    `json:"to"`
	ReplyTo                 []address    `json:"replyTo,omitempty"`
	Cc                      []address    `json:"cc,omitempty"`
	Bcc                     []address    `json:"bbc,omitempty"`
	MessageId               string       `json:"messageId"`
	InReplyTo               string       `json:"inReplyTo,omitempty"`
	Date                    time.Time    `json:"date"`
	Subject                 string       `json:"subject"`
	ContentType             string       `json:"contentType,omitempty"`
	ContentTransferEncoding string       `json:"contentTransferEncoding,omitempty"`
	Body                    string       `json:"body"`
	Attachments             []attachment `json:"attachments,omitempty"`
	Size                    int          `json:"size"`
}

type address struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address"`
}

type attachment struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
	ContentId   string `json:"contentId,omitempty"`
}

func getMailServices(store *runtime.MailStore, m *monitor.Monitor) []interface{} {
	list := store.List()
	result := make([]interface{}, 0, len(list))
	for _, hs := range list {
		s := service{
			Name:        hs.Info.Name,
			Description: hs.Info.Description,
			Type:        ServiceMail,
			Version:     hs.Info.Version,
		}
		if hs.Info.Contact != nil {
			s.Contact = &contact{
				Name:  hs.Info.Contact.Name,
				Url:   hs.Info.Contact.Url,
				Email: hs.Info.Contact.Email,
			}
		}

		if m != nil {
			s.Metrics = m.FindAll(metrics.ByNamespace("mail"), metrics.ByLabel("service", hs.Info.Name))
		}

		result = append(result, &mailSummary{service: s})
	}
	return result
}

func (h *handler) handleMailService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	n := len(segments)

	// segment mails is deprecated
	if n > 4 && (segments[4] == "mails" || segments[4] == "messages") {
		if n > 6 && segments[6] == "attachments" {
			h.getMailAttachment(w, segments[5], segments[7])
			return
		} else {
			h.getMail(w, segments[5])
			return
		}
	}

	name := segments[4]

	s := h.app.Mail.Get(name)
	if s == nil {
		w.WriteHeader(404)
		return
	}

	if len(segments) > 7 && segments[7] == "messages" {
		if len(segments) > 6 {
			h.getMailboxMessages(w, r, name, segments[6])
		}
		return
	}

	if len(segments) > 5 && segments[5] == "mailboxes" {
		if len(segments) == 7 {
			h.getMailbox(w, r, name, segments[6])
		} else if len(segments) == 6 {
			h.getMailboxes(w, r, name)
		} else {
			w.WriteHeader(401)
		}
		return
	}

	result := &mailInfo{
		Name:        s.Info.Name,
		Description: s.Info.Description,
		Version:     s.Info.Version,
		Configs:     getConfigs(s.Configs()),
	}

	if s.Info.Contact != nil {
		result.Contact = &contact{
			Name:  s.Info.Contact.Name,
			Url:   s.Info.Contact.Url,
			Email: s.Info.Contact.Email,
		}
	}

	for n, ser := range s.Servers {
		result.Servers = append(result.Servers, mailServer{
			Host:        ser.Host,
			Protocol:    ser.Protocol,
			Name:        n,
			Description: ser.Description,
		})
	}

	if s.Store.Settings != nil {
		result.Settings.MaxRecipients = s.Store.Settings.MaxRecipients
		result.Settings.AutoCreateMailbox = s.Store.Settings.AutoCreateMailbox
	}

	for mName, m := range s.Store.Mailboxes {
		result.Mailboxes = append(result.Mailboxes, mailbox{
			Name:        mName,
			Username:    m.Username,
			Password:    m.Password,
			Description: m.Description,
			NumMessages: m.NumMessages(),
		})
	}
	for _, r := range s.Rules {
		result.Rules = append(result.Rules, rule{
			Name:           r.Name,
			Description:    r.Description,
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
	for _, s := range h.app.Mail.List() {
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
	for _, s := range h.app.Mail.List() {
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
		if a.Name == name || a.ContentId == name {
			att = a
			break
		}
	}

	contentType := att.ContentType
	if contentType == "application/octet-stream" && filepath.Ext(name) == "" {
		contentType = http.DetectContentType(att.Data)
		ct := media.ParseContentType(contentType)
		if len(ct.Subtype) > 0 {
			name = fmt.Sprintf("%v.%v", name, ct.Subtype)
		}
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", name))
	w.Header().Set("Content-Type", att.ContentType)
	_, _ = w.Write(att.Data)
}

func (h *handler) getMailboxes(w http.ResponseWriter, _ *http.Request, service string) {
	s := h.app.Mail.Get(service)
	var result []mailbox

	var names []string
	for name := range s.Store.Mailboxes {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		mb := s.Store.Mailboxes[name]

		result = append(result, mailbox{
			Name:        name,
			Username:    mb.Username,
			Password:    mb.Password,
			NumMessages: mb.NumMessages(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func (h *handler) getMailbox(w http.ResponseWriter, _ *http.Request, service, name string) {
	s := h.app.Mail.Get(service)
	mb, ok := s.Store.Mailboxes[name]
	if !ok {
		w.WriteHeader(404)
		return
	}

	result := mailboxDetails{
		mailbox: mailbox{
			Name:        name,
			Username:    mb.Username,
			Password:    mb.Password,
			NumMessages: mb.NumMessages(),
		},
	}

	for folder := range mb.Folders {
		result.Folders = append(result.Folders, folder)
	}
	sort.Strings(result.Folders)

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func (h *handler) getMailboxMessages(w http.ResponseWriter, r *http.Request, service, name string) {
	s := h.app.Mail.Get(service)
	mb, ok := s.Store.Mailboxes[name]
	if !ok {
		w.WriteHeader(404)
		return
	}

	var messages []*mail.Mail

	path := getQueryParamInsensitive(r.URL.Query(), "folder")
	folders := mb.List(path)

	index, limit, err := getPageInfo(r)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	for _, f := range folders {
		messages = append(messages, f.ListMessages()...)
	}

	slices.SortFunc(messages, func(a, b *mail.Mail) int {
		return a.Date.Compare(b.Date) * -1
	})

	from := index * limit
	var result []messageInfo
	if from < len(messages) {
		limit = min(limit, len(messages))
		for i := from; i < limit; i++ {
			msg := messages[i].Message
			result = append(result, messageInfo{
				MessageId: msg.MessageId,
				From:      toAddress(msg.From),
				To:        toAddress(msg.To),
				Subject:   msg.Subject,
				Date:      msg.Date,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func getRejectResponse(r *mail.Rule) *rejectResponse {
	if r.RejectResponse == nil {
		return nil
	}
	return &rejectResponse{
		StatusCode:         int(r.RejectResponse.StatusCode),
		EnhancedStatusCode: r.RejectResponse.EnhancedStatusCode,
		Message:            r.RejectResponse.Message,
	}
}

func toMessage(m *smtp.Message) *message {
	r := &message{
		Server:                  m.Server,
		From:                    toAddress(m.From),
		To:                      toAddress(m.To),
		ReplyTo:                 toAddress(m.ReplyTo),
		Cc:                      toAddress(m.Cc),
		Bcc:                     toAddress(m.Bcc),
		MessageId:               m.MessageId,
		InReplyTo:               m.InReplyTo,
		Date:                    m.Date,
		Subject:                 m.Subject,
		ContentType:             m.ContentType,
		ContentTransferEncoding: m.ContentTransferEncoding,
		Body:                    m.Body,
		Size:                    m.Size,
	}

	if m.Sender != nil {
		r.Sender = &address{
			Name:    m.Sender.Name,
			Address: m.Sender.Address,
		}
	}

	for _, a := range m.Attachments {
		name := a.Name
		if len(a.ContentId) > 0 {
			name = a.ContentId
		}
		r.Attachments = append(r.Attachments, attachment{
			Name:        name,
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
