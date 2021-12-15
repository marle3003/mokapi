package protocol

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
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

type Context interface {
	WithValue(key string, val interface{})
	Value(key string) interface{}
	Close()
}

type context struct {
	values map[string]interface{}
}

func (c *context) WithValue(key string, val interface{}) {
	c.values[key] = val
}

func (c *context) Value(key string) interface{} {
	v, ok := c.values[key]
	if ok {
		return v
	}
	return nil
}

func (c *context) Close() {}

type response struct {
	conn   net.Conn
	header *Header
}

type Server struct {
	Addr        string
	Handler     Handler
	ConnContext func(ctx Context, conn net.Conn) Context

	mu         sync.Mutex
	closeChan  chan bool
	activeConn map[net.Conn]Context
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
		ctx.Close()
		conn.Close()
	}
}

func (s *Server) serve(conn net.Conn, ctx Context) {
	defer func() {
		r := recover()
		if r != nil {
			log.Errorf("kafka panic: %v", r)
		}
		s.closeConn(conn)
	}()

	for {
		r := &Request{Context: s.activeConn[conn]}
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

		ctx.WithValue("clientId", r.Header.ClientId)

		go func() {
			defer func() {
				r := recover()
				if r != nil {
					log.Errorf("kafka panic: %v", r)
					s.closeConn(conn)
				}
			}()
			s.Handler.ServeMessage(&response{conn: conn, header: r.Header}, r)
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
	ctx.Close()
	conn.Close()
	delete(s.activeConn, conn)
}

func (s *Server) trackConn(conn net.Conn) Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeConn == nil {
		s.activeConn = make(map[net.Conn]Context)
	}
	var ctx Context = &context{values: make(map[string]interface{})}
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
	if msg == nil {
		return nil
	}

	res := Response{Header: r.header, Message: msg}
	return res.Write(r.conn)
}
