package api

import (
	"mokapi/models"
	"net/mail"
	"time"
)

type mailSummary struct {
	Id      string          `json:"id"`
	From    []*mail.Address `json:"from"`
	To      []*mail.Address `json:"to"`
	Subject string          `json:"subject"`
	Time    time.Time       `json:"time"`
}

type mailFull struct {
	Sender  *mail.Address   `json:"sender"`
	From    []*mail.Address `json:"from"`
	ReplyTo []*mail.Address `json:"replyTo"`
	To      []*mail.Address `json:"to"`
	Cc      []*mail.Address `json:"cc"`
	Bcc     []*mail.Address `json:"bcc"`
	Time    time.Time       `json:"time"`

	ContentType string `json:"contentType"`
	Encoding    string `json:"encoding"`

	Subject  string `json:"subject"`
	TextBody string `json:"textBody"`
	HtmlBody string `json:"htmlBody"`

	Actions []workflowSummary `json:"actions"`
}

func newMailSummary(mail *models.Mail) mailSummary {
	return mailSummary{
		Id:      mail.Id,
		From:    mail.From,
		To:      mail.To,
		Subject: mail.Subject,
		Time:    mail.Time,
	}
}

func newMail(m *models.MailMetric) mailFull {
	r := mailFull{
		Sender:  m.Mail.Sender,
		From:    m.Mail.From,
		ReplyTo: m.Mail.ReplyTo,
		To:      m.Mail.To,
		Cc:      m.Mail.Cc,
		Bcc:     m.Mail.Bcc,
		Time:    m.Mail.Time,

		ContentType: m.Mail.ContentType,
		Encoding:    m.Mail.Encoding,

		Subject:  m.Mail.Subject,
		TextBody: m.Mail.TextBody,
		HtmlBody: m.Mail.HtmlBody,
	}

	for _, a := range m.Summary.Workflows {
		r.Actions = append(r.Actions, newActionSummary(a))
	}

	return r
}
