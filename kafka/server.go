package kafka

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"mokapi/safe"
	"net"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

var ErrServerClosed = errors.New("kafka: Server closed")

type Handler interface {
	ServeMessage(rw ResponseWriter, req *Request)
}

type HandlerFunc func(rw ResponseWriter, req *Request)

func (f HandlerFunc) ServeMessage(rw ResponseWriter, req *Request) {
	f(rw, req)
}

type ResponseWriter interface {
	WriteHeader(key ApiKey, version int, correlationId int)
	Write(msg Message) error
}

type response struct {
	ctx    context.Context
	writer io.Writer
	header *Header
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
	s.listener, err = net.Listen("tcp", s.Addr)
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
				log.Errorf("kafka: Accept error: %v", err)
				continue
			}
		}

		ctx := s.trackConn(conn)
		go s.serve(conn, ctx)
	}
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
		r := &Request{Context: ctx, Host: conn.LocalAddr().String()}
		err := r.Read(conn)
		if err != nil {
			switch {
			case err == io.EOF || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET):
				return
			default:
				log.Errorf("kafka: %v", err)
				return
			}
		}
		if r.Header == nil {
			continue
		}

		go func() {
			defer func() {
				err := recover()
				if err != nil {
					log.Debugf("kafka panic: %v", string(debug.Stack()))
					log.Errorf("kafka panic: %v", err)
				}
			}()
			s.handleMessage(&response{writer: conn, header: r.Header, ctx: ctx}, r)
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
	ctx := NewClientContext(context.Background(), conn.RemoteAddr().String())

	s.activeConn[conn] = ctx
	return ctx
}

func (s *Server) getCloseChan() chan bool {
	if s.closeChan == nil {
		s.closeChan = make(chan bool, 1)
	}
	return s.closeChan
}

func (s *Server) handleMessage(rw ResponseWriter, req *Request) {
	client := ClientFromContext(req)
	client.Heartbeat = time.Now()
	client.ClientId = req.Header.ClientId

	s.Handler.ServeMessage(rw, req)
}

func (r *response) WriteHeader(key ApiKey, version int, correlationId int) {
	r.header = &Header{ApiKey: key, ApiVersion: int16(version), CorrelationId: int32(correlationId)}
}

func (r *response) Write(msg Message) error {
	if r.ctx.Err() != nil || msg == nil {
		return nil
	}

	res := Response{Header: r.header, Message: msg}
	err := res.Write(r.writer)
	return err
}
