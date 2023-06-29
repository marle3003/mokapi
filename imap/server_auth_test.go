package imap

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap/imaptest"
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
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p, err := try.GetFreePort()
			require.NoError(t, err)
			s := &Server{Addr: fmt.Sprintf(":%v", p)}
			defer s.Close()
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, ErrServerClosed)
			}()

			c := imaptest.NewClient(fmt.Sprintf("localhost:%v", p))

			tc.test(t, c)
		})
	}
}
