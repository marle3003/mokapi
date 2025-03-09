package smtp

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"mokapi/media"
	"net/http"
	"net/mail"
	"net/textproto"
	"os"
	"strings"
	"time"
)

const DateTimeLayout = "02-Jan-2006 15:04:05 -0700"

type Message struct {
	Server                  string       `json:"server"`
	Sender                  *Address     `json:"sender"`
	From                    []Address    `json:"from"`
	To                      []Address    `json:"to"`
	ReplyTo                 []Address    `json:"replyTo"`
	Cc                      []Address    `json:"cc"`
	Bcc                     []Address    `json:"bcc"`
	MessageId               string       `json:"messageId"`
	InReplyTo               string       `json:"inReplyTo"`
	Time                    time.Time    `json:"time"`
	Subject                 string       `json:"subject"`
	ContentType             string       `json:"contentType"`
	Encoding                string       `json:"encoding"`
	ContentTransferEncoding string       `json:"contentTransferEncoding"`
	Body                    string       `json:"body"`
	Attachments             []Attachment `json:"attachments"`
	Size                    int
	Headers                 map[string]string
}

type Address struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Attachment struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Disposition string `json:"disposition"`
	Data        []byte `json:"data"`
	ContentId   string `json:"contentId"`
	Header      map[string]string
}

func (m *Message) readFrom(tc textproto.Reader) error {
	header, err := tc.ReadMIMEHeader()
	if err != nil {
		return err
	}

	m.Headers = map[string]string{}
	for key, val := range header {
		switch strings.ToLower(key) {
		case "sender":
			m.Sender, err = parseAddress(val[0])
		case "from":
			m.From, err = parseAddressList(val[0])
		case "to":
			m.To, err = parseAddressList(val[0])
		case "reply-to":
			m.ReplyTo, err = parseAddressList(val[0])
		case "cc":
			m.Cc, err = parseAddressList(val[0])
		case "bcc":
			m.Bcc, err = parseAddressList(val[0])
		case "message-id":
			m.MessageId = val[0]
		case "in-reply-to":
			m.InReplyTo = val[0]
		case "date":
			m.Time, err = mail.ParseDate(val[0])
		case "subject":
			m.Subject = val[0]
		case "content-type":
			m.ContentType = val[0]
		case "encoding":
			m.Encoding = val[0]
		case "content-transfer-encoding":
			m.ContentTransferEncoding = val[0]
		}
		m.Headers[key] = val[0]
		m.Size += len(key) + 2 + len(val) + 2 // "Key: Value\r\n"
	}

	m.MessageId = header.Get("Message-ID")
	if len(m.MessageId) == 0 {
		m.MessageId = newMessageId()
		m.Size += len("Message-ID") + 2 + len(m.MessageId) + 2
	}

	if date := header.Get("Date"); date != "" {
		m.Time, err = mail.ParseDate(date)
		if err != nil {
			return err
		}
	} else {
		m.Time = time.Now()
	}
	m.Size += len("Date") + 2 + len(m.Time.Format(DateTimeLayout)) + 2

	m.Size += 2 // Extra CRLF before body

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

			if p.Header.Get("Content-Disposition") != "" {
				a, err := newAttachment(p)
				if err != nil {
					return err
				}
				m.Attachments = append(m.Attachments, a)
			} else {
				m.ContentType = p.Header.Get("Content-Type")
				encoding := p.Header.Get("Content-Transfer-Encoding")
				b, err := parse(p, encoding)
				if err != nil {
					return err
				}
				m.Body = string(b)
			}
		}
	// https://www.ietf.org/rfc/rfc2387.txt
	case mime.Key() == "multipart/related":
		r := multipart.NewReader(tc.DotReader(), mime.Parameters["boundary"])
		m.ContentType = strings.Trim(mime.Parameters["type"], "\"")
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
					return err
				}
				m.Body = string(b)
			} else {
				a, err := newAttachment(p)
				if err != nil {
					return err
				}
				m.Attachments = append(m.Attachments, a)
			}
			first = false
		}
	default:
		b, err := parse(tc.DotReader(), m.Encoding)
		if err != nil {
			return err
		}
		if len(b) > 0 {
			m.Body = string(b[0 : len(b)-1]) // remove last \n
		}
		m.Size += len(b)
	}

	return nil
}

