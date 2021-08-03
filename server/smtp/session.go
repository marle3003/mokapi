package smtp

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime"
	"mokapi/models"
	"mokapi/providers/workflow"
	"mokapi/providers/workflow/event"
	"net/mail"
	"strings"
	"time"
)

type session struct {
	current  *models.Mail
	received chan *models.MailMetric
	wh       EventHandler
	state    *smtp.ConnectionState
}

func newSession(received chan *models.MailMetric, wh EventHandler, state *smtp.ConnectionState) *session {
	return &session{
		received: received,
		wh:       wh,
		state:    state,
	}
}

func (s *session) Reset() {
	if s.current != nil {
		summary, err := s.wh(event.WithSmtpEvent(event.SmtpEvent{Received: true, Address: s.state.LocalAddr.String()}), workflow.WithContext("mail", s.current))
		if err != nil {
			log.Errorf("error on smtp: %v", err)
		}

		if summary == nil {
			log.Debugf("no actions found")
		} else {
			log.WithField("action summary", summary).Debugf("executed actions")
		}

		s.received <- &models.MailMetric{Mail: s.current, Summary: summary}
	}
}

func (s *session) Logout() error {
	s.wh(event.WithSmtpEvent(event.SmtpEvent{Logout: true}))
	return nil
}

func (s *session) Mail(from string, opts smtp.MailOptions) error {
	return nil
}

func (s *session) Rcpt(to string) error {
	return nil
}

func (s *session) Data(r io.Reader) error {
	m, err := mail.ReadMessage(r)
	if err != nil {
		log.Errorf("error parsing mail: %v", err)
		return err
	}

	email := &models.Mail{}
	p := parser{}
	email.Sender = p.parseAddress(m.Header.Get("Sender"))
	email.From = p.parseAddressList(m.Header.Get("From"))
	email.ReplyTo = p.parseAddressList(m.Header.Get("Reply-To"))
	email.To = p.parseAddressList(m.Header.Get("To"))
	email.Cc = p.parseAddressList(m.Header.Get("Cc"))
	email.Bcc = p.parseAddressList(m.Header.Get("Bcc"))
	email.MessageId = p.parseId(m.Header.Get("Message-ID"))
	email.Subject = m.Header.Get("Subject")
	email.ContentType = m.Header.Get("Content-Type")
	email.Encoding = m.Header.Get("Content-Transfer-Encoding")
	email.Time = p.parseTime(m.Header.Get("Date"))

	var buf bytes.Buffer
	tee := io.TeeReader(m.Body, &buf)

	email.RawBody = p.parseString(tee)

	email.TextBody, email.HtmlBody, email.Attachments = p.parseBody(&buf, email.ContentType, email.Encoding)

	if p.err != nil {
		log.Errorf("error parsing mail: %v", err)
		return err
	}

	s.current = email

	return nil
}

type parser struct {
	err error
}

func (p parser) parseAddress(s string) (a *mail.Address) {
	if p.err != nil {
		return
	}

	a, p.err = mail.ParseAddress(s)
	return
}

func (p parser) parseAddressList(s string) (a []*mail.Address) {
	if p.err != nil {
		return
	}

	a, p.err = mail.ParseAddressList(s)
	return
}

func (p parser) parseId(s string) string {
	if p.err != nil {
		return ""
	}

	return strings.Trim(s, "<>")
}

func (p parser) parseString(r io.Reader) string {
	if p.err != nil {
		return ""
	}

	var b []byte
	b, p.err = ioutil.ReadAll(r)
	return string(b)
}

func (p parser) parseSubject(s string) string {
	r := make([]string, 0)

	for _, w := range strings.Split(s, " ") {
		d := new(mime.WordDecoder)
		w, p.err = d.Decode(w)
		if p.err != nil {
			return ""
		}
		r = append(r, w)
	}

	return strings.Join(r, "")
}

func (p parser) parseBody(r io.Reader, contentType, encoding string) (text, html string, attachments []models.Attachment) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	_ = params
	if err != nil {
		p.err = err
		return
	}

	r, err = decode(r, encoding)
	if err != nil {
		p.err = err
		return
	}

	switch mediaType {
	case "text/plain":
		b, err := ioutil.ReadAll(r)
		if err != nil {
			p.err = err
		}
		html = strings.TrimRight(string(b), "\r\n")
	case "text/html":
		b, err := ioutil.ReadAll(r)
		if err != nil {
			p.err = err
		}
		html = strings.TrimRight(string(b), "\r\n")
	default:
		b, err := ioutil.ReadAll(r)
		if err != nil {
			p.err = err
		}
		text = strings.TrimRight(string(b), "\r\n")
	}

	return
}

func (p parser) parseTime(s string) (t time.Time) {
	if p.err != nil || s == "" {
		return
	}

	formats := []string{
		time.RFC1123Z,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		time.RFC1123Z + " (MST)",
		"Mon, 2 Jan 2006 15:04:05 -0700 (MST)",
	}

	var err error
	for _, format := range formats {
		t, err = time.Parse(format, s)
		if err == nil {
			return
		}
	}

	p.err = err

	return
}

func decode(r io.Reader, encoding string) (io.Reader, error) {
	switch encoding {
	case "base64":
		b := base64.NewDecoder(base64.StdEncoding, r)
		return b, nil
	case "7bit":
	case "":
		return r, nil
	}

	return nil, fmt.Errorf("unsupported encoding: %s", encoding)
}
