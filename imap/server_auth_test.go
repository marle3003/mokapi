package imap_test

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
)

func TestServer_Auth(t *testing.T) {
	testcases := []struct {
		name    string
		handler newHandler
		test    func(t *testing.T, c *imaptest.Client)
	}{
		{
			name: "plain",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					LoginFunc: func(username, password string, session map[string]interface{}) error {
						require.Equal(t, "bob", username)
						require.Equal(t, "password", password)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imaptest.Client) {
				mustDial(t, c)
				err := c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
			},
		},
		{
			name: "plain wrong credentials",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					LoginFunc: func(username, password string, session map[string]interface{}) error {
						return fmt.Errorf("wrong password")
					},
				}
				return h
			},
			test: func(t *testing.T, c *imaptest.Client) {
				mustDial(t, c)
				err := c.PlainAuth("", "bob", "password")
				require.EqualError(t, err, "A1 BAD wrong password")
			},
		},
		{
			name: "send unsupported mechanism",
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{}
			},
			test: func(t *testing.T, c *imaptest.Client) {
				mustDial(t, c)
				r, err := c.SendRaw("A1 AUTHENTICATE foo")
				require.NoError(t, err)
				require.Equal(t, "A1 NO Unsupported authentication mechanism", r)
			},
		},
		{
			name: "plain wrong format",
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{}
			},
			test: func(t *testing.T, c *imaptest.Client) {
				mustDial(t, c)
				r, err := c.SendRaw("A1 AUTHENTICATE PLAIN")
				require.NoError(t, err)
				require.Equal(t, "+ ", r)
				r, err = c.SendRaw("foo")
				require.NoError(t, err)
				require.Equal(t, "A1 BAD Invalid response", r)
			},
		},
		{
			name: "plain without initial",
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{}
			},
			test: func(t *testing.T, c *imaptest.Client) {
				mustDial(t, c)
				r, err := c.SendRaw("A1 AUTHENTICATE PLAIN")
				require.NoError(t, err)
				require.Equal(t, "+ ", r)
				secret := "\x00" + "bob" + "\x00" + "password"
				secret = base64.StdEncoding.EncodeToString([]byte(secret))
				r, err = c.SendRaw(secret)
				require.NoError(t, err)
				require.Equal(t, "A1 OK Authenticated", r)
			},
		},
		{
			name: "capability after auth",
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{}
			},
			test: func(t *testing.T, c *imaptest.Client) {
				mustDial(t, c)
				err := c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				lines, err := c.Send("CAPABILITY")
				require.NoError(t, err)
				require.Equal(t, "* CAPABILITY IMAP4rev1 SASL-IR SELECT LIST FETCH CLOSE", lines[0])
				require.Equal(t, "A2 OK CAPABILITY completed", lines[1])
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p, err := try.GetFreePort()
			require.NoError(t, err)
			s := &imap.Server{
				Addr:    fmt.Sprintf(":%v", p),
				Handler: tc.handler(t),
			}
			defer s.Close()
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imaptest.NewClient(fmt.Sprintf("localhost:%v", p))

			tc.test(t, c)
		})
	}
}
