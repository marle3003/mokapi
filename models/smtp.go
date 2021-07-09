package models

import (
	"mokapi/providers/workflow/runtime"
	"net/mail"
	"time"
)

type MailMetric struct {
	Mail    *Mail
	Summary *runtime.Summary
}

type Mail struct {
	Id        string
	MessageId string
	Sender    *mail.Address
	From      []*mail.Address
	ReplyTo   []*mail.Address
	To        []*mail.Address
	Cc        []*mail.Address
	Bcc       []*mail.Address
	Time      time.Time

	Subject string

	ContentType string
	Encoding    string
	HtmlBody    string
	TextBody    string
	RawBody     string

	Attachments []Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
}

func (m *Metrics) AddMail(mm *MailMetric) {
	mm.Mail.Id = newId(10)
	if mm.Mail.Time.IsZero() {
		mm.Mail.Time = time.Now()
	}
	m.TotalMails += 1
	if len(m.LastMails) > 10 {
		m.LastMails = m.LastMails[1:]
	}
	m.LastMails = append(m.LastMails, mm)
}
