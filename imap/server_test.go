package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/try"
	"testing"
)

func TestServer(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, c *imap.Client)
	}{
		{
			name: "expect greeting",
			test: func(t *testing.T, c *imap.Client) {
				caps, err := c.Dial()
				require.NoError(t, err)
				require.Equal(t, []string{"IMAP4rev1", "SASL-IR", "AUTH=PLAIN"}, caps)
			},
		},
		{
			name: "unknown command",
			test: func(t *testing.T, c *imap.Client) {
				mustDial(t, c)
				r, err := c.Send("foo")
				require.NoError(t, err)
				require.Equal(t, "A0001 BAD Unknown command FOO", r[0])
			},
		},
		{
			name: "capability",
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				caps, err := c.Capability()
				require.NoError(t, err)
				require.Equal(t, []string{"IMAP4rev1", "SASL-IR", "AUTH=PLAIN"}, caps)
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

			c := imap.NewClient(fmt.Sprintf("localhost:%v", p))

			tc.test(t, c)
		})
	}
}

func mustDial(t *testing.T, c *imap.Client) {
	_, err := c.Dial()
	require.NoError(t, err)
}
