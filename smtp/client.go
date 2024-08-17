package smtp

import (
	"crypto/tls"
	"crypto/x509"
	"mokapi/server/cert"
	"net"
	"net/textproto"
	"time"
)

type Client struct {
	Addr      string
	TlsConfig *tls.Config

	conn *conn
	ext  map[string]string
}

func NewClient(addr string) *Client {
	return &Client{Addr: addr}
}

func (c *Client) Close() error {
	if c.conn != nil {
		err := c.conn.quit()
		if err != nil {
			return err
		}
		c.conn.close()
	}
	return nil
}

func (c *Client) Send(sender Address, to []Address, msg *Message) error {
	err := c.ensureConn()
	if err != nil {
		return err
	}
	c.ext, err = c.conn.ehlo()
	if err != nil {
		return err
	}
	err = c.tryStartTls()
	if err != nil {
		return err
	}
	err = c.conn.mail(sender.Address)
	if err != nil {
		return err
	}
	for _, rcpt := range to {
		err = c.conn.rcpt(rcpt.Address)
		if err != nil {
			return err
		}
	}

	w, err := c.conn.data()
	if err != nil {
		return err
	}
	err = msg.WriteTo(w)
	if err != nil {
		return err
	}
	return w.Close()
}

func (c *Client) tryStartTls() error {
	if _, ok := c.ext["STARTTLS"]; !ok {
		return nil
	}

	serverName, _, err := net.SplitHostPort(c.Addr)
	if err != nil {
		return err
	}
	_ = serverName
	cfg := c.TlsConfig
	if cfg == nil {
		caCertPool := x509.NewCertPool()
		caCertPool.AddCert(cert.DefaultRootCert())
		cfg = &tls.Config{ServerName: serverName, RootCAs: caCertPool}
	}
	return c.conn.startTls(cfg)
}

func (c *Client) ensureConn() error {
	if c.conn == nil {
		backoff := 50 * time.Millisecond
		d := net.Dialer{}
		var err error
		var base net.Conn
		for i := 0; i < 10; i++ {
			base, err = d.Dial("tcp", c.Addr)
			if err != nil {
				time.Sleep(backoff)
				continue
			}
		}
		if err != nil {
			return err
		}
		c.conn = &conn{
			conn: base,
			tpc:  textproto.NewConn(base),
		}

		_, _, err = c.conn.tpc.ReadResponse(220)
		return err

	}
	return nil
}
