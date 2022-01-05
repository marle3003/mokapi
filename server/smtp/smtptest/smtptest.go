package smtptest

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/smtp"
	"net/url"
	"time"
)

type MailOptions func(m *mail)

type mail struct {
	from      string
	to        string
	subject   string
	body      []byte
	auth      smtp.Auth
	u         *url.URL
	tlsConfig *tls.Config
}

func SendMail(
	from,
	to,
	addr string,
	opts ...MailOptions,
) error {
	m := &mail{
		from: from,
		to:   to,
	}

	u, err := url.Parse(addr)
	if err != nil {
		return err
	}
	m.u = u

	for _, opt := range opts {
		opt(m)
	}

	conn, err := getConn(m)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, m.u.Hostname())
	if err != nil {
		return err
	}

	if m.auth != nil {
		err := c.Auth(m.auth)
		if err != nil {
			return err
		}
	}

	err = c.Mail(m.from)
	if err != nil {
		return err
	}

	err = c.Rcpt(m.to)
	if err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("From: %v\r\n"+
		"To: %v\r\n"+
		"Subject: %v\r\n\r\n", from, to, m.subject)

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}

	_, err = w.Write(m.body)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

func WithRootCa(cert *x509.Certificate) MailOptions {
	return func(m *mail) {
		if m.tlsConfig == nil {
			m.tlsConfig = &tls.Config{}
		}
		pool := x509.NewCertPool()
		pool.AddCert(cert)
		m.tlsConfig.RootCAs = pool
	}
}

func InsecureSkipVerfiy() MailOptions {
	return func(m *mail) {
		if m.tlsConfig == nil {
			m.tlsConfig = &tls.Config{}
		}
		m.tlsConfig.InsecureSkipVerify = true
	}
}

func WithSubject(title string) MailOptions {
	return func(m *mail) {
		m.subject = title
	}
}

func WithBody(body string) MailOptions {
	return func(m *mail) {
		m.body = []byte(body)
	}
}

func WithPlainAuth(username, password string) MailOptions {
	return func(m *mail) {
		m.auth = smtp.PlainAuth("", username, password, m.u.Hostname())
	}
}

func getConn(m *mail) (net.Conn, error) {
	addr := fmt.Sprintf("%v:%v", m.u.Hostname(), m.u.Port())
	if m.u.Scheme == "smtps" || m.u.Port() == "587" {
		tlsDialer := tls.Dialer{
			NetDialer: &net.Dialer{
				Timeout: 30 * time.Second,
			},
			Config: m.tlsConfig,
		}
		return tlsDialer.Dial("tcp", addr)
	}
	return net.Dial("tcp", addr)
}
