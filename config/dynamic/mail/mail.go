package mail

import (
	"mokapi/smtp"
	"time"
)

type Mail struct {
	Sender      Address
	From        []Address
	To          []Address
	ReplyTo     []Address
	Cc          []Address
	Bcc         []Address
	MessageId   string
	Date        time.Time
	Subject     string
	ContentType string
	Encoding    string
	Body        []string
}

type Address struct {
	Name    string
	Address string
}

func NewMail(msg *smtp.MailMessage) *Mail {
	m := &Mail{
		MessageId:   msg.MessageId,
		Date:        msg.Date,
		Subject:     msg.Subject,
		ContentType: msg.ContentType,
		Encoding:    msg.Encoding,
		Body:        msg.Body,
	}

	if msg.Sender != nil {
		m.Sender = Address{Name: msg.Sender.Name, Address: msg.Sender.Address}
	}

	for _, a := range msg.From {
		m.From = append(m.From, Address{Name: a.Name, Address: a.Address})
	}

	for _, a := range msg.To {
		m.To = append(m.To, Address{Name: a.Name, Address: a.Address})
	}

	for _, a := range msg.ReplyTo {
		m.ReplyTo = append(m.ReplyTo, Address{Name: a.Name, Address: a.Address})
	}

	for _, a := range msg.Cc {
		m.Cc = append(m.Cc, Address{Name: a.Name, Address: a.Address})
	}

	for _, a := range msg.Bcc {
		m.Bcc = append(m.Bcc, Address{Name: a.Name, Address: a.Address})
	}

	return m
}
