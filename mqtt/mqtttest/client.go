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
	ctx  *mqtt.ClientContext
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

func (c *Client) Send(m *mqtt.Message) (*mqtt.Message, error) {
	if err := c.ensureConnection(); err != nil {
		return nil, err
	}

	if connect, ok := m.Payload.(*mqtt.ConnectRequest); ok {
		c.ctx.ProtocolVersion = connect.Version
	}

	err := m.Write(c.conn, c.ctx)
	if err != nil {
		return nil, err
	}

	_ = c.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	res := &mqtt.Message{}
	err = res.Read(c.conn, c.ctx)
	return res, err
}

func (c *Client) SendNoResponse(r *mqtt.Message) error {
	if err := c.ensureConnection(); err != nil {
		return err
	}

	return r.Write(c.conn, c.ctx)
}

func (c *Client) Recv() (*mqtt.Message, error) {
	if err := c.ensureConnection(); err != nil {
		return nil, err
	}

	res := &mqtt.Message{}
	err := res.Read(c.conn, c.ctx)
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
	if c.ctx == nil {
		c.ctx = &mqtt.ClientContext{}
	}
	return err
}
