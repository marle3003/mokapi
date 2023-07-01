package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
)

func TestServer_Select(t *testing.T) {
	testcases := []struct {
		name    string
		handler imap.Handler
		test    func(t *testing.T, c *imaptest.Client)
	}{
		{
			name: "select inbox",
			handler: &imaptest.Handler{
				SelectFunc: func(mailbox string) (*imap.Selected, error) {
					return &imap.Selected{
						NumMessages: 172,
						NumRecent:   1,
						FirstUnseen: 12,
						UIDValidity: 3857529045,
						UIDNext:     4392,
						Flags:       []imap.Flag{imap.FlagAnswered, imap.FlagFlagged, imap.FlagDeleted, imap.FlagSeen, imap.FlagDraft},
					}, nil
				},
			},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				lines, err := c.Send("SELECT INBOX")
				require.NoError(t, err)
				require.Equal(t, "* 172 EXISTS", lines[0])
				require.Equal(t, "* 1 RECENT", lines[1])
				require.Equal(t, "* OK [UNSEEN 12] Message 12 is first unseen", lines[2])
				require.Equal(t, "* OK [UIDVALIDITY 3857529045] UIDs valid", lines[3])
				require.Equal(t, "* OK [UIDNEXT 4392] Predicted next UID", lines[4])
				require.Equal(t, "* FLAGS (\\Answered \\Flagged \\Deleted \\Seen \\Draft)", lines[5])
				require.Equal(t, "A2 OK [READ-WRITE] SELECT completed", lines[6])
			},
		},
		{
			name: "not authenticated",
			handler: &imaptest.Handler{
				SelectFunc: func(mailbox string) (*imap.Selected, error) {
					return &imap.Selected{
						NumMessages: 172,
						NumRecent:   1,
						FirstUnseen: 12,
						UIDValidity: 3857529045,
						UIDNext:     4392,
						Flags:       []imap.Flag{imap.FlagAnswered, imap.FlagFlagged, imap.FlagDeleted, imap.FlagSeen, imap.FlagDraft},
					}, nil
				},
			},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				lines, err := c.Send("SELECT INBOX")
				require.NoError(t, err)
				require.Equal(t, "A1 BAD Command is only valid in authenticated state", lines[0])
			},
		},
		{
			name: "no such mailbox",
			handler: &imaptest.Handler{
				SelectFunc: func(mailbox string) (*imap.Selected, error) {
					return nil, fmt.Errorf("no mailbox")
				},
			},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				lines, err := c.Send("SELECT INBOX")
				require.NoError(t, err)
				require.Equal(t, "A2 NO No such mailbox, can't access mailbox", lines[0])
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
