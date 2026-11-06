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
