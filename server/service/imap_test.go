package service_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"mokapi/config/static"
	"mokapi/imap"
	"mokapi/providers/mail"
	"mokapi/server/cert"
	"mokapi/server/service"
	"mokapi/try"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImapServer(t *testing.T) {
	testcases := []struct {
		name    string
		h       *mail.Handler
		cert    func() *cert.Store
		tlsMode imap.TlsMode
		test    func(t *testing.T, c *imap.Client)
	}{
		{
			name: "StartTLS",
			h:    &mail.Handler{},
			cert: func() *cert.Store {
				s, _ := cert.NewStore(&static.Config{})
				return s
			},
			tlsMode: imap.StartTls,
			test: func(t *testing.T, c *imap.Client) {
				res, err := c.Dial()
				require.NoError(t, err)
				require.Equal(t, []string{"IMAP4rev1", "STARTTLS", "AUTH=PLAIN", "SASL-IR"}, res)
			},
		},
		{
			name: "Implicit",
			h:    &mail.Handler{},
			cert: func() *cert.Store {
				s, _ := cert.NewStore(&static.Config{})
				return s
			},
			tlsMode: imap.Implicit,
			test: func(t *testing.T, c *imap.Client) {
				rootCAs := x509.NewCertPool()
				rootCAs.AddCert(cert.DefaultRootCert())
				cfg := &tls.Config{
					RootCAs: rootCAs,
				}

				res, err := c.DialTls(cfg)
				require.NoError(t, err)
				require.Equal(t, []string{"IMAP4rev1", "AUTH=PLAIN", "SASL-IR"}, res)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			port := fmt.Sprintf("%v", try.GetFreePort())
			s := service.NewImapServer(port, tc.h, tc.cert(), tc.tlsMode)
			require.NoError(t, s.Start())
			defer s.Stop()

			c := imap.NewClient(fmt.Sprintf("localhost:%v", port))

			tc.test(t, c)
		})
	}
}

func TestImapServer_Addr(t *testing.T) {
	s := service.NewImapServer("1234", nil, nil, imap.None)
	require.Equal(t, ":1234", s.Addr())
}

func newImapServer(tlsMode imap.TlsMode) func(port string) *service.ImapServer {
	return func(port string) *service.ImapServer {
		store, _ := cert.NewStore(&static.Config{})
		return service.NewImapServer(port, &mail.Handler{}, store, tlsMode)
	}
}

func TestImapServer_Start_PortIsBoundWhenStartReturns(t *testing.T) {
	t.Run("imap", func(t *testing.T) { requirePortIsBoundWhenStartReturns(t, newImapServer(imap.None)) })
	t.Run("imaps", func(t *testing.T) { requirePortIsBoundWhenStartReturns(t, newImapServer(imap.Implicit)) })
}

func TestImapServer_Start_ReturnsErrorWhenPortIsInUse(t *testing.T) {
	t.Run("imap", func(t *testing.T) { requireStartReturnsErrorWhenPortIsInUse(t, newImapServer(imap.None)) })
	t.Run("imaps", func(t *testing.T) { requireStartReturnsErrorWhenPortIsInUse(t, newImapServer(imap.Implicit)) })
}

func TestImapServer_Stop_FreesPort(t *testing.T) {
	t.Run("imap", func(t *testing.T) { requireStopFreesPort(t, newImapServer(imap.None)) })
	t.Run("imaps", func(t *testing.T) { requireStopFreesPort(t, newImapServer(imap.Implicit)) })
}
