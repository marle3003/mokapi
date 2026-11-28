package mail

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/dop251/goja"
	"github.com/google/uuid"
	"mokapi/engine/common"
	"mokapi/smtp"
	"net"
	"net/http"
	"net/mail"
	netsmtp "net/smtp"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type Module struct {
	rt         *goja.Runtime
	host       common.Host
	workingDir string
}

type Mail struct {
	MessageId   string        `json:"messageId"`
	Sender      interface{}   `json:"sender"`
	From        interface{}   `json:"from"`
	To          interface{}   `json:"to"`
	Cc          interface{}   `json:"cc"`
	Bcc         interface{}   `json:"bcc"`
	ReplyTo     interface{}   `json:"replyTo"`
	InReplyTo   string        `json:"inReplyTo"`
	Time        time.Time     `json:"time"`
	Subject     string        `json:"subject"`
	Body        string        `json:"body"`
	ContentType string        `json:"contentType"`
	Encoding    string        `json:"encoding"`
	Attachments []*Attachment `json:"attachments"`
}

type Attachment struct {
	Name        string `json:"name"`
	Data        string `json:"data"`
	Path        string `json:"path"`
	Disposition string `json:"disposition"`
	ContentType string `json:"contentType"`
}

func Require(vm *goja.Runtime, module *goja.Object) {
	o := vm.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)
	f := &Module{
		rt:   vm,
		host: host,
	}
	obj := module.Get("exports").(*goja.Object)
	_ = obj.Set("send", f.Send)
}

