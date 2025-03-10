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
		test    func(t *testing.T, c *imap.Client)
	}{
		{
			name: "empty",
			handler: &imaptest.Handler{
				ListFunc: func(ref, pattern string, _ []imap.MailboxFlags, session map[string]interface{}) ([]imap.ListEntry, error) {
					return nil, nil
				},
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)

				list, err := c.List("", "*")
				require.NoError(t, err)
				require.Len(t, list, 0)
			},
		},
		{
			name: "one found",
			handler: &imaptest.Handler{
				ListFunc: func(ref, pattern string, _ []imap.MailboxFlags, session map[string]interface{}) ([]imap.ListEntry, error) {
					return []imap.ListEntry{
						{
							Flags:     nil,
							Delimiter: "/",
							Name:      "foo",
						},
					}, nil
				},
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)

				list, err := c.List("", "*")
				require.NoError(t, err)
				require.Len(t, list, 1)
				require.Equal(t, "foo", list[0].Name)
			},
		},
		{
			name:    "not authenticated",
			handler: &imaptest.Handler{},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				lines, err := c.Send("LIST \"\" \"*\"")
				require.Equal(t, "A0001 BAD Command is only valid in authenticated state", lines[0])
			},
		},
		{
			name: "one with flags",
			handler: &imaptest.Handler{
				ListFunc: func(ref, pattern string, _ []imap.MailboxFlags, session map[string]interface{}) ([]imap.ListEntry, error) {
					return []imap.ListEntry{
						{
							Flags:     []imap.MailboxFlags{imap.HasNoChildren},
							Delimiter: "/",
							Name:      "INBOX",
						},
					}, nil
				},
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)

				list, err := c.List("", "*")
				require.NoError(t, err)
				require.Len(t, list, 1)
				require.Equal(t, "INBOX", list[0].Name)
				require.Equal(t, imap.HasNoChildren, list[0].Flags[0])
			},
		},
		{
			name: "LSub empty",
			handler: &imaptest.Handler{
				ListFunc: func(ref, pattern string, _ []imap.MailboxFlags, session map[string]interface{}) ([]imap.ListEntry, error) {
					return nil, nil
				},
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)

				list, err := c.LSub("", "*")
				require.NoError(t, err)
				require.Len(t, list, 0)
			},
		},
		{
			name: "LSub returns one",
			handler: &imaptest.Handler{
				ListFunc: func(ref, pattern string, _ []imap.MailboxFlags, session map[string]interface{}) ([]imap.ListEntry, error) {
					return []imap.ListEntry{
						{
							Flags:     []imap.MailboxFlags{imap.Subscribed, imap.HasNoChildren},
							Delimiter: "/",
							Name:      "INBOX",
						},
					}, nil
				},
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)

				list, err := c.LSub("", "*")
				require.NoError(t, err)
				require.Len(t, list, 1)
				require.Equal(t, "INBOX", list[0].Name)
				require.Equal(t, []imap.MailboxFlags{imap.Subscribed, imap.HasNoChildren}, list[0].Flags)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := try.GetFreePort()
			s := &imap.Server{
				Addr:    fmt.Sprintf(":%v", p),
				Handler: tc.handler,
			}
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
