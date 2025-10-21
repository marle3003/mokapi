package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
)

func TestAuth(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, c *imap.Client)
	}{
		{
			name: "not authenticated",
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.List("", "")
				require.EqualError(t, err, "imap status [BAD]: Command is only valid in authenticated state")
			},
		},
		{
			name: "login",
			test: func(t *testing.T, c *imap.Client) {
				res, err := c.Send("LOGIN username password")
				require.NoError(t, err)
				require.Equal(t, []string{"A0001 OK [IMAP4rev1 IDLE MOVE UIDPLUS UNSELECT] Logged in"}, res)
				_, err = c.List("", "")
				require.NoError(t, err)
			},
		},
		{
			name: "plain auth",
			test: func(t *testing.T, c *imap.Client) {
				err := c.PlainAuth("", "", "")
				require.NoError(t, err)
				_, err = c.List("", "")
				require.NoError(t, err)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := try.GetFreePort()
			s := &imap.Server{
				Addr: fmt.Sprintf(":%v", p),
				Handler: &imaptest.Handler{
					ListFunc: func(ref, pattern string, flags []imap.MailboxFlags, session map[string]interface{}) ([]imap.ListEntry, error) {
						return nil, nil
					},
				},
			}
			defer s.Close()
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imap.NewClient(fmt.Sprintf("localhost:%v", p))
			defer func() { _ = c.Close() }()

			_, err := c.Dial()
			require.NoError(t, err)

			tc.test(t, c)
		})
	}
}
