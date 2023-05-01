package js

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/dop251/goja"
	"github.com/google/uuid"
	"mokapi/engine/common"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type mailModule struct {
	rt         *goja.Runtime
	host       common.Host
	workingDir string
}

type Mail struct {
	MessageId   string        `json:"messageId"`
	Sender      string        `json:"sender"`
	From        interface{}   `json:"from"`
	To          interface{}   `json:"to"`
	Cc          interface{}   `json:"cc"`
	Bcc         interface{}   `json:"bcc"`
	ReplyTo     interface{}   `json:"replyTo"`
	InReplyTo   string        `json:"inReplyTo"`
	Subject     string        `json:"subject"`
	Body        string        `json:"body"`
	ContentType string        `json:"contentType"`
	Encoding    string        `json:"encoding"`
	Attachments []*Attachment `json:"attachments"`
}

type Attachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	Path        string `json:"path"`
	ContentType string `json:"contentType"`
}

func newMail(h common.Host, rt *goja.Runtime, workingDir string) interface{} {
	return &mailModule{rt: rt, host: h, workingDir: workingDir}
}

func (m *mailModule) Send(addr string, msg *Mail) {
	var body bytes.Buffer
	w := textproto.NewWriter(bufio.NewWriter(&body))

	_ = w.PrintfLine("Date: %v", time.Now().Format(time.RFC1123Z))

	if len(msg.MessageId) > 0 {
		_ = w.PrintfLine("Message-ID: %s", msg.MessageId)
	}

	if len(msg.Sender) > 0 {
		_ = w.PrintfLine("Sender: %s", msg.Sender)
	}

	from, fromHeader, err := parseAddressList(toList(msg.From))
	if err != nil {
		toJsError(m.rt, err)
	}
	_ = w.PrintfLine("From: %s", strings.Join(fromHeader, ","))

	to, toHeader, err := parseAddressList(toList(msg.To))
	if err != nil {
		toJsError(m.rt, err)
	}
	if len(toHeader) > 0 {
		_ = w.PrintfLine("To: %s", strings.Join(toHeader, ","))
	}

	cc, ccHeader, err := parseAddressList(toList(msg.Cc))
	if err != nil {
		toJsError(m.rt, err)
	}
	if len(ccHeader) > 0 {
		_ = w.PrintfLine("Cc: %s", strings.Join(ccHeader, ","))
	}
	to = append(to, cc...)

	bcc, bccHeader, err := parseAddressList(toList(msg.Bcc))
	if err != nil {
		toJsError(m.rt, err)
	}
	if len(bccHeader) > 0 {
		_ = w.PrintfLine("Bcc: %s", strings.Join(bccHeader, ","))
	}
	to = append(to, bcc...)

	_, replyTo, err := parseAddressList(toList(msg.ReplyTo))
	if err != nil {
		toJsError(m.rt, err)
	}
	if len(replyTo) > 0 {
		_ = w.PrintfLine("Reply-To: %s", strings.Join(replyTo, ","))
	}
	to = append(to, bcc...)

	if len(msg.InReplyTo) > 0 {
		_ = w.PrintfLine("In-Reply-To: %s", msg.InReplyTo)
	}

	_ = w.PrintfLine("Subject: %v", msg.Subject)

	u, err := url.Parse(addr)
	if err != nil {
		toJsError(m.rt, err)
	}

	sender := msg.Sender
	if len(sender) == 0 {
		if len(from) == 0 {
			toJsError(m.rt, fmt.Errorf(" A sender or from address must be specified."))
		}
		if len(from) > 1 {
			toJsError(m.rt, fmt.Errorf("sender required if using multiple from addresses"))
		}
		sender = from[0]
	}

	if len(msg.Attachments) == 0 {
		if len(msg.ContentType) == 0 {
			msg.ContentType = "text/plain; charset=UTF-8"
		}
		_ = w.PrintfLine("Content-Type: %v", msg.ContentType)

		if len(msg.Encoding) > 0 {
			_ = w.PrintfLine("Content-Transfer-Encoding: %s", msg.Encoding)
		}

		body.Write([]byte(fmt.Sprintf("\n%s", msg.Body)))
	} else {
		err := m.writeAttachments(w, msg)
		if err != nil {
			toJsError(m.rt, err)
		}
	}

	err = smtp.SendMail(u.Host, nil, sender, to, body.Bytes())
	if err != nil {
		toJsError(m.rt, err)
	}
}

func (m *mailModule) writeAttachments(w *textproto.Writer, msg *Mail) error {
	boundary := fmt.Sprintf("boundary_%v", uuid.New().String())
	_ = w.PrintfLine("Content-Type: multipart/mixed; boundary=%v", boundary)
	_ = w.PrintfLine("")

	_ = w.PrintfLine("--%v", boundary)
	if len(msg.ContentType) == 0 {
		msg.ContentType = "text/plain; charset=UTF-8"
	}
	_ = w.PrintfLine("Content-Type: %v", msg.ContentType)

	if len(msg.Encoding) > 0 {
		_ = w.PrintfLine("Content-Transfer-Encoding: %s", msg.Encoding)
	}
	w.W.WriteString(fmt.Sprintf("\n%s\n", msg.Body))

	for _, attach := range msg.Attachments {
		content := []byte(attach.Content)
		fileName := attach.Filename
		if len(attach.Path) > 0 {
			f, err := m.host.OpenFile(attach.Path, m.workingDir)
			if err != nil {
				return err
			}
			content = f.Raw
			if len(fileName) == 0 {
				fileName = filepath.Base(attach.Path)
			}
		}
		contentType := attach.ContentType
		if len(contentType) == 0 {
			contentType = http.DetectContentType(content)
		}
		if len(fileName) > 0 {
			contentType += fmt.Sprintf("; name=%s", fileName)
		}

		_ = w.PrintfLine("--%v", boundary)
		_ = w.PrintfLine("Content-Type: %v", contentType)
		_ = w.PrintfLine("Content-Transfer-Encoding: base64")
		_ = w.PrintfLine("Content-Disposition: attachment")
		_ = w.PrintfLine("")
		data := base64.StdEncoding.EncodeToString(content)
		_, err := w.W.WriteString(data)
		if err != nil {
			return err
		}
		w.W.WriteRune('\n')
	}

	_ = w.PrintfLine("--%v--", boundary)
	_ = w.PrintfLine("")

	return nil
}

func parseAddressList(list []interface{}) ([]string, []string, error) {
	var raw []string
	var header []string
	for _, i := range list {
		addr, err := toMailAddress(i)
		if err != nil {
			return nil, nil, err
		}
		raw = append(raw, addr.Address)
		header = append(header, addr.String())
	}
	return raw, header, nil
}

func toMailAddress(i interface{}) (*mail.Address, error) {
	switch v := i.(type) {
	case string:
		return &mail.Address{Address: v}, nil
	case map[string]interface{}:
		address, ok := v["address"]
		if !ok {
			return nil, fmt.Errorf("expected address field in %v", i)
		}
		name := v["name"]
		return &mail.Address{Name: fmt.Sprintf("%s", name), Address: fmt.Sprintf("%s", address)}, nil
	}
	return nil, fmt.Errorf("expected mail address but got: %s", i)
}

func toJsError(rt *goja.Runtime, err error) {
	panic(rt.ToValue(err.Error()))
}

func toList(i interface{}) []interface{} {
	if list, ok := i.([]interface{}); ok {
		return list
	} else if i != nil {
		return []interface{}{i}
	}
	return nil
}
