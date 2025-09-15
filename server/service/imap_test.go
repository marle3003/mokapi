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
				require.Equal(t, []string{"IMAP4rev1", "SASL-IR", "STARTTLS", "AUTH=PLAIN"}, res)
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
				require.Equal(t, []string{"IMAP4rev1", "SASL-IR", "AUTH=PLAIN"}, res)
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
			s.Start()
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
