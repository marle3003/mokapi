package protocol

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"runtime/debug"
	"sync"
)

var ErrServerClosed = errors.New("kafka: Server closed")

type Handler interface {
	ServeMessage(rw ResponseWriter, req *Request)
}

type ResponseWriter interface {
	WriteHeader(key ApiKey, version int, correlationId int)
	Write(msg Message) error
}

type response struct {
	ctx    context.Context
	conn   net.Conn
	header *Header
}

type Server struct {
	Addr        string
	Handler     Handler
	ConnContext func(ctx context.Context, conn net.Conn) context.Context

	mu         sync.Mutex
	closeChan  chan bool
	activeConn map[net.Conn]context.Context
	listener   net.Listener
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Serve(l)
}

func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-s.getCloseChan():
				return ErrServerClosed
			default:
				log.Errorf("kafka: Accept error: %v", err)
				continue
			}
		}

		ctx := s.trackConn(conn)
		go s.serve(conn, ctx)
	}
}

func (s *Server) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.getCloseChan() <- true
	if s.listener != nil {
		s.listener.Close()
	}
	for conn, ctx := range s.activeConn {
		ctx.Done()
		conn.Close()
	}
}

func (s *Server) serve(conn net.Conn, ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		r := recover()
		if r != nil {
			log.Debugf("kafka panic: %v", string(debug.Stack()))
			log.Errorf("kafka panic: %v", r)
		}
		cancel()
		s.closeConn(conn)
	}()

	for {
		r := &Request{Context: ctx}
		var err error
		r.Header, r.Message, err = ReadMessage(conn)
		if err != nil {
			switch {
			case err == io.EOF:
				return
			default:
				log.Errorf("kafka: %v", err)
				return
			}
		}

		ctx = context.WithValue(ctx, "clientId", r.Header.ClientId)
		go func() {
			defer func() {
				err := recover()
				if err != nil {
					log.Debugf("kafka panic: %v", string(debug.Stack()))
					log.Errorf("kafka panic: %v", err)
				}
			}()
			s.Handler.ServeMessage(&response{conn: conn, header: r.Header, ctx: ctx}, r)
		}()
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

func (s *Server) trackConn(conn net.Conn) context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeConn == nil {
		s.activeConn = make(map[net.Conn]context.Context)
	}
	ctx := context.Background()
	if s.ConnContext != nil {
		ctx = s.ConnContext(ctx, conn)
	}
	s.activeConn[conn] = ctx
	return ctx
}

func (s *Server) getCloseChan() chan bool {
	if s.closeChan == nil {
		s.closeChan = make(chan bool, 1)
	}
	return s.closeChan
}

func (r *response) WriteHeader(key ApiKey, version int, correlationId int) {
	r.header = &Header{ApiKey: key, ApiVersion: int16(version), CorrelationId: int32(correlationId)}
}

func (r *response) Write(msg Message) error {
	if r.ctx.Err() != nil || msg == nil {
		return nil
	}

	res := Response{Header: r.header, Message: msg}
	err := res.Write(r.conn)
	return err
}
