package mqtt

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/safe"
	"net"
	"sync"
)

var ErrServerClosed = errors.New("mqtt: Server closed")

type Handler interface {
	ServeMessage(rw ResponseWriter, req *Request)
}

type HandlerFunc func(rw ResponseWriter, req *Request)

func (f HandlerFunc) ServeMessage(rw ResponseWriter, req *Request) {
	f(rw, req)
}

type ResponseWriter interface {
	Write(messageType Type, msg Message)
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
		rw, err := l.Accept()
		if err != nil {
			select {
			case <-closeCh:
				return ErrServerClosed
			default:
				log.Errorf("mqtt: accept error: %v", err)
				continue
			}
		}

		c := conn{
			server: s,
			conn:   rw,
			ctx:    s.trackConn(rw),
		}
		go c.serve()
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
	ctx := NewClientContext(context.Background(), conn.RemoteAddr().String())

	s.activeConn[conn] = ctx
	return ctx
}
