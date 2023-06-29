package imap

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"mokapi/safe"
	"net"
	"net/textproto"
	"sync"
)

var ErrServerClosed = errors.New("imap: Server closed")

type Server struct {
	Addr string

	mu         sync.Mutex
	activeConn map[net.Conn]context.Context
	listener   net.Listener
	inShutdown safe.AtomicBool
}

func (s *Server) ListenAndServe() error {
	if s.inShutdown.IsSet() {
		return ErrServerClosed
	}

	var err error
	s.listener, err = net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Serve(s.listener)
}

func (s *Server) Serve(l net.Listener) error {
	for {
		c, err := l.Accept()
		if err != nil {
			if s.inShutdown.IsSet() {
				return ErrServerClosed
			}
			return fmt.Errorf("imap: %v", err)
		}

		s.trackConn(c)

		ic := conn{
			tpc: textproto.NewConn(c),
		}
		go ic.serve()
	}
}

func (s *Server) Close() (err error) {
	s.inShutdown.SetTrue()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		err = s.listener.Close()
	}

	for c, ctx := range s.activeConn {
		ctx.Done()
		c.Close()
		delete(s.activeConn, c)
	}
	return
}

func (s *Server) trackConn(conn net.Conn) context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeConn == nil {
		s.activeConn = make(map[net.Conn]context.Context)
	}
	ctx := NewClientContext(context.Background(), conn.RemoteAddr().String())
	s.activeConn[conn] = ctx
	return ctx
}
