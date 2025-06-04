package mqtt

import (
	"context"
	"mokapi/buffer"
	"net"
	"sync"
)

const clientKey = "client"

type ClientContext struct {
	Addr     string
	ClientId string

	conn      net.Conn
	inflight  []*packet
	m         sync.Mutex
	messageId int16
}

type packet struct {
	header  *Header
	payload buffer.Buffer
	retries int
}

func ClientFromContext(ctx context.Context) *ClientContext {
	if ctx == nil {
		return nil
	}

	return ctx.Value(clientKey).(*ClientContext)
}

func NewClientContext(ctx context.Context, conn net.Conn) context.Context {
	return context.WithValue(ctx, clientKey, &ClientContext{Addr: conn.RemoteAddr().String(), conn: conn})
}

func (c *ClientContext) Send(r *Request) {
	b := buffer.NewPageBuffer()

	e := NewEncoder(b)
	r.Message.Write(e)

	r.Header.Size = b.Size()

	c.sendOrQueue(&packet{
		header:  r.Header,
		payload: b,
	})
}

func (c *ClientContext) sendOrQueue(p *packet) {
	err := c.writePacket(p)
	if err != nil && p.header.QoS > 0 {
		c.m.Lock()
		defer c.m.Unlock()

		c.inflight = append(c.inflight, p)
	} else {
		p.payload.Unref()
	}
}

func (c *ClientContext) retry() {
	c.m.Lock()
	inflight := c.inflight
	c.inflight = nil
	c.m.Unlock()

	for _, p := range inflight {
		c.sendOrQueue(p)
	}
}

func (c *ClientContext) NextMessageId() int16 {
	c.m.Lock()
	defer c.m.Unlock()

	c.messageId++
	return c.messageId
}

func (c *ClientContext) writePacket(p *packet) error {
	err := p.header.Write(c.conn)
	if err != nil {
		return err
	}

	_, err = p.payload.WriteTo(c.conn)
	return err
}
