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
