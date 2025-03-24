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
		test      func(t *testing.T, c *imap.Client)
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
			test: func(t *testing.T, c *imap.Client) {
				caps, err := c.Dial()
				require.NoError(t, err)
				require.Equal(t, []string{"IMAP4rev1", "SASL-IR", "STARTTLS", "AUTH=PLAIN"}, caps)
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
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.StartTLS()
				require.NoError(t, err)
				caps, err := c.Capability()
				require.NoError(t, err)
				require.NotContains(t, caps, "STARTTLS")
			},
		},
		{
			name: "after auth starttls is not available",
			test: func(t *testing.T, c *imap.Client) {
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
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.StartTLS()
				require.EqualError(t, err, "imap status [BAD]: STARTTLS not available")
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := try.GetFreePort()
			s := &imap.Server{
				Addr: fmt.Sprintf(":%v", p),
				Handler: &imaptest.Handler{
					LoginFunc: func(username, password string, session map[string]interface{}) error {
						return nil
					},
				},
			}
			defer s.Close()
			if tc.tlsConfig != nil {
				s.TLSConfig = tc.tlsConfig()
			}
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imap.NewClient(fmt.Sprintf("localhost:%v", p))

			tc.test(t, c)
		})
	}
}
