package mqtt

import (
	"mokapi/buffer"
	"sync"
)

type ClientSession struct {
	conn     *conn
	inflight []*packet
	m        sync.Mutex
}

type packet struct {
	header  *Header
	payload buffer.Buffer
	retries int
}

func (c *ClientSession) sendOrQueue(p *packet) {
	err := c.conn.writePacket(p)
	if err != nil && p.header.Qos > 0 {
		c.m.Lock()
		defer c.m.Unlock()

		c.inflight = append(c.inflight, p)
	} else {
		p.payload.Unref()
	}
}

func (c *ClientSession) retry() {
	c.m.Lock()
	inflight := c.inflight
	c.inflight = nil
	c.m.Unlock()

	for _, p := range inflight {
		c.sendOrQueue(p)
	}
}
