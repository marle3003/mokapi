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
				SelectFunc: func(mailbox string) *imap.Selected {
					return &imap.Selected{
						NumMessages: 172,
						NumRecent:   1,
						FirstUnseen: 12,
						UIDValidity: 3857529045,
						UIDNext:     4392,
						Flags:       []imap.Flag{imap.FlagAnswered, imap.FlagFlagged, imap.FlagDeleted, imap.FlagSeen, imap.FlagDraft},
					}
				},
			},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				r, err := c.Send("SELECT INBOX")
				require.NoError(t, err)
				require.Equal(t, "* 172 EXISTS", r)

				r, err = c.ReadLine()
				require.NoError(t, err)
				require.Equal(t, "* 1 RECENT", r)

				r, err = c.ReadLine()
				require.NoError(t, err)
				require.Equal(t, "* OK [UNSEEN 12] Message 12 is first unseen", r)

				r, err = c.ReadLine()
				require.NoError(t, err)
				require.Equal(t, "* OK [UIDVALIDITY 3857529045] UIDs valid", r)

				r, err = c.ReadLine()
				require.NoError(t, err)
				require.Equal(t, "* OK [UIDNEXT 4392] Predicted next UID", r)

				r, err = c.ReadLine()
				require.NoError(t, err)
				require.Equal(t, "* FLAGS (\\Answered \\Flagged \\Deleted \\Seen \\Draft)", r)

				r, err = c.ReadLine()
				require.NoError(t, err)
				require.Equal(t, "A2 OK [READ-WRITE] SELECT completed", r)
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
