package imaptest

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/textproto"
	"time"
)

type Client struct {
	Addr    string
	Timeout time.Duration

	tpc  *textproto.Conn
	conn net.Conn
	tag  uint64
}

func NewClient(addr string) *Client {
	return &Client{Addr: addr, Timeout: time.Second * 10}
}

func (c *Client) Dial() (string, error) {
	d := net.Dialer{Timeout: c.Timeout}
	var err error
	c.conn, err = d.Dial("tcp", c.Addr)
	if err != nil {
		return "", err
	}
	err = c.conn.SetDeadline(time.Now().Add(time.Second * 5))
	if err != nil {
		return "", fmt.Errorf("unable to set deadline: %v", err)
	}
	c.tpc = textproto.NewConn(c.conn)
	r, err := c.tpc.ReadLine()
	return r, err
}

func (c *Client) Send(line string) (string, error) {
	return c.send(line)
}

func (c *Client) SendRaw(line string) (string, error) {
	err := c.tpc.PrintfLine(line)
	if err != nil {
		return "", err
	}
	return c.tpc.ReadLine()
}

func (c *Client) StartTLS() (string, error) {
	r, err := c.send("STARTTLS")

	tlsConn := tls.Client(c.conn, &tls.Config{InsecureSkipVerify: true})
	c.conn = tlsConn
	c.tpc = textproto.NewConn(tlsConn)
	return r, err
}

func (c *Client) Login(username, password string) (string, error) {
	c.tag++
	c.tpc.PrintfLine("A%v AUTHENTICATE PLAIN", c.tag)
	r, err := c.tpc.ReadLine()

	return r, err
}

func (c *Client) send(line string) (string, error) {
	c.tag++
	err := c.tpc.PrintfLine("A%v %v", c.tag, line)
	if err != nil {
		return "", err
	}
	return c.tpc.ReadLine()
}
