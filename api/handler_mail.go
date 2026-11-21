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
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

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
	Service string       `json:"service"`
	Data    *messageData `json:"data"`
}

type messageData struct {
	Server                  string       `json:"server,omitempty"`
	Sender                  *address     `json:"sender,omitempty"`
	From                    []address    `json:"from"`
	To                      []address    `json:"to"`
	ReplyTo                 []address    `json:"replyTo,omitempty"`
	Cc                      []address    `json:"cc,omitempty"`
	Bcc                     []address    `json:"bcc,omitempty"`
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

type mailMetrics struct {
	NumMails float64 `json:"mails_total"`
	LastMail float64 `json:"mail_timestamp"`
}

func getMailServices(store *runtime.MailStore, m *monitor.Monitor) []service {
	list := store.List()
	result := make([]service, 0, len(list))
	for _, mi := range list {
		cfg := mi.Config()
		if cfg == nil {
			continue
		}
		s := service{
			Name:        cfg.Info.Name,
			Description: cfg.Info.Description,
			Type:        ServiceMail,
			Version:     cfg.Info.Version,
			Items:       getMailboxItems(mi.Store),
		}
		if cfg.Info.Contact != nil {
			s.Contact = &contact{
				Name:  cfg.Info.Contact.Name,
				Url:   cfg.Info.Contact.Url,
				Email: cfg.Info.Contact.Email,
			}
		}

		s.Metrics = mailMetrics{
			NumMails: m.Mail.Mails.Sum(metrics.NewQuery(metrics.ByLabel("service", cfg.Info.Name))),
			LastMail: m.Mail.LastMail.Max(metrics.NewQuery(metrics.ByLabel("service", cfg.Info.Name))),
		}

		result = append(result, s)
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

	// a service without a config was removed while this request is running
	s := h.app.Mail.Get(name)
	if s == nil {
		w.WriteHeader(404)
		return
	}
	cfg, files := s.Snapshot()
	if cfg == nil {
		w.WriteHeader(404)
		return
	}

	if len(segments) > 7 && segments[7] == "messages" {
		if len(segments) > 6 {
			h.getMailboxMessages(w, r, s.Store, segments[6])
		}
		return
	}

	if len(segments) > 5 && segments[5] == "mailboxes" {
		if len(segments) == 7 {
			h.getMailbox(w, s.Store, segments[6])
		} else if len(segments) == 6 {
			h.getMailboxes(w, s.Store)
		} else {
			w.WriteHeader(401)
		}
		return
	}

	result := &mailInfo{
		Name:        cfg.Info.Name,
		Description: cfg.Info.Description,
		Version:     cfg.Info.Version,
		Configs:     getConfigs(files),
	}

	if cfg.Info.Contact != nil {
		result.Contact = &contact{
			Name:  cfg.Info.Contact.Name,
			Url:   cfg.Info.Contact.Url,
			Email: cfg.Info.Contact.Email,
		}
	}

	for n, ser := range cfg.Servers {
		result.Servers = append(result.Servers, mailServer{
			Host:        ser.Host,
			Protocol:    ser.Protocol,
			Name:        n,
			Description: ser.Description,
		})
	}

	if cfg.Settings != nil {
		result.Settings.MaxRecipients = cfg.Settings.MaxRecipients
		result.Settings.AutoCreateMailbox = cfg.Settings.AutoCreateMailbox
	} else {
		result.Settings.AutoCreateMailbox = true
	}

	for _, m := range s.Mailboxes() {
		username, password := m.Credentials()
		result.Mailboxes = append(result.Mailboxes, mailbox{
			Name:        m.Name,
			Username:    username,
			Password:    password,
			Description: m.Description(),
			NumMessages: m.NumMessages(),
		})
	}
	for _, r := range cfg.Rules {
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
	var store *mail.Store
	for _, s := range h.app.Mail.List() {
		m = s.Store.GetMail(messageId)
		if m != nil {
			store = s.Store
			break
		}
	}
	if m == nil {
		w.WriteHeader(404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, message{
		Service: store.Name,
		Data:    toMessage(m),
	})
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

func (h *handler) getMailboxes(w http.ResponseWriter, store *mail.Store) {
	var result []mailbox
	for _, mb := range store.Mailboxes() {
		username, password := mb.Credentials()
		result = append(result, mailbox{
			Name:        mb.Name,
			Username:    username,
			Password:    password,
			NumMessages: mb.NumMessages(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func (h *handler) getMailbox(w http.ResponseWriter, store *mail.Store, name string) {
	mb, ok := store.Mailbox(name)
	if !ok {
		w.WriteHeader(404)
		return
	}

	username, password := mb.Credentials()
	result := mailboxDetails{
		mailbox: mailbox{
			Name:        name,
			Username:    username,
			Password:    password,
			NumMessages: mb.NumMessages(),
		},
		Folders: mb.FolderNames(),
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func (h *handler) getMailboxMessages(w http.ResponseWriter, r *http.Request, store *mail.Store, name string) {
	mb, ok := store.Mailbox(name)
	if !ok {
		w.WriteHeader(404)
		return
	}

	index, limit, err := getPageInfo(r)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	path := getQueryParamInsensitive(r.URL.Query(), "folder")
	messages := mb.Messages(path)

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

func toMessage(m *smtp.Message) *messageData {
	subject, err := smtp.DecodeHeaderValue(m.Subject)
	if err != nil {
		log.Printf("failed to decode subject '%s': %v", m.Subject, err)
		subject = m.Subject
	}
	r := &messageData{
		Server:                  m.Server,
		From:                    toAddress(m.From),
		To:                      toAddress(m.To),
		ReplyTo:                 toAddress(m.ReplyTo),
		Cc:                      toAddress(m.Cc),
		Bcc:                     toAddress(m.Bcc),
		MessageId:               m.MessageId,
		InReplyTo:               m.InReplyTo,
		Date:                    m.Date,
		Subject:                 subject,
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
		name, err := smtp.DecodeHeaderValue(a.Name)
		if err != nil {
			log.Printf("failed to decode address '%v': %v", a.Name, err)
			name = a.Name
		}
		r = append(r, address{
			Name:    name,
			Address: a.Address,
		})
	}
	return r
}
