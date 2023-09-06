package ldap

import (
	"fmt"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
	"net"
	"time"
)

type Client struct {
	Addr    string
	Timeout time.Duration

	conn      net.Conn
	messageId int64
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
}

func (c *Client) Bind(username, password string) (*BindResponse, error) {
	r, err := c.newRequest(&BindRequest{
		Version:  3,
		Name:     username,
		Password: password,
		Auth:     Simple,
	})
	if err != nil {
		return nil, err
	}

	_, err = c.conn.Write(r.Bytes())
	if err != nil {
		return nil, err
	}

	p, err := ber.ReadPacket(c.conn)

	return readBindResponse(p.Children[1])
}

func (c *Client) Unbind() error {
	r, err := c.newRequest(&UnbindRequest{})
	if err != nil {
		return err
	}

	_, err = c.conn.Write(r.Bytes())
	return err
}

func (c *Client) Search(request *SearchRequest) (*SearchResponse, error) {
	r, err := c.newRequest(request)
	if err != nil {
		return nil, err
	}
	n, err := c.conn.Write(r.Bytes())
	_ = n
	if err != nil {
		return nil, err
	}

	var packets []*ber.Packet
	for {
		p, err := ber.ReadPacket(c.conn)
		if err != nil {
			return nil, err
		}
		body := p.Children[1]
		packets = append(packets, body)
		if body.Tag == searchDone {
			break
		}
	}

	return readSearchResponse(packets)
}

func (c *Client) AbandonSearch(messageId int64) error {
	r, err := c.newRequest(&AbandonRequest{MessageId: messageId})
	if err != nil {
		return err
	}

	_, err = c.conn.Write(r.Bytes())
	return err
}

func (c *Client) newRequest(msg Message) (*ber.Packet, error) {
	c.messageId++
	p := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Request")
	p.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, c.messageId, "Message ID"))
	switch b := msg.(type) {
	case *BindRequest:
		p.AppendChild(b.toPacket())
	case *UnbindRequest:
		p.AppendChild(b.toPacket())
	case *AbandonRequest:
		p.AppendChild(b.toPacket())
	case *SearchRequest:
		body, err := b.toPacket()
		if err != nil {
			return nil, err
		}
		p.AppendChild(body)
	default:
		return nil, fmt.Errorf("unsupported request type %t", msg)
	}
	return p, nil
}

func (c *Client) Dial() error {
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
