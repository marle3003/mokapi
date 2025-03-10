package imap

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mokapi/sasl"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

type Client struct {
	Addr    string
	Timeout time.Duration

	tpc  *textproto.Conn
	conn net.Conn

	tag uint64
	m   sync.Mutex
}

func NewClient(addr string) *Client {
	return &Client{Addr: addr, Timeout: time.Second * 10}
}

func (c *Client) Dial() ([]string, error) {
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
			break
		}
		if err != nil {
			return nil, err
		}
	}

	err = c.conn.SetDeadline(time.Now().Add(time.Second * 5))
	if err != nil {
		return nil, fmt.Errorf("unable to set deadline: %v", err)
	}
	c.tpc = textproto.NewConn(c.conn)
	d := &Decoder{}
	d.msg, err = c.tpc.ReadLine()

	return c.readGreetings(d)
}

func (c *Client) Capability() ([]string, error) {
	tag := c.nextTag()
	err := c.tpc.PrintfLine("%s CAPABILITY", tag)
	if err != nil {
		return nil, err
	}

	d := &Decoder{}
	var caps []string
	for {
		d.msg, err = c.tpc.ReadLine()

		if d.is(tag) {
			return caps, d.EndCmd(tag)
		}

		if err = d.expect("*"); err != nil {
			return nil, err
		}

		if err = d.SP().expect("CAPABILITY"); err != nil {
			return nil, err
		}

		var name string
		for d.IsSP() {
			name, err = d.SP().String()
			if err != nil {
				return nil, err
			}
			caps = append(caps, name)
		}
	}
}

func (c *Client) Send(line string) ([]string, error) {
	tag := c.nextTag()
	err := c.tpc.PrintfLine("%v %v", tag, line)
	if err != nil {
		return nil, err
	}
	var lines []string
	for {
		resp, err := c.tpc.ReadLine()
		if err != nil {
			return lines, err
		}
		lines = append(lines, resp)
		if strings.HasPrefix(resp, tag) {
			break
		}
	}

	return lines, nil
}

func (c *Client) SendRaw(line string) (string, error) {
	err := c.tpc.PrintfLine(line)
	if err != nil {
		return "", err
	}
	return c.tpc.ReadLine()
}

func (c *Client) StartTLS() error {
	tag := c.nextTag()
	err := c.tpc.PrintfLine("%s STARTTLS", tag)
	if err != nil {
		return err
	}

	d := &Decoder{}
	d.msg, err = c.tpc.ReadLine()
	if err != nil {
		return err
	}
	err = d.EndCmd(tag)
	if err != nil {
		return err
	}

	tlsConn := tls.Client(c.conn, &tls.Config{InsecureSkipVerify: true})
	c.conn = tlsConn
	c.tpc = textproto.NewConn(tlsConn)
	return nil
}

func (c *Client) Login(username, password string) (string, error) {
	c.tpc.PrintfLine("A%v AUTHENTICATE PLAIN", c.nextTag())
	r, err := c.tpc.ReadLine()

	return r, err
}

func (c *Client) PlainAuth(identity, username, password string) error {
	_, err := c.send("AUTHENTICATE PLAIN")
	saslClient := sasl.NewPlainClient(identity, username, password)
	cred, err := saslClient.Next(nil)
	if err != nil {
		return err
	}
	res, err := c.SendRaw(base64.StdEncoding.EncodeToString(cred))
	if err != nil {
		return err
	}
	if !strings.Contains(res, "OK") {
		return fmt.Errorf(res)
	}
	return nil
}

func (c *Client) send(line string) (string, error) {
	tag := c.nextTag()
	err := c.tpc.PrintfLine("%v %v", tag, line)
	if err != nil {
		return "", err
	}
	return c.tpc.ReadLine()
}

func (c *Client) ReadLine() (string, error) {
	return c.tpc.ReadLine()
}

func (c *Client) readGreetings(d *Decoder) ([]string, error) {
	var err error
	if err = d.expect("*"); err != nil {
		return nil, err
	}

	if err = d.SP().expect("OK"); err != nil {
		return nil, err
	}

	var caps []string
	if d.SP().is("[") {
		_ = d.expect("[")
		var key string
		key, err = d.String()
		switch strings.ToUpper(key) {
		case "CAPABILITY":
			var name string
			for d.IsSP() {
				name, err = d.SP().String()
				if err != nil {
					return nil, err
				}
				caps = append(caps, name)
			}
		}
	}
	return caps, err
}

func (c *Client) nextTag() string {
	c.m.Lock()
	defer c.m.Unlock()

	c.tag++
	return fmt.Sprintf("A%04d", c.tag)
}
