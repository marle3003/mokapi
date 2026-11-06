package mqtt

import (
	"context"
	"io"
	"mokapi/safe"
	"net"
	"runtime/debug"
	"sync"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var ErrServerClosed = errors.New("mqtt: Server closed")

type Handler interface {
	ServeMessage(rw MessageWriter, m *Message)
}

type HandlerFunc func(rw MessageWriter, m *Message)

func (f HandlerFunc) ServeMessage(rw MessageWriter, m *Message) {
	f(rw, m)
}

type MessageWriter interface {
	Write(msg *Message) error
}

type Server struct {
	Addr    string
	Handler Handler

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
	s.mu.Lock()
	// Close sets inShutdown before taking s.mu, so either we see it here or Close sees our listener.
	if s.inShutdown.IsSet() {
		s.mu.Unlock()
		return ErrServerClosed
	}
	s.listener, err = net.Listen("tcp", s.Addr)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	return s.Serve(s.listener)
}

func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			if s.inShutdown.IsSet() {
				return ErrServerClosed
			}
			log.Errorf("mqtt: accept error: %v", err)
			continue
		}

		ctx := s.trackConn(conn)
		go s.serve(conn, ctx)
	}
}

func (s *Server) serve(conn net.Conn, ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		r := recover()
		if r != nil {
			log.Debugf("mqtt panic: %v", string(debug.Stack()))
			log.Errorf("mqtt panic: %v", r)
		}
		cancel()
		s.closeConn(conn)
	}()

	client := ClientFromContext(ctx)
	client.Addr = conn.RemoteAddr().String()
	client.conn = conn
	for {
		r := &Message{Context: ctx}
		err := r.Read(conn, client)
		if err != nil {
			switch {
			case err == io.EOF || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET):
				return
			default:
				log.Errorf("mqtt: %v", err)
				return
			}
		}

		res := &messageWriter{
			conn: conn,
			ctx:  client,
		}

		s.Handler.ServeMessage(res, r)
	}
}

func (s *Server) closeConn(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, ok := s.activeConn[conn]
	if !ok {
		return
	}
	ctx.Done()
	conn.Close()
	delete(s.activeConn, conn)
}

func (s *Server) Close() {
	if s.inShutdown.IsSet() {
		return
	}
	s.inShutdown.SetTrue()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		s.listener.Close()
	}
	for conn, ctx := range s.activeConn {
		ctx.Done()
		conn.Close()
		delete(s.activeConn, conn)
	}
}

func (s *Server) trackConn(conn net.Conn) context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeConn == nil {
		s.activeConn = make(map[net.Conn]context.Context)
	}

	ctx := NewClientContext(context.Background(), conn)

	s.activeConn[conn] = ctx
	return ctx
}

type messageWriter struct {
	conn net.Conn
	ctx  *ClientContext
}

func (mw messageWriter) Write(msg *Message) error {
	return msg.Write(mw.conn, mw.ctx)
}
