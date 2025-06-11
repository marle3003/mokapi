package mqtt

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"mokapi/safe"
	"net"
	"runtime/debug"
	"sync"
	"syscall"
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
	closeChan  chan bool
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
	s.listener, err = net.Listen("tcp", s.Addr)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	return s.Serve(s.listener)
}

func (s *Server) Serve(l net.Listener) error {
	closeCh := s.getCloseChan()
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-closeCh:
				return ErrServerClosed
			default:
				log.Errorf("mqtt: accept error: %v", err)
				continue
			}
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
		err := r.Read(conn)
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

	if s.closeChan != nil {
		s.closeChan <- true
		close(s.closeChan)
	}

	if s.listener != nil {
		s.listener.Close()
	}
	for conn, ctx := range s.activeConn {
		ctx.Done()
		conn.Close()
		delete(s.activeConn, conn)
	}
}

func (s *Server) getCloseChan() chan bool {
	if s.closeChan == nil {
		s.closeChan = make(chan bool, 1)
	}
	return s.closeChan
}

func (s *Server) trackConn(conn net.Conn) context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeConn == nil {
		s.activeConn = make(map[net.Conn]context.Context)
	}

	// delete conn struct and implement all in server
	// NewClientContext(context, net.Conn)
	ctx := NewClientContext(context.Background(), conn)

	s.activeConn[conn] = ctx
	return ctx
}

type messageWriter struct {
	conn net.Conn
}

func (mw messageWriter) Write(msg *Message) error {
	return msg.Write(mw.conn)
}
