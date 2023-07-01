package imap_test

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/sasl"
	"mokapi/try"
	"testing"
)

func TestServer_Auth(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, c *imaptest.Client)
	}{
		{
			name: "send unsupported mechanism",
			test: func(t *testing.T, c *imaptest.Client) {
				mustDial(t, c)
				r, err := c.Send("AUTHENTICATE foo")
				require.NoError(t, err)
				require.Equal(t, "A1 NO Unsupported authentication mechanism", r)
			},
		},
		{
			name: "plain wrong format",
			test: func(t *testing.T, c *imaptest.Client) {
				mustDial(t, c)
				r, err := c.Send("AUTHENTICATE PLAIN")
				require.NoError(t, err)
				require.Equal(t, "+ ", r)
				r, err = c.SendRaw("foo")
				require.NoError(t, err)
				require.Equal(t, "A1 BAD Invalid response", r)
			},
		},
		{
			name: "plain without initial",
			test: func(t *testing.T, c *imaptest.Client) {
				mustDial(t, c)
				r, err := c.Send("AUTHENTICATE PLAIN")
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
			test: func(t *testing.T, c *imaptest.Client) {
				mustDial(t, c)
				_, err := c.Send("AUTHENTICATE PLAIN")
				require.NoError(t, err)
				saslClient := sasl.NewPlainClient("", "bob", "password")
				secret, err := saslClient.Next(nil)
				_, err = c.SendRaw(base64.StdEncoding.EncodeToString(secret))
				require.NoError(t, err)
				r, err := c.Send("CAPABILITY")
				require.NoError(t, err)
				require.Equal(t, "* CAPABILITY IMAP4rev1 SELECT", r)
				r, err = c.ReadLine()
				require.NoError(t, err)
				require.Equal(t, "A2 OK CAPABILITY completed", r)
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
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imaptest.NewClient(fmt.Sprintf("localhost:%v", p))

			tc.test(t, c)
		})
	}
}
