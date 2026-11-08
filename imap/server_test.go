package imap_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"mokapi/config/static"
	"mokapi/imap"
	"mokapi/server/cert"
	"mokapi/try"
	"net"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		test func(t *testing.T, c *imap.Client)
	}{
		{
			name: "expect greeting",
			test: func(t *testing.T, c *imap.Client) {
				caps, err := c.Dial()
				require.NoError(t, err)
				require.Equal(t, []string{"IMAP4rev1", "AUTH=PLAIN", "SASL-IR"}, caps)
			},
		},
		{
			name: "unknown command",
			test: func(t *testing.T, c *imap.Client) {
				mustDial(t, c)
				r, err := c.Send("foo")
				require.NoError(t, err)
				require.Equal(t, "A0001 BAD Unknown command FOO", r[0])
			},
		},
		{
			name: "capability",
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				caps, err := c.Capability()
				require.NoError(t, err)
				require.Equal(t, []string{"IMAP4rev1", "AUTH=PLAIN", "SASL-IR"}, caps)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := try.GetFreePort()
			s := &imap.Server{Addr: fmt.Sprintf(":%v", p)}
			defer s.Close()
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imap.NewClient(fmt.Sprintf("localhost:%v", p))

			tc.test(t, c)
		})
	}
}

func TestServer_Tls(t *testing.T) {
	p := try.GetFreePort()
	store, err := cert.NewStore(&static.Config{})
	require.NoError(t, err)

	s := &imap.Server{
		Addr:    fmt.Sprintf(":%v", p),
		TlsMode: imap.Implicit,
		TLSConfig: &tls.Config{
			GetCertificate: store.GetCertificate,
		},
	}

	go func() {
		err := s.ListenAndServe()
		require.ErrorIs(t, err, imap.ErrServerClosed)
	}()

	c := imap.NewClient(fmt.Sprintf("localhost:%v", p))
	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(cert.DefaultRootCert())
	cfg := &tls.Config{
		RootCAs: rootCAs,
	}
	res, err := c.DialTls(cfg)
	require.NoError(t, err)
	// should not contain StartTls
	require.Equal(t, []string{"IMAP4rev1", "AUTH=PLAIN", "SASL-IR"}, res)
}

func mustDial(t *testing.T, c *imap.Client) {
	_, err := c.Dial()
	require.NoError(t, err)
}

func TestServer_Serve_ClosedServerClosesListener(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()

	s := &imap.Server{}
	s.Close()

	require.ErrorIs(t, s.Serve(l), imap.ErrServerClosed)

	l2, err := net.Listen("tcp", addr)
	require.NoError(t, err, "listener passed to Serve is still open")
	_ = l2.Close()
}

func TestServer_Close_ClosesListenerPassedToServe(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()

	s := &imap.Server{}
	done := make(chan error, 1)
	go func() { done <- s.Serve(l) }()

	// Serve registers the listener asynchronously, a successful dial shows it accepts
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	_ = conn.Close()

	s.Close()
	select {
	case err = <-done:
		require.ErrorIs(t, err, imap.ErrServerClosed)
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

	s := &imap.Server{Addr: l.Addr().String()}
	err = s.ListenAndServe()
	require.Error(t, err)
	require.NotErrorIs(t, err, imap.ErrServerClosed)
}

func TestServer_ListenAndServe_ClosedServerDoesNotBind(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	s := &imap.Server{Addr: l.Addr().String()}
	s.Close()
	require.ErrorIs(t, s.ListenAndServe(), imap.ErrServerClosed)
}

func TestServer_Close_Imap(t *testing.T) {
	t.Run("close while serve starts is race-free", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			s := &imap.Server{Addr: "127.0.0.1:0"}

			done := make(chan error, 1)
			go func() { done <- s.ListenAndServe() }()
			go s.Close()

			select {
			case err := <-done:
				require.ErrorIs(t, err, imap.ErrServerClosed)
			case <-time.After(2 * time.Second):
				t.Fatal("ListenAndServe did not return after Close")
			}
		}
	})

	t.Run("close before listen", func(t *testing.T) {
		s := &imap.Server{Addr: "127.0.0.1:0"}
		s.Close()

		done := make(chan error, 1)
		go func() { done <- s.ListenAndServe() }()

		select {
		case err := <-done:
			require.ErrorIs(t, err, imap.ErrServerClosed)
		case <-time.After(2 * time.Second):
			t.Fatal("ListenAndServe did not return after Close")
		}
	})
}
