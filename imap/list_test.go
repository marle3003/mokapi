package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
)

func TestServer_List(t *testing.T) {
	testcases := []struct {
		name    string
		handler imap.Handler
		test    func(t *testing.T, c *imaptest.Client)
	}{
		{
			name: "empty",
			handler: &imaptest.Handler{
				ListFunc: func(ref, pattern string, session map[string]interface{}) ([]imap.ListEntry, error) {
					return nil, nil
				},
			},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				lines, err := c.Send("LIST \"\" \"*\"")
				require.Equal(t, "A2 OK List completed", lines[0])
			},
		},
		{
			name: "one found",
			handler: &imaptest.Handler{
				ListFunc: func(ref, pattern string, session map[string]interface{}) ([]imap.ListEntry, error) {
					return []imap.ListEntry{
						{
							Flags: nil,
							Name:  "foo",
						},
					}, nil
				},
			},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				lines, err := c.Send("LIST \"\" \"*\"")
				require.Equal(t, "* LIST () NIL foo", lines[0])
				require.Equal(t, "A2 OK List completed", lines[1])
			},
		},
		{
			name:    "not authenticated",
			handler: &imaptest.Handler{},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				lines, err := c.Send("LIST \"\" \"*\"")
				require.Equal(t, "A1 BAD Command is only valid in authenticated state", lines[0])
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
				Handler: tc.handler,
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
