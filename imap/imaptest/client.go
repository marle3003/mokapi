package imaptest

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mokapi/sasl"
	"net"
	"net/textproto"
	"strings"
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
	var err error
	backoff := 50 * time.Millisecond
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
			return "", err
		}
	}

	err = c.conn.SetDeadline(time.Now().Add(time.Second * 5))
	if err != nil {
		return "", fmt.Errorf("unable to set deadline: %v", err)
	}
	c.tpc = textproto.NewConn(c.conn)
	r, err := c.tpc.ReadLine()
	return r, err
}

func (c *Client) Capability() ([]string, error) {
	r, err := c.Send("CAPABILITY")
	if err != nil {
		return nil, err
	}
	args := strings.SplitN(r, " ", 2)
	caps := strings.Split(args[1], " ")
	return caps[1:], nil
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

func (c *Client) PlainAuth(identity, username, password string) error {
	_, err := c.Send("AUTHENTICATE PLAIN")
	saslClient := sasl.NewPlainClient(identity, username, password)
	cred, err := saslClient.Next(nil)
	if err != nil {
		return err
	}
	_, err = c.SendRaw(base64.StdEncoding.EncodeToString(cred))
	return err
}

func (c *Client) send(line string) (string, error) {
	c.tag++
	err := c.tpc.PrintfLine("A%v %v", c.tag, line)
	if err != nil {
		return "", err
	}
	return c.tpc.ReadLine()
}

func (c *Client) ReadLine() (string, error) {
	return c.tpc.ReadLine()
}
