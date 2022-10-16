package smtptest

import (
	"mokapi/smtp"
	"net"
	"time"
)

type Client struct {
	Addr     string
	Timeout  time.Duration
	ClientId string

	conn net.Conn
}

func NewClient(addr string) *Client {
	return &Client{Addr: addr, ClientId: "smtptest", Timeout: time.Second * 10}
}

func (c *Client) Close() {
	if c.conn == nil {
		return
	}
	err := c.conn.Close()
	if err != nil {
		panic(err)
	}
}

func (c *Client) Connect() (*Response, error) {
	backoff := 50 * time.Millisecond
	var err error
	if c.conn == nil {
		for i := 0; i < 10; i++ {
			d := net.Dialer{Timeout: c.Timeout}
			c.conn, err = d.Dial("tcp", c.Addr)
			if err != nil {
				time.Sleep(backoff)
				continue
			}
		}
		if err != nil {
			return nil, err
		}
	}

	res := NewResponse()
	err = res.Read(c.conn)
	return res, err
}

func (c *Client) Send(r *smtp.Request) (*Response, error) {
	err := r.Write(c.conn)
	if err != nil {
		return nil, err
	}
	res := NewResponse()
	err = res.Read(c.conn)
	return res, err
}
