package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
)

func TestServer(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, c *imaptest.Client)
	}{
		{
			name: "expect greeting",
			test: func(t *testing.T, c *imaptest.Client) {
				g, err := c.Dial()
				require.NoError(t, err)
				require.Equal(t, "* OK [CAPABILITY IMAP4rev1 SASL-IR AUTH=PLAIN] Mokapi Ready", g)
			},
		},
		{
			name: "unknown command",
			test: func(t *testing.T, c *imaptest.Client) {
				mustDial(t, c)
				r, err := c.Send("foo")
				require.NoError(t, err)
				require.Equal(t, "A1 BAD Unknown command", r[0])
			},
		},
		{
			name: "capability",
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				lines, err := c.Send("CAPABILITY")
				require.NoError(t, err)
				require.Equal(t, "* CAPABILITY IMAP4rev1 SASL-IR AUTH=PLAIN", lines[0])
				require.Equal(t, "A1 OK CAPABILITY completed", lines[1])
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := try.GetFreePort()
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

func mustDial(t *testing.T, c *imaptest.Client) {
	_, err := c.Dial()
	require.NoError(t, err)
}
