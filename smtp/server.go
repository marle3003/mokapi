package smtp

import (
	"context"
	"crypto/tls"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"sync/atomic"
)

var ErrServerClosed = errors.New("ldap: Server closed")

type atomicBool int32

func (a *atomicBool) isSet() bool { return atomic.LoadInt32((*int32)(a)) != 0 }
func (a *atomicBool) setFalse()   { atomic.StoreInt32((*int32)(a), 0) }
func (a *atomicBool) setTrue()    { atomic.StoreInt32((*int32)(a), 1) }

type Handler interface {
	ServeSMTP(rw ResponseWriter, req Request)
}

type HandlerFunc func(rw ResponseWriter, req Request)

func (f HandlerFunc) ServeSMTP(rw ResponseWriter, req Request) {
	f(rw, req)
}

type Request interface {
	Context() context.Context
	WithContext(ctx context.Context)
}

type Response interface {
	write(conn *textproto.Conn) error
}

type Command int

type ResponseWriter interface {
	Write(r Response) error
}

type response struct {
	conn *textproto.Conn
}

type Server struct {
	Addr      string
	Handler   Handler
	TLSConfig *tls.Config

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

func (s *Server) ListenAndServeTLS() error {
	if s.inShutdown.isSet() {
		return ErrServerClosed
	}

	var err error
	s.listener, err = tls.Listen("tcp", s.Addr, s.TLSConfig)
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
				log.Errorf("smtp: Accept error: %v", err)
				continue
			}
		}

		conn := conn{
			server: s,
			conn:   rw,
			ctx:    s.trackConn(rw),
		}
		go conn.serve()
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

//func (s *Server) serve(conn net.Conn, ctx context.Context) {
//	ctx, cancel := context.WithCancel(ctx)
//	defer func() {
//		r := recover()
//		if r != nil {
//			log.Debugf("smtp panic: %v", string(debug.Stack()))
//			log.Errorf("smtp panic: %v", r)
//		}
//		cancel()
//		s.closeConn(conn)
//	}()
//
//	client := ctx.Value(clientKey).(*ClientContext)
//
//	tc := textproto.NewConn(conn)
//	tc.PrintfLine("220 localhost ESMTP Service Ready")
//
//	for {
//		line, err := tc.ReadLine()
//		if err != nil {
//			switch {
//			case err == io.EOF || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET):
//				return
//			default:
//				log.Errorf("smtp: %v", err)
//				return
//			}
//		}
//
//		cmd, param := parseLine(line)
//
//		r := &Request{
//			Context: ctx,
//		}
//
//		switch cmd {
//		case "EHLO":
//			client.Client = param
//			client.Proto = "ESMTP"
//			r.Cmd = Hello
//		case "AUTH":
//			err := s.serveAuth(tc, param)
//			if err != nil {
//				log.Errorf("smtp error: %v", err)
//				return
//			}
//			continue
//		case "MAIL":
//			if !strings.HasPrefix(param, "FROM") && len(param) < 7 {
//				write(tc, StatusSyntaxError, SyntaxError, "Expected parameter FROM:<address>")
//				return
//			}
//			r.Cmd = From
//			r.Param = param[len("FROM:<") : len(param)-1]
//		case "RCPT":
//			if !strings.HasPrefix(param, "TO") && len(param) < 5 {
//				write(tc, StatusSyntaxError, SyntaxError, "Expected parameter TO:<address>")
//				return
//			}
//			r.Cmd = Recipient
//			r.Param = param[len("TO:<") : len(param)-1]
//		case "DATA":
//			write(tc, StatusStartMailInput, Success, "Send message, ending in CRLF.CRLF")
//			msg, err := ReadMessage(tc.Reader)
//			if err != nil {
//				if err == io.EOF {
//					return
//				} else {
//					log.Errorf("smtp: %v", err)
//					write(tc, StatusSyntaxError, SyntaxError, err.Error())
//				}
//			}
//			r.Cmd = Data
//			r.Message = msg
//		case "QUIT":
//			r.Cmd = Quit
//		default:
//			log.Debugf("unknown smtp command: %v", cmd)
//			continue
//		}
//
//		r.Proto = client.Proto
//
//		s.Handler.Serve(&response{
//			conn: tc,
//		}, r)
//	}
//}

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

func (r *response) Write(res Response) error {
	return res.write(r.conn)
}

func parseLine(line string) (cmd string, param string) {
	a := strings.SplitN(line, " ", 2)
	cmd = strings.ToUpper(a[0])
	if len(a) == 2 {
		param = strings.TrimSpace(a[1])
	}
	return
}

func write(conn *textproto.Conn, code StatusCode, status EnhancedStatusCode, lines ...string) error {
	for _, line := range lines[:len(lines)-1] {
		err := conn.PrintfLine("%v-%v", code, line)
		if err != nil {
			return err
		}
	}
	if status == Undefined {
		return conn.PrintfLine("%v %v", code, lines[len(lines)-1])
	} else {
		return conn.PrintfLine("%v %v %v", code, status, lines[len(lines)-1])
	}
}
