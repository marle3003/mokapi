package mqtt

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"runtime/debug"
	"syscall"
)

type conn struct {
	server *Server
	conn   net.Conn
	ctx    context.Context
}

func (c *conn) serve() {
	ctx, cancel := context.WithCancel(c.ctx)
	defer func() {
		r := recover()
		if r != nil {
			log.Debugf("mqtt panic: %v", string(debug.Stack()))
			log.Errorf("mqtt panic: %v", r)
		}
		cancel()
		c.server.closeConn(c.conn)
	}()

	client := ClientFromContext(ctx)
	client.Addr = c.conn.RemoteAddr().String()
	client.conn = c
	for {
		r := &Request{Context: ctx}
		err := r.Read(c.conn)
		if err != nil {
			switch {
			case err == io.EOF || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET):
				return
			default:
				log.Errorf("mqtt: %v", err)
				return
			}
		}

		res := &response{
			h:   r.Header,
			ctx: client,
		}

		c.server.Handler.ServeMessage(res, r)
	}
}

func (c *conn) writePacket(p *packet) error {
	err := p.header.Write(c.conn)
	if err != nil {
		return err
	}

	_, err = p.payload.WriteTo(c.conn)
	return err
}
