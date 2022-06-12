package ldap

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
	"io"
	"net"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

var ErrServerClosed = errors.New("ldap: Server closed")

type atomicBool int32

func (a *atomicBool) isSet() bool { return atomic.LoadInt32((*int32)(a)) != 0 }
func (a *atomicBool) setFalse()   { atomic.StoreInt32((*int32)(a), 0) }
func (a *atomicBool) setTrue()    { atomic.StoreInt32((*int32)(a), 1) }

type Handler interface {
	Serve(ResponseWriter, *Request)
}

type Request struct {
	Context   context.Context
	MessageId int64
	Body      *ber.Packet
}

type ResponseWriter interface {
	Write(packet *ber.Packet) error
}

type response struct {
	conn net.Conn
}

type Server struct {
	Addr    string
	Handler Handler

	mu         sync.Mutex
	closeChan  chan bool
	activeConn map[net.Conn]context.Context
	listener   net.Listener
	inShutdown atomicBool
}

func (s *Server) ListenAndServe() error {
	if s.inShutdown.isSet() {
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
	s.inShutdown.setTrue()

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
			log.Debugf("ldap panic: %v", string(debug.Stack()))
			log.Errorf("ldap panic: %v", r)
		}
		cancel()
		s.closeConn(conn)
	}()

	for {
		packet, err := ber.ReadPacket(conn)
		if err == io.EOF { // Client closed connection
			return
		} else if err != nil {
			log.Debugf("handleConnection ber.ReadPacket ERROR: %v", err.Error())
			return
		}

		if len(packet.Children) < 2 {
			log.Infof("invalid packat length %v expected at least 2", len(packet.Children))
			return
		}
		o := packet.Children[0].Value
		messageId, ok := packet.Children[0].Value.(int64)
		if !ok {
			log.Infof("malformed messageId %v", reflect.TypeOf(o))
			return
		}
		body := packet.Children[1]
		if body.ClassType != ber.ClassApplication {
			log.Infof("classType of packet is not ClassApplication was %v", body.ClassType)
			return
		}

		s.Handler.Serve(&response{
			conn: conn,
		}, &Request{
			Context:   ctx,
			MessageId: messageId,
			Body:      body,
		})
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
	s.activeConn[conn] = ctx
	return ctx
}

func (s *Server) getCloseChan() chan bool {
	if s.closeChan == nil {
		s.closeChan = make(chan bool, 1)
	}
	return s.closeChan
}

func (r *response) Write(body *ber.Packet) error {
	_, err := r.conn.Write(body.Bytes())
	return err
}
