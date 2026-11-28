package common

import (
	"mokapi/smtp"
	"slices"
	"time"
)

// MailFilter limits a handler to one mail API (by title) and one mailbox. An empty field
// matches everything.
type MailFilter struct {
	Api     string
	Mailbox string
}

// MailEvent is a mail accepted by a mail API. Recipients is the SMTP envelope and decides
// where the mail is delivered, Message is the content that is stored afterwards.
type MailEvent struct {
	Api        string
	Recipients []string
	Message    *smtp.Message
}

// MailEventMessage is what event handlers see and may change. It is a copy of the mail, so
// that the stored mail changes only when ApplyTo is called.
type MailEventMessage struct {
	Api         string            `json:"api"`
	Server      string            `json:"server"`
	Sender      *smtp.Address     `json:"sender,omitempty"`
	From        []smtp.Address    `json:"from"`
	To          []smtp.Address    `json:"to"`
	ReplyTo     []smtp.Address    `json:"replyTo,omitempty"`
	Cc          []smtp.Address    `json:"cc,omitempty"`
	Bcc         []smtp.Address    `json:"bcc,omitempty"`
	MessageId   string            `json:"messageId"`
	InReplyTo   string            `json:"inReplyTo,omitempty"`
	Time        time.Time         `json:"time"`
	Subject     string            `json:"subject"`
	ContentType string            `json:"contentType,omitempty"`
	Encoding    string            `json:"encoding,omitempty"`
	Body        string            `json:"body"`
	Attachments []smtp.Attachment `json:"attachments,omitempty"`
}

func NewMailEventMessage(api string, m *smtp.Message) *MailEventMessage {
	e := &MailEventMessage{Api: api, Server: m.Server}
	e.copyContentFrom(m)
	return e
}

func (e *MailEventMessage) copyContentFrom(m *smtp.Message) {
	e.Sender = cloneAddress(m.Sender)
	e.From = slices.Clone(m.From)
	e.To = slices.Clone(m.To)
	e.ReplyTo = slices.Clone(m.ReplyTo)
	e.Cc = slices.Clone(m.Cc)
	e.Bcc = slices.Clone(m.Bcc)
	e.MessageId = m.MessageId
	e.InReplyTo = m.InReplyTo
	e.Time = m.Date
	e.Subject = m.Subject
	e.ContentType = m.ContentType
	e.Encoding = m.ContentTransferEncoding
	e.Body = m.Body
	e.Attachments = slices.Clone(m.Attachments)
}

// ApplyTo writes the content a handler may have changed back into the mail. The server and
// the API are not content, and the envelope is not part of the message at all.
func (e *MailEventMessage) ApplyTo(m *smtp.Message) {
	m.Sender = cloneAddress(e.Sender)
	m.From = slices.Clone(e.From)
	m.To = slices.Clone(e.To)
	m.ReplyTo = slices.Clone(e.ReplyTo)
	m.Cc = slices.Clone(e.Cc)
	m.Bcc = slices.Clone(e.Bcc)
	m.MessageId = e.MessageId
	m.InReplyTo = e.InReplyTo
	m.Date = e.Time
	m.Subject = e.Subject
	m.ContentType = e.ContentType
	m.ContentTransferEncoding = e.Encoding
	m.Body = e.Body
	m.Attachments = slices.Clone(e.Attachments)
}

func cloneAddress(a *smtp.Address) *smtp.Address {
	if a == nil {
		return nil
	}
	c := *a
	return &c
}
