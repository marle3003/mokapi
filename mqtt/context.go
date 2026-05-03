package mqtt

import (
	"context"
	"mokapi/buffer"
	"net"
)

const clientKey = "client"

type ClientContext struct {
	Addr            string
	ClientId        string
	ProtocolVersion byte
	ServerAddress   string

	conn net.Conn
}

func ClientFromContext(ctx context.Context) *ClientContext {
	if ctx == nil {
		return nil
	}

	return ctx.Value(clientKey).(*ClientContext)
}

func NewClientContext(ctx context.Context, conn net.Conn) context.Context {
	return context.WithValue(ctx, clientKey, &ClientContext{Addr: conn.RemoteAddr().String(), ServerAddress: conn.LocalAddr().String(), conn: conn})
}

func (c *ClientContext) Send(r *Message) error {
	b := buffer.NewPageBuffer()

	e := NewEncoder(b, c.ProtocolVersion)
	r.Payload.Write(e, r.Header)

	r.Header.Size = b.Size()

	err := r.Header.Write(c.conn)
	if err != nil {
		return err
	}

	_, err = b.WriteTo(c.conn)
	return err
}
