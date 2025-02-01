package ldap

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
	"io"
	"net"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"
)

var ErrServerClosed = errors.New("ldap: Server closed")

type atomicBool int32

func (a *atomicBool) isSet() bool { return atomic.LoadInt32((*int32)(a)) != 0 }
func (a *atomicBool) setFalse()   { atomic.StoreInt32((*int32)(a), 0) }
func (a *atomicBool) setTrue()    { atomic.StoreInt32((*int32)(a), 1) }

type Handler interface {
	ServeLDAP(ResponseWriter, *Request)
}

type HandlerFunc func(rw ResponseWriter, req *Request)

func (f HandlerFunc) ServeLDAP(rw ResponseWriter, req *Request) {
	f(rw, req)
}

type Message interface {
}

type Request struct {
	Context   context.Context
	MessageId int64
	Message   Message
}

type ResponseWriter interface {
	Write(msg Message) error
}

type response struct {
	messageId int64
	conn      net.Conn
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
				log.Errorf("kafka: Accept error: %v", err)
				continue
			}
		}

		ctx := s.trackConn(conn)
		go s.serve(conn, ctx)
	}
}

func (s *Server) Close() {
	if s.inShutdown.isSet() {
		return
	}
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
	ctx = NewPagingFromContext(ctx)
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
		if err != nil && (err == io.EOF || err.Error() == "unexpected EOF" || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET)) {
			return
		} else if err != nil {
			log.Debugf("ldap: handleConnection ber.ReadPacket ERROR: %v", err.Error())
			return
		}

		if len(packet.Children) < 2 {
			log.Infof("ldap: invalid packat length %v expected at least 2", len(packet.Children))
			return
		}
		messageId, ok := packet.Children[0].Value.(int64)
		if !ok {
			log.Infof("ldap: malformed messageId %v", reflect.TypeOf(packet.Children[0].Value))
			return
		}
		body := packet.Children[1]
		if body.ClassType != ber.ClassApplication {
			log.Infof("ldap: classType of packet is not ClassApplication was %v", body.ClassType)
			return
		}

		var controls []Control
		if len(packet.Children) > 2 {
			controls, err = decodeControls(packet.Children[2])
			if err != nil {
				log.Infof("parse controls failed: %v", err)
			}
		}

		var msg Message
		switch body.Tag {
		case bindRequest:
			msg, err = readBindRequest(body)
		case unbindRequest:
			log.Debugf("ldap: received unbind request")
			return
		case searchRequest:
			msg, err = decodeSearchRequest(body, controls)
		case abandonRequest:
			msg = &SearchResponse{Status: CannotCancel}
		default:
			log.Errorf("ldap: unknown operation %v, %v", packet.Tag, packet.Description)
		}

		if err != nil {
			log.Errorf("ldap error: %v", err)
		}

		s.Handler.ServeLDAP(&response{
			messageId: messageId,
			conn:      conn,
		}, &Request{
			Context:   ctx,
			MessageId: messageId,
			Message:   msg,
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

func (r *response) Write(msg Message) error {
	envelope := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
	envelope.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, r.messageId, "Message ID"))

	switch res := msg.(type) {
	case *BindResponse:
		return r.write(res.toPacket())
	case *SearchResponse:
		for _, p := range res.Results {
			env := r.getEnvelope()
			p.appendTo(env)
			if _, err := r.conn.Write(env.Bytes()); err != nil {
				return err
			}
		}
		env := r.getEnvelope()
		res.appendSearchDone(env)
		_, err := r.conn.Write(env.Bytes())
		return err
	default:
		return fmt.Errorf("unsupported message: %t", msg)
	}
}

func (r *response) write(body *ber.Packet) error {
	p := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
	p.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, r.messageId, "Message ID"))
	p.AppendChild(body)
	_, err := r.conn.Write(p.Bytes())
	return err
}

func (r *response) getEnvelope() *ber.Packet {
	p := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
	p.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, r.messageId, "Message ID"))
	return p
}
