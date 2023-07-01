package imap_test

import (
	"crypto/tls"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/server/cert"
	"mokapi/try"
	"testing"
)

func TestServer_StartTLS(t *testing.T) {
	testcases := []struct {
		name      string
		tlsConfig func() *tls.Config
		test      func(t *testing.T, c *imaptest.Client)
	}{
		{
			name: "expect greetings with STARTTLS",
			tlsConfig: func() *tls.Config {
				store, err := cert.NewStore(&static.Config{})
				if err != nil {
					panic(err)
				}
				return &tls.Config{GetCertificate: store.GetCertificate}
			},
			test: func(t *testing.T, c *imaptest.Client) {
				r, err := c.Dial()
				require.NoError(t, err)
				require.Equal(t, "* OK [CAPABILITY IMAP4rev1 STARTTLS AUTH=PLAIN] Mokapi Ready", r)
			},
		},
		{
			name: "start TLS",
			tlsConfig: func() *tls.Config {
				store, err := cert.NewStore(&static.Config{})
				if err != nil {
					panic(err)
				}
				return &tls.Config{GetCertificate: store.GetCertificate}
			},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				r, err := c.StartTLS()
				require.NoError(t, err)
				require.Equal(t, "A1 OK Begin TLS negotiation now", r)
				caps, err := c.Capability()
				require.NoError(t, err)
				require.NotContains(t, caps, "STARTTLS")
			},
		},
		{
			name: "after auth starttls is not available",
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				caps, err := c.Capability()
				require.NoError(t, err)
				require.NotContains(t, caps, "STARTTLS")
			},
		},
		{
			name: "STARTTLS not available",
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				r, err := c.StartTLS()
				require.NoError(t, err)
				require.Equal(t, "A1 BAD STARTTLS not available", r)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p, err := try.GetFreePort()
			require.NoError(t, err)
			s := &imap.Server{Addr: fmt.Sprintf(":%v", p)}
			defer s.Close()
			if tc.tlsConfig != nil {
				s.TLSConfig = tc.tlsConfig()
			}
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imaptest.NewClient(fmt.Sprintf("localhost:%v", p))

			tc.test(t, c)
		})
	}
}
