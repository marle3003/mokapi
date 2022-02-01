package ldaptest

import (
	"fmt"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
	"mokapi/server/ldap"
	"net"
	"reflect"
)

type Client struct {
	Addr string

	conn net.Conn
}

func NewClient(addr string) *Client {
	return &Client{Addr: addr}
}

func (c *Client) Send(r *ldap.Request) (*Response, error) {
	var err error
	if c.conn == nil {
		d := net.Dialer{}
		c.conn, err = d.Dial("tcp", c.Addr)
		if err != nil {
			return nil, err
		}
	}

	p := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Request")
	p.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, r.MessageId, "Message ID"))
	p.AppendChild(r.Body)

	if _, err := c.conn.Write(p.Bytes()); err != nil {
		return nil, err
	}

	res, err := ber.ReadPacket(c.conn)
	if err != nil {
		return nil, err
	}

	if len(res.Children) < 2 {
		return nil, fmt.Errorf("invalid packat length %v expected at least 2", len(res.Children))
	}
	o := res.Children[0].Value
	messageId, ok := res.Children[0].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("malformed messageId %v", reflect.TypeOf(o))
	}
	body := res.Children[1]
	if body.ClassType != ber.ClassApplication {
		return nil, fmt.Errorf("classType of packet is not ClassApplication was %v", body.ClassType)
	}

	return &Response{MessageId: messageId, Body: body}, nil
}
