package mqtttest

import (
	"mokapi/mqtt"
	"net"
	"time"
)

type Client struct {
	Addr    string
	Timeout time.Duration

	conn net.Conn
}

func NewClient(addr string) *Client {
	return &Client{Addr: addr, Timeout: time.Second * 10}
}

func (c *Client) Close() {
	if c.conn == nil {
		return
	}
	err := c.conn.Close()
	if err != nil {
		panic(err)
	}
	c.conn = nil
}

func (c *Client) Send(r *mqtt.Request) (*mqtt.Response, error) {
	if err := c.ensureConnection(); err != nil {
		return nil, err
	}

	err := r.Write(c.conn)
	if err != nil {
		return nil, err
	}

	c.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	res, err := mqtt.ReadResponse(c.conn)
	return res, err
}

func (c *Client) SendNoResponse(r *mqtt.Request) error {
	if err := c.ensureConnection(); err != nil {
		return err
	}

	return r.Write(c.conn)
}

func (c *Client) Recv() (*mqtt.Response, error) {
	if err := c.ensureConnection(); err != nil {
		return nil, err
	}

	res, err := mqtt.ReadResponse(c.conn)
	return res, err
}

func (c *Client) ensureConnection() error {
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
			return err
		}
	}
	return err
}
