package smtp

import (
	"net/mail"
	"time"
)

type Mail struct {
	Id        string          `json:"id"`
	MessageId string          `json:"messageId"`
	Sender    *mail.Address   `json:"sender"`
	From      []*mail.Address `json:"from"`
	ReplyTo   []*mail.Address `json:"replyTo"`
	To        []*mail.Address `json:"to"`
	Cc        []*mail.Address `json:"cc"`
	Bcc       []*mail.Address `json:"bcc"`
	Time      time.Time       `json:"time"`

	Subject string `json:"subject"`

	ContentType string `json:"contentType"`
	Encoding    string `json:"encoding"`
	HtmlBody    string `json:"htmlBody"`
	TextBody    string `json:"textBody"`
	RawBody     string `json:"rawBody"`

	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
}

type MailWriter struct {
}
