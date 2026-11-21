package ldap

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
	"io"
	"mokapi/try"
	"net"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	testcases := []struct {
		name    string
		handler func(t *testing.T, rw ResponseWriter, req *Request)
		test    func(t *testing.T, c Client)
	}{
		{
			name: "SimpleBind",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				bind := req.Message.(*BindRequest)
				require.Equal(t, Simple, bind.Auth)
				require.Equal(t, "foo", bind.Name)
				require.Equal(t, "bar", bind.Password)
				err := rw.Write(&BindResponse{
					Result: Success,
				})
				require.NoError(t, err)
			},
			test: func(t *testing.T, c Client) {
				r, err := c.Bind("foo", "bar")
				require.NoError(t, err)
				require.Equal(t, Success, r.Result)
				require.Equal(t, "", r.MatchedDN)
				require.Equal(t, "", r.Message)
			},
		},
		{
			name: "unbind",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				panic("should not call handler")
			},
			test: func(t *testing.T, c Client) {
				err := c.Unbind()
				require.NoError(t, err)
			},
		},
		{
			name: "abandon",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				panic("should not call handler")
			},
			test: func(t *testing.T, c Client) {
				err := c.AbandonSearch(0)
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s, c := newTestServer(HandlerFunc(func(rw ResponseWriter, req *Request) {
				tc.handler(t, rw, req)
			}))
			defer s.Close()
			require.NoError(t, c.Dial())
			tc.test(t, c)
		})
	}
}

func newTestServer(h Handler) (*Server, Client) {
	p := try.GetFreePort()
	addr := fmt.Sprintf("127.0.0.1:%v", p)

	s := &Server{
		Addr:    addr,
		Handler: h,
	}

	go func() {
		err := s.ListenAndServe()
		if err != nil && !errors.Is(err, ErrServerClosed) {
			panic(err)
		}
	}()

	c := Client{Addr: addr}

	return s, c
}

func TestServer_Close(t *testing.T) {
	t.Run("close while serve starts is race-free", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			s := &Server{Addr: "127.0.0.1:0"}

			done := make(chan error, 1)
			go func() { done <- s.ListenAndServe() }()
			go s.Close()

			select {
			case err := <-done:
				require.ErrorIs(t, err, ErrServerClosed)
			case <-time.After(2 * time.Second):
				t.Fatal("ListenAndServe did not return after Close")
			}
		}
	})

	t.Run("close before listen", func(t *testing.T) {
		s := &Server{Addr: "127.0.0.1:0"}
		s.Close()

		done := make(chan error, 1)
		go func() { done <- s.ListenAndServe() }()

		select {
		case err := <-done:
			require.ErrorIs(t, err, ErrServerClosed)
		case <-time.After(2 * time.Second):
			t.Fatal("ListenAndServe did not return after Close")
		}
	})
}

func TestServer_Serve_ClosedServerClosesListener(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()

	s := &Server{}
	s.Close()

	require.ErrorIs(t, s.Serve(l), ErrServerClosed)

	l2, err := net.Listen("tcp", addr)
	require.NoError(t, err, "listener passed to Serve is still open")
	_ = l2.Close()
}

func TestServer_Close_ClosesListenerPassedToServe(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()

	s := &Server{}
	done := make(chan error, 1)
	go func() { done <- s.Serve(l) }()

	// Serve registers the listener asynchronously, a successful dial shows it accepts
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	_ = conn.Close()

	s.Close()
	select {
	case err = <-done:
		require.ErrorIs(t, err, ErrServerClosed)
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return after Close")
	}

	l2, err := net.Listen("tcp", addr)
	require.NoError(t, err, "Close did not close the listener passed to Serve")
	_ = l2.Close()
}

func TestServer_ListenAndServe_PortInUse(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	s := &Server{Addr: l.Addr().String()}
	err = s.ListenAndServe()
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrServerClosed)
}

func TestServer_ListenAndServe_ClosedServerDoesNotBind(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	s := &Server{Addr: l.Addr().String()}
	s.Close()
	require.ErrorIs(t, s.ListenAndServe(), ErrServerClosed)
}

