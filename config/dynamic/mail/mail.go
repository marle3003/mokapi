package mail

import (
	"bytes"
	"encoding/base64"
	log "github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"mokapi/media"
	"mokapi/smtp"
	"strings"
	"time"
)

type Mail struct {
	Sender      *Address     `json:"sender"`
	From        []Address    `json:"from"`
	To          []Address    `json:"to"`
	ReplyTo     []Address    `json:"replyTo,omitempty"`
	Cc          []Address    `json:"cc,omitempty"`
	Bcc         []Address    `json:"bbc,omitempty"`
	MessageId   string       `json:"messageId"`
	InReplyTo   string       `json:"inReplyTo"`
	Date        time.Time    `json:"time"`
	Subject     string       `json:"subject"`
	ContentType string       `json:"contentType"`
	Encoding    string       `json:"encoding"`
	Body        string       `json:"body"`
	Attachments []Attachment `json:"attachments"`
}

type Address struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Attachment struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Data        []byte `json:"data"`
}

func NewMail(msg *smtp.Message) *Mail {
	m := &Mail{
		MessageId:   msg.MessageId,
		Date:        msg.Date,
		Subject:     msg.Subject,
		ContentType: msg.ContentType,
		Encoding:    msg.Encoding,
		InReplyTo:   msg.InReplyTo,
	}

	if msg.Sender != nil {
		m.Sender = &Address{Name: msg.Sender.Name, Address: msg.Sender.Address}
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

	mime := media.ParseContentType(m.ContentType)
	switch {
	case mime.Key() == "multipart/mixed":
		r := multipart.NewReader(strings.NewReader(msg.Body), mime.Parameters["boundary"])
		for {
			p, err := r.NextPart()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Errorf("smtp: unable to read message part: %v", err)
				break
			}
			if p.Header.Get("Content-Disposition") == "attachment" {
				m.Attachments = append(m.Attachments, newAttachment(p))
			} else {
				b, err := io.ReadAll(p)
				if err != nil {
					log.Errorf("smtp: unable to read part: %v", err)
				}
				m.Body += string(b)
			}
		}
	default:
		m.Body = msg.Body
	}

	return m
}

func newAttachment(part *multipart.Part) Attachment {
	contentType := part.Header.Get("Content-Type")
	name := part.FormName()
	if len(name) == 0 {
		name = part.FileName()
		if len(name) == 0 {
			m := media.ParseContentType(contentType)
			name = m.Parameters["name"]
		}
	}
	encoding := part.Header.Get("Content-Transfer-Encoding")
	var data []byte
	var err error
	b, err := io.ReadAll(part)
	if err != nil {
		log.Errorf("unable to read multipart: %v", err)
	}
	switch strings.ToUpper(encoding) {
	case "BASE64":

		data, err = base64.StdEncoding.DecodeString(string(b))
		if err != nil {
			log.Errorf("error on base64 decoding attachment: %v", err)
		}
	case "QUOTED-PRINTABLE":
		data, err = io.ReadAll(quotedprintable.NewReader(bytes.NewReader(b)))
		if err != nil {
			log.Println("Error decoding quoted-printable -", err)
		}
	}
	return Attachment{
		Name:        name,
		ContentType: part.Header.Get("Content-Type"),
		Data:        data,
	}
}
