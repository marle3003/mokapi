package imap

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"mokapi/media"
	"mokapi/smtp"
	"net/mail"
	"net/textproto"
	"strings"
	"time"
)

type AppendOptions struct {
	Flags []Flag
	Date  time.Time
}

func (c *conn) handleAppend(tag string, d *Decoder) error {
	if c.state != AuthenticatedState && c.state != SelectedState {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Command is only valid in authenticated state",
		})
	}

	mailbox, err := d.SP().String()
	if err != nil {
		return err
	}
	d.SP()

	opt := AppendOptions{}
	if d.IsList() {
		err = d.List(func() error {
			f, err := d.ReadFlag()
			if err != nil {
				return err
			}
			opt.Flags = append(opt.Flags, Flag(f))
			return nil
		})
		if err != nil {
			return err
		}
		d.SP()
	}

	if err := d.expect("{"); err != nil {
		return err
	}
	size, err := d.Int64()
	if err != nil {
		return err
	}
	if err := d.SP().expect("}"); err != nil {
		return err
	}

	c.tpc.PrintfLine("+ Ready for literal data")

	r := io.LimitReader(c.tpc.R, size)
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	m, err := readMessage(bytes.NewReader(data))

	err = c.handler.Append(mailbox, m, opt, c.ctx)
	if err != nil {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   err.Error(),
		})
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "APPEND completed",
	})
}

func readMessage(r io.Reader) (*smtp.Message, error) {
	tc := textproto.NewReader(bufio.NewReader(r))
	header, err := tc.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	m := &smtp.Message{Headers: map[string]string{}}

	for key, val := range header {
		switch strings.ToLower(key) {
		case "sender":
			m.Sender, err = smtp.ParseAddress(val[0])
		case "from":
			m.From, err = smtp.ParseAddressList(val[0])
		case "to":
			m.To, err = smtp.ParseAddressList(val[0])
		case "reply-to":
			m.ReplyTo, err = smtp.ParseAddressList(val[0])
		case "cc":
			m.Cc, err = smtp.ParseAddressList(val[0])
		case "bcc":
			m.Bcc, err = smtp.ParseAddressList(val[0])
		case "message-id":
			m.MessageId = val[0]
		case "in-reply-to":
			m.InReplyTo = val[0]
		case "date":
			m.Date, err = mail.ParseDate(val[0])
		case "subject":
			m.Subject = val[0]
		case "content-type":
			m.ContentType = val[0]
		case "content-transfer-encoding":
			m.ContentTransferEncoding = val[0]
		}
		m.Headers[key] = val[0]
		m.Size += len(key) + 2 + len(val) + 2 // "Key: Value\r\n"
	}

	m.MessageId = header.Get("Message-ID")
	if len(m.MessageId) == 0 {
		return nil, fmt.Errorf("missing Message-ID")
	}

	if date := header.Get("Date"); date != "" {
		m.Date, err = mail.ParseDate(date)
		if err != nil {
			return nil, err
		}
	} else {
		m.Date = time.Now()
	}
	m.Size += len("Date") + 2 + len(m.Date.Format(DateTimeLayout)) + 2

	m.Size += 2 // Extra CRLF before body

	ct := media.ParseContentType(m.ContentType)
	switch {
	case ct.Key() == "multipart/mixed":
		r := multipart.NewReader(tc.DotReader(), ct.Parameters["boundary"])
		for {
			p, err := r.NextPart()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Errorf("smtp: unable to read message part: %v", err)
				break
			}

			if p.Header.Get("Content-Disposition") != "" {
				a, err := newAttachment(p)
				if err != nil {
					return nil, err
				}
				m.Attachments = append(m.Attachments, a)
			} else {
				m.ContentType = p.Header.Get("Content-Type")
				encoding := p.Header.Get("Content-Transfer-Encoding")
				b, err := parse(p, encoding)
				if err != nil {
					return nil, err
				}
				m.Body = string(b)
			}
		}
	// https://www.ietf.org/rfc/rfc2387.txt
	case ct.Key() == "multipart/related":
		r := multipart.NewReader(tc.R, ct.Parameters["boundary"])
		m.ContentType = strings.Trim(ct.Parameters["type"], "\"")
		first := true
		for {
			p, err := r.NextPart()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Errorf("smtp: unable to read message part: %v", err)
				break
			}

			if first {
				partContentType := p.Header.Get("Content-Type")
				partType := media.ParseContentType(partContentType)
				root := media.ParseContentType(m.ContentType)
				if !root.Match(partType) {
					log.Warnf("received mail message multipart/related '%v' type parameter and root body part differ", m.Subject)
				}
				encoding := p.Header.Get("Content-Transfer-Encoding")
				b, err := parse(p, encoding)
				if err != nil {
					return nil, err
				}
				m.Body = string(b)
			} else {
				a, err := newAttachment(p)
				if err != nil {
					return nil, err
				}
				m.Attachments = append(m.Attachments, a)
			}
			first = false
		}
	default:
		b, err := parse(tc.R, m.ContentTransferEncoding)
		if err != nil {
			return nil, err
		}
		if len(b) > 1 {
			m.Body = string(b[0 : len(b)-2]) // remove last \r\n
		}
		m.Size += len(b)
	}

	return m, nil
}

func parse(r io.Reader, encoding string) ([]byte, error) {
	switch strings.ToLower(encoding) {
	case "quoted-printable":
		r = quotedprintable.NewReader(r)
	case "base64":
		r = base64.NewDecoder(base64.StdEncoding, r)
	case "7bit", "8bit", "binary", "":
	default:
		return nil, fmt.Errorf("unsupported encoding %v", encoding)
	}

	var data bytes.Buffer
	_, err := data.ReadFrom(r)
	return data.Bytes(), err
}

func newAttachment(part *multipart.Part) (smtp.Attachment, error) {
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
	b, err := parse(part, encoding)
	if err != nil {
		return smtp.Attachment{}, err
	}
	att := smtp.Attachment{
		Name:                    name,
		ContentType:             part.Header.Get("Content-Type"),
		ContentTransferEncoding: encoding,
		ContentDescription:      part.Header.Get("Content-Description"),
		Disposition:             part.Header.Get("Content-Disposition"),
		Data:                    b,
	}

	contentId := part.Header.Get("Content-ID")
	if len(contentId) > 0 {
		att.ContentId = strings.Trim(contentId, "<>")
	}

	return att, nil
}
