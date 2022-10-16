package smtp

import (
	"io"
	"net/mail"
	"net/textproto"
	"strings"
	"time"
)

type Mail struct {
	From    string
	To      []string
	Message *MailMessage
}

type MailMessage struct {
	Sender      *mail.Address
	From        []*mail.Address
	To          []*mail.Address
	ReplyTo     []*mail.Address
	Cc          []*mail.Address
	Bcc         []*mail.Address
	MessageId   string
	Date        time.Time
	Subject     string
	ContentType string
	Encoding    string
	Body        io.Reader
}

func ReadMessage(tc textproto.Reader) (*MailMessage, error) {
	msg := &MailMessage{}
	header, err := tc.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	if sender := header.Get("Sender"); sender != "" {
		msg.Sender, err = mail.ParseAddress(sender)
		if err != nil {
			return nil, err
		}
	}

	if from := header.Get("From"); from != "" {
		msg.From, err = mail.ParseAddressList(header.Get("From"))
		if err != nil {
			return nil, err
		}
	}

	if to := header.Get("From"); to != "" {
		msg.To, err = mail.ParseAddressList(to)
		if err != nil {
			return nil, err
		}
	}

	if replyTo := header.Get("Reply-To"); replyTo != "" {
		msg.ReplyTo, err = mail.ParseAddressList(replyTo)
		if err != nil {
			return nil, err
		}
	}

	if cc := header.Get("Cc"); cc != "" {
		msg.Cc, err = mail.ParseAddressList(cc)
		if err != nil {
			return nil, err
		}
	}

	if bcc := header.Get("Bcc"); bcc != "" {
		msg.Bcc, err = mail.ParseAddressList(bcc)
		if err != nil {
			return nil, err
		}
	}

	msg.MessageId = header.Get("Message-ID")

	if date := header.Get("Date"); date != "" {
		msg.Date, err = mail.ParseDate(date)
		if err != nil {
			return nil, err
		}
	}

	msg.Subject = header.Get("Subject")
	msg.ContentType = header.Get("Content-Type")
	msg.Encoding = header.Get("Content-Transfer-Encoding")

	var body []string
	for {
		line, err := tc.ReadLine()
		if err != nil {
			return nil, err
		}
		if line == "." {
			break
		}
		body = append(body, line)
	}
	msg.Body = strings.NewReader(strings.Join(body, "\n"))

	return msg, nil
}