func (m *Module) Send(addr string, msg *Mail, auth *Auth) {
	var body bytes.Buffer
	w := textproto.NewWriter(bufio.NewWriter(&body))

	if len(msg.MessageId) > 0 {
		_ = w.PrintfLine("Message-ID: %s", msg.MessageId)
	}

	sender, senderHeader, err := parseSender(msg.Sender)
	if err != nil {
		toJsError(m.rt, err)
	}
	if len(sender) > 0 {
		_ = w.PrintfLine("Sender: %s", senderHeader)
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
	host := addr
	if err == nil {
		host = u.Host
	}

	sender, err = smtp.EnvelopeSender(sender, from)
	if err != nil {
		toJsError(m.rt, err)
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

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	var conn net.Conn
	if u != nil && u.Scheme == "smtps" {
		conn, err = tls.Dial("tcp", host, tlsConfig)
		if err != nil {
			toJsError(m.rt, err)
			return
		}
	} else {
		conn, err = net.Dial("tcp", host)
	}

	c, err := netsmtp.NewClient(conn, host)
	if err != nil {
		toJsError(m.rt, err)
		return
	}

	if err = c.Hello("localhost"); err != nil {
		toJsError(m.rt, err)
		return
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err = c.StartTLS(tlsConfig); err != nil {
			toJsError(m.rt, err)
			return
		}
	}

	if auth != nil {
		if err = c.Auth(auth.getAuth()); err != nil {
			toJsError(m.rt, err)
			return
		}
	}

	if err = c.Mail(sender); err != nil {
		toJsError(m.rt, err)
		return
	}

	for _, t := range to {
		if err = c.Rcpt(t); err != nil {
			toJsError(m.rt, err)
			return
		}
	}

	wc, err := c.Data()
	if err != nil {
		toJsError(m.rt, err)
		return
	}
	_, err = wc.Write(body.Bytes())
	if err != nil {
		toJsError(m.rt, err)
		return
	}
	err = wc.Close()
	if err != nil {
		toJsError(m.rt, err)
		return
	}
	if err = c.Quit(); err != nil {
		toJsError(m.rt, err)
		return
	}
}

// ToMessage converts the mail of a script into a message to deliver. A missing messageId and
// time are generated. Several from addresses need a sender, as for send(). Attachment
// paths are read with host.
func (m *Mail) ToMessage(host common.Host) (*smtp.Message, error) {
	from, err := toAddresses(m.From)
	if err != nil {
		return nil, err
	}
	msg := &smtp.Message{
		From:                    from,
		MessageId:               m.MessageId,
		InReplyTo:               m.InReplyTo,
		Date:                    m.Time,
		Subject:                 m.Subject,
		ContentType:             m.ContentType,
		ContentTransferEncoding: m.Encoding,
		Body:                    m.Body,
	}
	for _, target := range []struct {
		list *[]smtp.Address
		from interface{}
	}{{&msg.To, m.To}, {&msg.Cc, m.Cc}, {&msg.Bcc, m.Bcc}, {&msg.ReplyTo, m.ReplyTo}} {
		if *target.list, err = toAddresses(target.from); err != nil {
			return nil, err
		}
	}

	if m.Sender != nil {
		sender, _, err := parseSender(m.Sender)
		if err != nil {
			return nil, err
		}
		msg.Sender = &smtp.Address{Address: sender}
		if a, ok := m.Sender.(map[string]interface{}); ok && a["name"] != nil {
			msg.Sender.Name = fmt.Sprintf("%s", a["name"])
		}
	}
	if _, err := msg.EnvelopeSender(); err != nil {
		return nil, err
	}

	if msg.MessageId == "" {
		msg.MessageId = smtp.NewMessageId()
	}
	if msg.Date.IsZero() {
		msg.Date = time.Now()
	}
	for _, a := range m.Attachments {
		att, err := a.toAttachment(host)
		if err != nil {
			return nil, err
		}
		msg.Attachments = append(msg.Attachments, att)
	}
	return msg, nil
}

func (a *Attachment) toAttachment(host common.Host) (smtp.Attachment, error) {
	content := []byte(a.Data)
	name := a.Name
	if len(a.Path) > 0 {
		f, err := host.OpenFile(a.Path, "")
		if err != nil {
			return smtp.Attachment{}, err
		}
		content = f.Raw
		if len(name) == 0 {
			name = filepath.Base(a.Path)
		}
	}
	contentType := a.ContentType
	if len(contentType) == 0 {
		contentType = http.DetectContentType(content)
	}
	return smtp.Attachment{
		Name:                    name,
		ContentType:             contentType,
		ContentTransferEncoding: "base64",
		Disposition:             a.Disposition,
		Data:                    content,
	}, nil
}

func toAddresses(i interface{}) ([]smtp.Address, error) {
	var result []smtp.Address
	for _, item := range toList(i) {
		a, err := toMailAddress(item)
		if err != nil {
			return nil, err
		}
		result = append(result, smtp.Address{Name: a.Name, Address: a.Address})
	}
	return result, nil
}

// parseSender accepts the sender as plain address or as address object. A plain address is
// kept as written for the Sender header.
func parseSender(sender interface{}) (address string, header string, err error) {
	switch v := sender.(type) {
	case nil:
		return "", "", nil
	case string:
		return v, v, nil
	default:
		a, err := toMailAddress(v)
		if err != nil {
			return "", "", err
		}
		return a.Address, a.String(), nil
	}
}

func (m *Module) writeAttachments(w *textproto.Writer, msg *Mail) error {
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
	_, _ = w.W.WriteString(fmt.Sprintf("\n%s\n", msg.Body))

	for _, attach := range msg.Attachments {
		content := []byte(attach.Data)
		name := attach.Name
		if len(attach.Path) > 0 {
			f, err := m.host.OpenFile(attach.Path, m.workingDir)
			if err != nil {
				return err
			}
			content = f.Raw
			if len(name) == 0 {
				name = filepath.Base(attach.Path)
			}
		}
		contentType := attach.ContentType
		if len(contentType) == 0 {
			contentType = http.DetectContentType(content)
		}
		if len(name) > 0 {
			contentType += fmt.Sprintf("; name=%s", name)
		}

		_ = w.PrintfLine("--%v", boundary)
		_ = w.PrintfLine("Content-Type: %v", contentType)
		_ = w.PrintfLine("Content-Transfer-Encoding: base64")
		disposition := attach.Disposition
		if len(disposition) == 0 {
			disposition = "attachment"
		}
		_ = w.PrintfLine("Content-Disposition: %v", disposition)
		_ = w.PrintfLine("")
		data := base64.StdEncoding.EncodeToString(content)
		_, err := w.W.WriteString(data)
		if err != nil {
			return err
		}
		_, _ = w.W.WriteRune('\n')
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
		addr := &mail.Address{Address: fmt.Sprintf("%s", address)}
		if name != nil {
			addr.Name = fmt.Sprintf("%s", name)
		}
		return addr, nil
	case smtp.Address:
		return &mail.Address{Name: v.Name, Address: v.Address}, nil
	}
	return nil, fmt.Errorf("expected mail address but got: %s", i)
}

func toJsError(rt *goja.Runtime, err error) {
	panic(rt.ToValue(err.Error()))
}

func toList(i interface{}) []interface{} {
	switch v := i.(type) {
	case []interface{}:
		return v
	case []smtp.Address:
		var r []interface{}
		for _, a := range v {
			r = append(r, a)
		}
		return r
	default:
		if i != nil {
			return []interface{}{i}
		}
		return nil
	}
}
