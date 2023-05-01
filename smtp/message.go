package smtp

import (
	"bytes"
	"net/mail"
	"net/textproto"
	"time"
)

type Message struct {
	Sender      *mail.Address
	From        []*mail.Address
	To          []*mail.Address
	ReplyTo     []*mail.Address
	Cc          []*mail.Address
	Bcc         []*mail.Address
	MessageId   string
	InReplyTo   string
	Date        time.Time
	Subject     string
	ContentType string
	Encoding    string
	Body        string
}

func (m *Message) readFrom(tc textproto.Reader) error {
	header, err := tc.ReadMIMEHeader()
	if err != nil {
		return err
	}

	if sender := header.Get("Sender"); sender != "" {
		m.Sender, err = mail.ParseAddress(sender)
		if err != nil {
			return err
		}
	}

	if from := header.Get("From"); from != "" {
		m.From, err = mail.ParseAddressList(header.Get("From"))
		if err != nil {
			return err
		}
	}

	if to := header.Get("To"); to != "" {
		m.To, err = mail.ParseAddressList(to)
		if err != nil {
			return err
		}
	}

	if replyTo := header.Get("Reply-To"); replyTo != "" {
		m.ReplyTo, err = mail.ParseAddressList(replyTo)
		if err != nil {
			return err
		}
	}

	if cc := header.Get("Cc"); cc != "" {
		m.Cc, err = mail.ParseAddressList(cc)
		if err != nil {
			return err
		}
	}

	if bcc := header.Get("Bcc"); bcc != "" {
		m.Bcc, err = mail.ParseAddressList(bcc)
		if err != nil {
			return err
		}
	}

	m.MessageId = header.Get("Message-ID")
	m.InReplyTo = header.Get("In-Reply-To")

	if date := header.Get("Date"); date != "" {
		m.Date, err = mail.ParseDate(date)
		if err != nil {
			return err
		}
	}

	m.Subject = header.Get("Subject")
	m.ContentType = header.Get("Content-Type")
	m.Encoding = header.Get("Content-Transfer-Encoding")

	var data bytes.Buffer
	// read dot-encoded block from r
	data.ReadFrom(tc.DotReader())

	if data.Len() > 0 {
		m.Body = data.String()[0 : data.Len()-1] // remove last \n
	}

	return nil
}
