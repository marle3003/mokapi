package mqtt

import (
	"context"
	"mokapi/buffer"
	"net"
)

const clientKey = "client"

type ClientContext struct {
	Addr     string
	ClientId string

	conn net.Conn
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

func (c *ClientContext) Send(r *Message) error {
	b := buffer.NewPageBuffer()

	e := NewEncoder(b)
	r.Payload.Write(e)

	r.Header.Size = b.Size()

	err := r.Header.Write(c.conn)
	if err != nil {
		return err
	}

	_, err = b.WriteTo(c.conn)
	return err
}
