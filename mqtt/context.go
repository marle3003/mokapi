package mqtt

import (
	"context"
	"mokapi/buffer"
	"sync"
)

const clientKey = "client"

type ClientContext struct {
	Addr     string
	ClientId string

	conn      *conn
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

func NewClientContext(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, clientKey, &ClientContext{Addr: addr})
}

func (c *ClientContext) sendOrQueue(p *packet) {
	err := c.conn.writePacket(p)
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

func (c *ClientContext) nextMessageId() int16 {
	c.m.Lock()
	defer c.m.Unlock()

	c.messageId++
	return c.messageId
}
