package smtp

import (
	"crypto/tls"
	"crypto/x509"
	"mokapi/server/cert"
	"net"
	"net/textproto"
	"strings"
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

func (c *Client) Write(line string, expectCode int) ([]string, error) {
	err := c.ensureConn()
	if err != nil {
		return nil, err
	}

	err = c.conn.tpc.PrintfLine(line)
	if err != nil {
		return nil, err
	}

	_, line, err = c.conn.tpc.ReadResponse(expectCode)
	lines := strings.Split(line, "\n")
	return lines, err
}

func (c *Client) Dial() (string, error) {
	var err error
	backoff := 50 * time.Millisecond
	if c.conn == nil {
		for i := 0; i < 10; i++ {
			d := &net.Dialer{}
			var co net.Conn
			co, err = d.Dial("tcp", c.Addr)
			if err != nil {
				time.Sleep(backoff)
				continue
			}
			c.conn = &conn{
				conn: co,
				tpc:  textproto.NewConn(co),
			}
			break
		}
	}
	if err != nil {
		return "", err
	}

	_, msg, err := c.conn.tpc.ReadResponse(220)
	return msg, err
}

func (c *Client) DialTls(cfg *tls.Config) (string, error) {
	var err error
	backoff := 50 * time.Millisecond
	if c.conn == nil {
		for i := 0; i < 10; i++ {
			d := &net.Dialer{}
			var tlsConn *tls.Conn
			tlsConn, err = tls.DialWithDialer(d, "tcp", c.Addr, cfg)
			if err != nil {
				time.Sleep(backoff)
				continue
			}
			c.conn = &conn{
				conn: tlsConn,
				tpc:  textproto.NewConn(tlsConn),
			}
			break
		}
	}
	if err != nil {
		return "", err
	}

	_, msg, err := c.conn.tpc.ReadResponse(220)
	return msg, err
}