func TestServer_MalformedRequest(t *testing.T) {
	str := func(v string) *ber.Packet {
		return ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, v, "")
	}
	integer := func(v int64) *ber.Packet {
		return ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, v, "")
	}
	sequence := func(children ...*ber.Packet) *ber.Packet {
		p := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "")
		for _, c := range children {
			p.AppendChild(c)
		}
		return p
	}
	constructed := func(class ber.Class, tag ber.Tag, children ...*ber.Packet) *ber.Packet {
		p := ber.Encode(class, ber.TypeConstructed, tag, nil, "")
		for _, c := range children {
			p.AppendChild(c)
		}
		return p
	}
	operation := func(tag ber.Tag, children ...*ber.Packet) *ber.Packet {
		return constructed(ber.ClassApplication, tag, children...)
	}
	message := func(children ...*ber.Packet) *ber.Packet { return sequence(children...) }

	testcases := []struct {
		name       string
		request    *ber.Packet
		diagnostic string
	}{
		{
			name: "search scope is a string",
			request: message(integer(1), operation(searchRequest,
				str(""), str("wholeSubtree"), integer(0), integer(0), integer(0),
				ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, false, ""),
				ber.NewString(ber.ClassContext, ber.TypePrimitive, 7, "objectClass", ""),
				sequence())),
			diagnostic: "field scope",
		},
		{
			name: "search with 7 children",
			request: message(integer(1), operation(searchRequest,
				str(""), integer(2), integer(0), integer(0), integer(0), integer(0), integer(0))),
			diagnostic: "search request length: 7",
		},
		{
			name:       "modify with 1 child",
			request:    message(integer(1), operation(modifyRequest, str("cn=foo"))),
			diagnostic: "modify request length: 1",
		},
		{
			// the decoder indexes the children of the change unchecked and panics
			name: "modify with an empty change",
			request: message(integer(1), operation(modifyRequest, str("cn=foo"),
				sequence(sequence()))),
			diagnostic: "index out of range [0] with length 0",
		},
		{
			name:       "add with non-string DN",
			request:    message(integer(1), operation(addRequest, integer(1), sequence())),
			diagnostic: "interface {} is int64, not string",
		},
		{
			name: "bind with non-int version",
			request: message(integer(1), operation(bindRequest, str("3"), str("foo"),
				ber.NewString(ber.ClassContext, ber.TypePrimitive, 0, "bar", ""))),
			diagnostic: "interface {} is string, not int64",
		},
		{
			name:       "modifyDN with 2 children",
			request:    message(integer(1), operation(modifyDNRequest, str("cn=foo"), str("cn=bar"))),
			diagnostic: "index out of range [2] with length 2",
		},
		{
			name: "compare with bad assertion",
			request: message(integer(1), operation(compareRequest, str("cn=foo"),
				sequence(str("cn")))),
			diagnostic: "invalid attribute value assertion",
		},
		{
			name:       "unknown tag",
			request:    message(integer(1), operation(20)),
			diagnostic: "does not support request tag 20",
		},
		{
			name:       "only 1 child",
			request:    message(integer(1)),
			diagnostic: "invalid packet length 1",
		},
		{
			name:       "messageId is a string",
			request:    message(str("1"), operation(deleteRequest)),
			diagnostic: "malformed messageId",
		},
		{
			name: "body is not application class",
			request: message(integer(1),
				constructed(ber.ClassContext, deleteRequest)),
			diagnostic: "not ClassApplication",
		},
		{
			name: "control with a non-string OID",
			request: message(integer(1),
				operation(deleteRequest),
				constructed(ber.ClassContext, 0, sequence(integer(1)))),
			diagnostic: "parse controls failed: invalid control type",
		},
		{
			// PagedResultsControl.Decode asserts the criticality unchecked and panics
			name: "paged results control with non-bool criticality",
			request: message(integer(1),
				operation(deleteRequest),
				constructed(ber.ClassContext, 0, sequence(str(PagedResultsControlType), str("true"), str("")))),
			diagnostic: "malformed request: interface conversion: interface {} is string, not bool",
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s, c := newTestServer(HandlerFunc(func(rw ResponseWriter, req *Request) {
				t.Errorf("handler must not be called, got %T", req.Message)
			}))
			defer s.Close()
			conn := dialRaw(t, c.Addr)

			_, err := conn.Write(tc.request.Bytes())
			require.NoError(t, err)

			requireNoticeOfDisconnection(t, conn, tc.diagnostic)
			requireClosed(t, conn)
		})
	}
}

