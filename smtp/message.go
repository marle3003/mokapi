package smtp

import (
	"bytes"
	"encoding/base64"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"mokapi/media"
	"net/mail"
	"net/textproto"
	"os"
	"strings"
	"time"
)

type Message struct {
	Sender      *Address     `json:"sender"`
	From        []Address    `json:"from"`
	To          []Address    `json:"to"`
	ReplyTo     []Address    `json:"replyTo"`
	Cc          []Address    `json:"cc"`
	Bcc         []Address    `json:"bcc"`
	MessageId   string       `json:"messageId"`
	InReplyTo   string       `json:"inReplyTo"`
	Time        time.Time    `json:"time"`
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

func (m *Message) readFrom(tc textproto.Reader) error {
	header, err := tc.ReadMIMEHeader()
	if err != nil {
		return err
	}

	if sender := header.Get("Sender"); sender != "" {
		m.Sender, err = parseAddress(sender)
		if err != nil {
			return err
		}
	}

	if from := header.Get("From"); from != "" {
		m.From, err = parseAddressList(header.Get("From"))
		if err != nil {
			return err
		}
	}

	if to := header.Get("To"); to != "" {
		m.To, err = parseAddressList(to)
		if err != nil {
			return err
		}
	}

	if replyTo := header.Get("Reply-To"); replyTo != "" {
		m.ReplyTo, err = parseAddressList(replyTo)
		if err != nil {
			return err
		}
	}

	if cc := header.Get("Cc"); cc != "" {
		m.Cc, err = parseAddressList(cc)
		if err != nil {
			return err
		}
	}

	if bcc := header.Get("Bcc"); bcc != "" {
		m.Bcc, err = parseAddressList(bcc)
		if err != nil {
			return err
		}
	}

	m.MessageId = header.Get("Message-ID")
	if len(m.MessageId) == 0 {
		m.MessageId = newMessageId()
	}
	m.InReplyTo = header.Get("In-Reply-To")

	if date := header.Get("Date"); date != "" {
		m.Time, err = mail.ParseDate(date)
		if err != nil {
			return err
		}
	}

	m.Subject = header.Get("Subject")
	m.ContentType = header.Get("Content-Type")
	m.Encoding = header.Get("Content-Transfer-Encoding")

	mime := media.ParseContentType(m.ContentType)
	switch {
	case mime.Key() == "multipart/mixed":
		r := multipart.NewReader(tc.DotReader(), mime.Parameters["boundary"])
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
		var r io.Reader
		switch strings.ToLower(m.Encoding) {
		case "quoted-printable":
			r = quotedprintable.NewReader(tc.DotReader())
		case "base64":
			r = base64.NewDecoder(base64.StdEncoding, tc.DotReader())
		case "7bit", "8bit", "binary", "":
			r = tc.DotReader()
		default:
			return fmt.Errorf("unsupported encoding %v", m.Encoding)
		}

		var data bytes.Buffer
		data.ReadFrom(r)
		if data.Len() > 0 {
			m.Body = data.String()[0 : data.Len()-1] // remove last \n
		}
	}

	return nil
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

func newMessageId() string {
	name, err := os.Hostname()
	if err != nil {
		name = "mokapi.io"
	}

	return fmt.Sprintf("%v-%v@%v", time.Now().Format("20060102-150405.000"), os.Getpid(), name)
}

func parseAddress(s string) (*Address, error) {
	a, err := mail.ParseAddress(s)
	if err != nil {
		return nil, err
	}
	return &Address{
		Name:    a.Name,
		Address: a.Address,
	}, nil
}

func parseAddressList(s string) ([]Address, error) {
	list, err := mail.ParseAddressList(s)
	if err != nil {
		return nil, err
	}
	var r []Address
	for _, a := range list {
		r = append(r, Address{
			Name:    a.Name,
			Address: a.Address,
		})
	}
	return r, nil
}
