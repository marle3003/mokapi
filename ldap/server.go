package ldap

import (
	"context"
	"fmt"
	"io"
	"net"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

var ErrServerClosed = errors.New("ldap: Server closed")

type atomicBool int32

func (a *atomicBool) isSet() bool { return atomic.LoadInt32((*int32)(a)) != 0 }
func (a *atomicBool) setFalse()   { atomic.StoreInt32((*int32)(a), 0) }
func (a *atomicBool) setTrue()    { atomic.StoreInt32((*int32)(a), 1) }

type Handler interface {
	ServeLDAP(ResponseWriter, *Request)
	Unbind(ctx context.Context)
}

type HandlerFunc func(rw ResponseWriter, req *Request)

func (f HandlerFunc) ServeLDAP(rw ResponseWriter, req *Request) {
	f(rw, req)
}

func (f HandlerFunc) Unbind(ctx context.Context) {}

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
	activeConn map[net.Conn]context.Context
	listener   net.Listener
	inShutdown atomicBool
}

func (s *Server) ListenAndServe() error {
	if s.inShutdown.isSet() {
		return ErrServerClosed
	}
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Serve(l)
}

func (s *Server) Serve(l net.Listener) error {
	s.mu.Lock()
	// Close sets inShutdown before taking s.mu, so either we see it here or Close sees our listener.
	if s.inShutdown.isSet() {
		s.mu.Unlock()
		_ = l.Close()
		return ErrServerClosed
	}
	s.listener = l
	s.mu.Unlock()

	for {
		conn, err := l.Accept()
		if err != nil {
			if s.inShutdown.isSet() {
				return ErrServerClosed
			}
			log.Errorf("ldap: Accept error: %v", err)
			continue
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
		messageId, err := decodeMessageId(packet)
		if err != nil {
			disconnectWithNotice(conn, err)
			return
		}

		if packet.Children[1].Tag == unbindRequest {
			log.Debugf("ldap: received unbind request")
			s.Handler.Unbind(ctx)
			return
		}

		msg, err := decodeRequest(packet)
		if err != nil {
			disconnectWithNotice(conn, err)
			return
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

// decodeMessageId checks the LDAPMessage envelope: message ID and protocol operation.
func decodeMessageId(packet *ber.Packet) (int64, error) {
	if len(packet.Children) < 2 {
		return 0, fmt.Errorf("invalid packet length %v, expected at least 2", len(packet.Children))
	}
	messageId, ok := packet.Children[0].Value.(int64)
	if !ok {
		return 0, fmt.Errorf("malformed messageId %v", reflect.TypeOf(packet.Children[0].Value))
	}
	if class := packet.Children[1].ClassType; class != ber.ClassApplication {
		return 0, fmt.Errorf("classType of packet is not ClassApplication, was %v", class)
	}
	return messageId, nil
}

// decodeRequest decodes controls and the protocol operation of a checked envelope.
// Some decoders index and type-assert unchecked, so a malformed request can panic;
// that is reported as an error as well.
func decodeRequest(packet *ber.Packet) (msg Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Debugf("ldap: decoder panic: %v", string(debug.Stack()))
			msg, err = nil, fmt.Errorf("malformed request: %v", r)
		}
	}()

	var controls []Control
	if len(packet.Children) > 2 {
		controls, err = decodeControls(packet.Children[2])
		if err != nil {
			return nil, fmt.Errorf("parse controls failed: %w", err)
		}
	}

	body := packet.Children[1]
	switch body.Tag {
	case bindRequest:
		return readBindRequest(body)
	case searchRequest:
		return decodeSearchRequest(body, controls)
	case abandonRequest:
		return &SearchResponse{Status: CannotCancel}, nil
	case modifyRequest:
		return decodeModifyRequest(body)
	case addRequest:
		return decodeAddRequest(body)
	case deleteRequest:
		return decodeDeleteRequest(body)
	case modifyDNRequest:
		return decodeModifyDNRequest(body)
	case compareRequest:
		return decodeCompareRequest(body)
	default:
		return nil, fmt.Errorf("server does not support request tag %v", body.Tag)
	}
}

// disconnectWithNotice reports a request that can't be processed as protocolError
// with a Notice of Disconnection. The caller closes the connection (RFC 4511 4.1.1).
func disconnectWithNotice(conn net.Conn, cause error) {
	log.Infof("ldap: %v", cause)
	notice := &noticeOfDisconnection{ResultCode: ProtocolError, Message: cause.Error()}
	r := &response{messageId: 0, conn: conn}
	if err := r.write(notice.toPacket()); err != nil {
		log.Debugf("ldap: send notice of disconnection failed: %v", err)
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
	case *ModifyResponse:
		res.encode(envelope)
	case *AddResponse:
		res.encode(envelope)
	case *DeleteResponse:
		res.encode(envelope)
	case *ModifyDNResponse:
		res.encode(envelope)
	case *CompareResponse:
		res.encode(envelope)
	default:
		return fmt.Errorf("unsupported message: %t", msg)
	}

	_, err := r.conn.Write(envelope.Bytes())
	return err
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