func TestServer_ValidRequestAfterValidRequest(t *testing.T) {
	s, c := newTestServer(HandlerFunc(func(rw ResponseWriter, req *Request) {
		require.NoError(t, rw.Write(&BindResponse{Result: Success}))
	}))
	defer s.Close()
	require.NoError(t, c.Dial())

	for i := 0; i < 2; i++ {
		r, err := c.Bind("foo", "bar")
		require.NoError(t, err)
		require.Equal(t, Success, r.Result)
	}
}

func TestServer_UnbindClosesWithoutNotice(t *testing.T) {
	unbound := make(chan struct{})
	s, c := newTestServer(unbindRecorder{done: unbound})
	defer s.Close()
	conn := dialRaw(t, c.Addr)

	unbind := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "")
	unbind.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, 1, ""))
	unbind.AppendChild(ber.Encode(ber.ClassApplication, ber.TypePrimitive, unbindRequest, nil, ""))
	_, err := conn.Write(unbind.Bytes())
	require.NoError(t, err)

	requireClosed(t, conn)
	select {
	case <-unbound:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not receive unbind")
	}
}

type unbindRecorder struct{ done chan struct{} }

func (unbindRecorder) ServeLDAP(ResponseWriter, *Request) { panic("should not call handler") }
func (u unbindRecorder) Unbind(context.Context)           { close(u.done) }

func dialRaw(t *testing.T, addr string) net.Conn {
	var conn net.Conn
	require.Eventually(t, func() bool {
		var err error
		conn, err = net.DialTimeout("tcp", addr, time.Second)
		return err == nil
	}, 2*time.Second, 10*time.Millisecond)
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, conn.SetDeadline(time.Now().Add(3*time.Second)))
	return conn
}

func requireNoticeOfDisconnection(t *testing.T, conn net.Conn, diagnostic string) {
	p, err := ber.ReadPacket(conn)
	require.NoError(t, err, "server should answer with a Notice of Disconnection")
	require.Len(t, p.Children, 2)
	require.Equal(t, int64(0), p.Children[0].Value, "messageID of unsolicited notification")
	res := p.Children[1]
	require.Equal(t, ber.ClassApplication, res.ClassType)
	require.Equal(t, ber.Tag(extendedResponse), res.Tag)
	require.Len(t, res.Children, 4)
	require.Equal(t, int64(ProtocolError), res.Children[0].Value, "resultCode")
	require.Equal(t, "", res.Children[1].Value, "matchedDN")
	require.Contains(t, res.Children[2].Value, diagnostic, "diagnosticMessage")
	require.Equal(t, noticeOfDisconnectionOID, res.Children[3].Data.String(), "responseName")
}

func requireClosed(t *testing.T, conn net.Conn) {
	_, err := conn.Read(make([]byte, 1))
	require.ErrorIs(t, err, io.EOF, "server should close the connection")
}

func TestServer_UnbindWithMalformedControlStillUnbinds(t *testing.T) {
	unbound := make(chan struct{})
	s, c := newTestServer(unbindRecorder{done: unbound})
	defer s.Close()
	conn := dialRaw(t, c.Addr)

	controls := ber.Encode(ber.ClassContext, ber.TypeConstructed, 0, nil, "")
	control := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "")
	control.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, 1, ""))
	controls.AppendChild(control)
	unbind := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "")
	unbind.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, 1, ""))
	unbind.AppendChild(ber.Encode(ber.ClassApplication, ber.TypePrimitive, unbindRequest, nil, ""))
	unbind.AppendChild(controls)
	_, err := conn.Write(unbind.Bytes())
	require.NoError(t, err)

	requireClosed(t, conn)
	select {
	case <-unbound:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not receive unbind")
	}
}
