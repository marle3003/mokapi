package ldap

import (
	"fmt"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
	"net"
)

type Client struct {
	Addr string

	conn      net.Conn
	messageId int64
}

func NewClient(addr string) *Client {
	return &Client{Addr: addr}
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
	d := net.Dialer{}
	var err error
	c.conn, err = d.Dial("tcp", c.Addr)
	return err
}

//func (c *Client) Send(r *ldap.Request) (*Response, error) {
//	var err error
//	if c.conn == nil {
//		d := net.Dialer{}
//		c.conn, err = d.Dial("tcp", c.Addr)
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	p := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Request")
//	p.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, r.MessageId, "Message ID"))
//	p.AppendChild(r.Message.WritePacket())
//
//	if _, err := c.conn.Write(p.Bytes()); err != nil {
//		return nil, err
//	}
//
//	res, err := ber.ReadPacket(c.conn)
//	if err != nil {
//		return nil, err
//	}
//
//	var msg ldap.Message
//	switch res.Tag {
//	case ldap.ApplicationBindRequest:
//		msg, err = bind.ReadFrom(body)
//	case ldap.ApplicationUnbindRequest:
//		log.Debugf("received unbind request with messageId %v", messageId)
//		// just close connection
//		return
//	case ApplicationAbandonRequest:
//		log.Debugf("received abandon request with messageId %v", messageId)
//		// todo stop any searches on this messageid
//		// The abandon operation does not have a response
//		continue
//	case ApplicationSearchRequest:
//		msg, err = search.ReadFrom(body)
//	}
//
//	if len(res.Children) < 2 {
//		return nil, fmt.Errorf("invalid packat length %v expected at least 2", len(res.Children))
//	}
//	o := res.Children[0].Value
//	messageId, ok := res.Children[0].Value.(int64)
//	if !ok {
//		return nil, fmt.Errorf("malformed messageId %v", reflect.TypeOf(o))
//	}
//	body := res.Children[1]
//	if body.ClassType != ber.ClassApplication {
//		return nil, fmt.Errorf("classType of packet is not ClassApplication was %v", body.ClassType)
//	}
//
//	return &Response{MessageId: messageId, Message: body}, nil
//}
