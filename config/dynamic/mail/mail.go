package mail

import (
	"mime/multipart"
	"mokapi/media"
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
	Body        string
	Attachment  []Attachement
}

type Address struct {
	Name    string
	Address string
}

type Attachement struct {
	Name        string
	ContentType string
	p           *multipart.Part
}

func NewMail(msg *smtp.MailMessage) *Mail {
	m := &Mail{
		MessageId:   msg.MessageId,
		Date:        msg.Date,
		Subject:     msg.Subject,
		ContentType: msg.ContentType,
		Encoding:    msg.Encoding,
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

func (m *Mail) Write(b []byte) {

}

func newAttachment(part *multipart.Part) Attachement {
	contentType := part.Header.Get("Content-Type")
	name := part.FormName()
	if len(name) == 0 {
		name = part.FileName()
		if len(name) == 0 {
			m := media.ParseContentType(contentType)
			name = m.Parameters["name"]
		}
	}
	return Attachement{
		Name:        name,
		ContentType: part.Header.Get("Content-Type"),
		p:           part,
	}
}

func (a *Attachement) Read(b []byte) (int, error) {
	return a.p.Read(b)
}
