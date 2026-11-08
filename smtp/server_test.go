package smtp_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/imap"
	"mokapi/server/cert"
	"mokapi/smtp"
	"mokapi/try"
	"net"
	"testing"
	"time"
)

func TestServer_Tls(t *testing.T) {
	p := try.GetFreePort()
	store, err := cert.NewStore(&static.Config{})
	require.NoError(t, err)

	s := &smtp.Server{
		Addr:    fmt.Sprintf(":%v", p),
		TlsMode: smtp.Implicit,
		TLSConfig: &tls.Config{
			GetCertificate: store.GetCertificate,
		},
	}

	go func() {
		err := s.ListenAndServe()
		require.ErrorIs(t, err, imap.ErrServerClosed)
	}()

	c := smtp.NewClient(fmt.Sprintf("localhost:%v", p))
	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(cert.DefaultRootCert())
	cfg := &tls.Config{
		RootCAs: rootCAs,
	}
	res, err := c.DialTls(cfg)
	require.NoError(t, err)
	// should not contain StartTls
	require.Equal(t, "localhost ESMTP Service Ready", res)

	var lines []string
	lines, err = c.Write("EHLO TEST mokapi", 250)
	require.NoError(t, err)
	// should not contain StartTls
	require.Equal(t, []string{
		"Hello TEST mokapi",
		"AUTH LOGIN PLAIN",
	}, lines)
}

func TestServer_StartTls(t *testing.T) {
	p := try.GetFreePort()
	store, err := cert.NewStore(&static.Config{})
	require.NoError(t, err)

	s := &smtp.Server{
		Addr:    fmt.Sprintf(":%v", p),
		TlsMode: smtp.StartTls,
		TLSConfig: &tls.Config{
			GetCertificate: store.GetCertificate,
		},
	}

	go func() {
		err := s.ListenAndServe()
		require.ErrorIs(t, err, imap.ErrServerClosed)
	}()

	c := smtp.NewClient(fmt.Sprintf("localhost:%v", p))
	res, err := c.Dial()
	require.NoError(t, err)
	// should not contain StartTls
	require.Equal(t, "localhost ESMTP Service Ready", res)

	var lines []string
	lines, err = c.Write("EHLO TEST mokapi", 250)
	require.NoError(t, err)
	// should contain StartTls
	require.Equal(t, []string{
		"Hello TEST mokapi",
		"AUTH LOGIN PLAIN",
		"STARTTLS",
	}, lines)

	lines, err = c.Write("STARTTLS", 220)
	require.NoError(t, err)
	// should contain StartTls
	require.Equal(t, []string{
		"[2 0 0] Starting TLS",
	}, lines)
}

func TestServer_Close(t *testing.T) {
	t.Run("close while serve starts is race-free", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			s := &smtp.Server{Addr: "127.0.0.1:0"}

			done := make(chan error, 1)
			go func() { done <- s.ListenAndServe() }()
			go s.Close()

			select {
			case err := <-done:
				require.ErrorIs(t, err, smtp.ErrServerClosed)
			case <-time.After(2 * time.Second):
				t.Fatal("ListenAndServe did not return after Close")
			}
		}
	})

	t.Run("close before listen", func(t *testing.T) {
		s := &smtp.Server{Addr: "127.0.0.1:0"}
		s.Close()

		done := make(chan error, 1)
		go func() { done <- s.ListenAndServe() }()

		select {
		case err := <-done:
			require.ErrorIs(t, err, smtp.ErrServerClosed)
		case <-time.After(2 * time.Second):
			t.Fatal("ListenAndServe did not return after Close")
		}
	})
}

func TestServer_Serve_ClosedServerClosesListener(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()

	s := &smtp.Server{}
	s.Close()

	require.ErrorIs(t, s.Serve(l), smtp.ErrServerClosed)

	l2, err := net.Listen("tcp", addr)
	require.NoError(t, err, "listener passed to Serve is still open")
	_ = l2.Close()
}

func TestServer_Close_ClosesListenerPassedToServe(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()

	s := &smtp.Server{}
	done := make(chan error, 1)
	go func() { done <- s.Serve(l) }()

	// Serve registers the listener asynchronously, a successful dial shows it accepts
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	_ = conn.Close()

	s.Close()
	select {
	case err = <-done:
		require.ErrorIs(t, err, smtp.ErrServerClosed)
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

	s := &smtp.Server{Addr: l.Addr().String()}
	err = s.ListenAndServe()
	require.Error(t, err)
	require.NotErrorIs(t, err, smtp.ErrServerClosed)
}

func TestServer_ListenAndServe_ClosedServerDoesNotBind(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	s := &smtp.Server{Addr: l.Addr().String()}
	s.Close()
	require.ErrorIs(t, s.ListenAndServe(), smtp.ErrServerClosed)
}