func (m *Message) WriteTo(w io.WriteCloser) error {
	text := textproto.NewWriter(bufio.NewWriter(w))
	var err error

	if len(m.MessageId) > 0 {
		err = text.PrintfLine("Message-ID: %s", m.MessageId)
		if err != nil {
			return err
		}
	}
	if len(m.Sender.Address) > 0 {
		err = text.PrintfLine("Sender: %s", m.Sender.Address)
		if err != nil {
			return err
		}
	}

	var to []string
	for _, addr := range m.To {
		to = append(to, addr.String())
	}
	if len(to) > 0 {
		err = text.PrintfLine("From: %s", strings.Join(to, ","))
		if err != nil {
			return err
		}
	}

	var cc []string
	for _, addr := range m.Cc {
		cc = append(cc, addr.String())
	}
	if len(cc) > 0 {
		err = text.PrintfLine("From: %s", strings.Join(cc, ","))
		if err != nil {
			return err
		}
	}

	var bcc []string
	for _, addr := range m.Bcc {
		bcc = append(bcc, addr.String())
	}
	if len(bcc) > 0 {
		err = text.PrintfLine("From: %s", strings.Join(bcc, ","))
		if err != nil {
			return err
		}
	}

	var replyTo []string
	for _, addr := range m.ReplyTo {
		replyTo = append(replyTo, addr.String())
	}
	if len(replyTo) > 0 {
		err = text.PrintfLine("From: %s", strings.Join(replyTo, ","))
		if err != nil {
			return err
		}
	}

	if len(m.InReplyTo) > 0 {
		err = text.PrintfLine("In-Reply-To: %s", m.InReplyTo)
		if err != nil {
			return err
		}
	}

	err = text.PrintfLine("Subject: %v", m.Subject)
	if err != nil {
		return err
	}

	if len(m.Attachments) == 0 {
		if len(m.ContentType) == 0 {
			m.ContentType = "text/plain; charset=UTF-8"
		}
		err = text.PrintfLine("Content-Type: %v", m.ContentType)
		if err != nil {
			return err
		}

		if len(m.Encoding) > 0 {
			_ = text.PrintfLine("Content-Transfer-Encoding: %s", m.Encoding)
			if err != nil {
				return err
			}
		}

		_, err = w.Write([]byte(fmt.Sprintf("\n%s", m.Body)))
		if err != nil {
			return err
		}
	} else {
		for _, att := range m.Attachments {
			err = att.WriteTo(text, m)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

func newAttachment(part *multipart.Part) (Attachment, error) {
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
		return Attachment{}, err
	}
	att := Attachment{
		Name:        name,
		ContentType: part.Header.Get("Content-Type"),
		Disposition: part.Header.Get("Content-Disposition"),
		Data:        b,
	}

	contentId := part.Header.Get("Content-ID")
	if len(contentId) > 0 {
		att.ContentId = strings.Trim(contentId, "<>")
	}

	return att, nil
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

func (a *Address) String() string {
	s := fmt.Sprintf("<%s>", a.Address)
	if a.Name == "" {
		return s
	}

	// Text in an encoded-word in a display-name must not contain certain
	// characters like quotes or parentheses (see RFC 2047 section 5.3).
	// When this is the case encode the name using base64 encoding.
	if strings.ContainsAny(a.Name, "\"#$%&'(),.:;<>@[]^`{|}~") {
		return fmt.Sprintf("%s %s", mime.BEncoding.Encode("utf-8", a.Name), s)
	}
	return fmt.Sprintf("%s %s", mime.QEncoding.Encode("utf-8", a.Name), s)
}

func (a *Attachment) WriteTo(w *textproto.Writer, m *Message) error {
	boundary := fmt.Sprintf("boundary_%v", uuid.New().String())
	err := w.PrintfLine("Content-Type: multipart/mixed; boundary=%v", boundary)
	if err != nil {
		return err
	}
	err = w.PrintfLine("")
	if err != nil {
		return err
	}

	err = w.PrintfLine("--%v", boundary)
	if err != nil {
		return err
	}
	if len(m.ContentType) == 0 {
		m.ContentType = "text/plain; charset=UTF-8"
	}
	err = w.PrintfLine("Content-Type: %v", m.ContentType)
	if err != nil {
		return err
	}

	if len(m.Encoding) > 0 {
		err = w.PrintfLine("Content-Transfer-Encoding: %s", m.Encoding)
		if err != nil {
			return err
		}
	}
	_, err = w.W.WriteString(fmt.Sprintf("\n%s\n", m.Body))
	if err != nil {
		return err
	}

	for _, attach := range m.Attachments {
		content := attach.Data
		name := attach.Name
		contentType := attach.ContentType
		if len(contentType) == 0 {
			contentType = http.DetectContentType(content)
		}
		if len(name) > 0 {
			contentType += fmt.Sprintf("; name=%s", name)
		}

		err = w.PrintfLine("--%v", boundary)
		if err != nil {
			return err
		}
		err = w.PrintfLine("Content-Type: %v", contentType)
		if err != nil {
			return err
		}
		err = w.PrintfLine("Content-Transfer-Encoding: base64")
		if err != nil {
			return err
		}
		disposition := attach.Disposition
		if len(disposition) == 0 {
			disposition = "attachment"
		}
		err = w.PrintfLine("Content-Disposition: %v", disposition)
		if err != nil {
			return err
		}
		err = w.PrintfLine("")
		if err != nil {
			return err
		}
		data := base64.StdEncoding.EncodeToString(content)
		_, err = w.W.WriteString(data)
		if err != nil {
			return err
		}
		_, err = w.W.WriteRune('\n')
		if err != nil {
			return err
		}
	}

	err = w.PrintfLine("--%v--", boundary)
	if err != nil {
		return err
	}
	err = w.PrintfLine("")

	return err
}
