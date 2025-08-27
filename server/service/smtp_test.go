package service_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"mokapi/config/static"
	"mokapi/providers/mail"
	"mokapi/server/cert"
	"mokapi/server/service"
	"mokapi/smtp"
	"mokapi/try"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmtpServer(t *testing.T) {
	testcases := []struct {
		name    string
		h       *mail.Handler
		cert    func() *cert.Store
		tlsMode smtp.TlsMode
		test    func(t *testing.T, c *smtp.Client)
	}{
		{
			name: "StartTLS",
			h:    &mail.Handler{},
			cert: func() *cert.Store {
				s, _ := cert.NewStore(&static.Config{})
				return s
			},
			tlsMode: smtp.StartTls,
			test: func(t *testing.T, c *smtp.Client) {
				res, err := c.Dial()
				require.NoError(t, err)
				require.Equal(t, "localhost ESMTP Service Ready", res)
			},
		},
		{
			name: "Implicit",
			h:    &mail.Handler{},
			cert: func() *cert.Store {
				s, _ := cert.NewStore(&static.Config{})
				return s
			},
			tlsMode: smtp.Implicit,
			test: func(t *testing.T, c *smtp.Client) {
				rootCAs := x509.NewCertPool()
				rootCAs.AddCert(cert.DefaultRootCert())
				cfg := &tls.Config{
					RootCAs: rootCAs,
				}

				res, err := c.DialTls(cfg)
				require.NoError(t, err)
				require.Equal(t, "localhost ESMTP Service Ready", res)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			port := fmt.Sprintf("%v", try.GetFreePort())
			s := service.NewSmtpServer(port, tc.h, tc.cert(), tc.tlsMode)
			s.Start()
			defer s.Stop()

			c := smtp.NewClient(fmt.Sprintf("localhost:%v", port))

			tc.test(t, c)
		})
	}
}

func TestSmtpServer_Addr(t *testing.T) {
	s := service.NewSmtpServer("1234", nil, nil, smtp.StartTls)
	require.Equal(t, ":1234", s.Addr())
}
