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
}

func newMailSummary(mail *models.Mail) mailSummary {
	return mailSummary{
		Id:   mail.Id,
		From: mail.From,
		To:   mail.To,
		Time: mail.Time,
	}
}

func newMail(m *models.Mail) mailFull {
	return mailFull{
		Sender:  m.Sender,
		From:    m.From,
		ReplyTo: m.ReplyTo,
		To:      m.To,
		Cc:      m.Cc,
		Bcc:     m.Bcc,
		Time:    m.Time,

		ContentType: m.ContentType,
		Encoding:    m.Encoding,

		Subject:  m.Subject,
		TextBody: m.TextBody,
		HtmlBody: m.HtmlBody,
	}
}
