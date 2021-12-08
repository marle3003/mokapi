package protocol

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
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
	conn   net.Conn
	header *Header
}

type Server struct {
	Addr    string
	Handler Handler

	closeChan  chan bool
	activeConn map[net.Conn]struct{}
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

		s.trackConn(conn)
		go s.serve(conn)
	}
}

func (s *Server) Close() {
	s.getCloseChan() <- true
	for conn := range s.activeConn {
		conn.Close()
	}
}

func (s *Server) serve(conn net.Conn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Errorf("kafka: Close error: %v", err)
		}
	}()

	for {
		r := &Request{}
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
		go s.Handler.ServeMessage(&response{conn: conn}, r)
	}
}

func (s *Server) trackConn(conn net.Conn) {
	if s.activeConn == nil {
		s.activeConn = make(map[net.Conn]struct{})
	}
	s.activeConn[conn] = struct{}{}
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
