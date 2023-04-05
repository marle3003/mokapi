package smtp

import (
	"net/mail"
	"time"
)

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
