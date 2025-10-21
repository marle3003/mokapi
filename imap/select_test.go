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
		test    func(t *testing.T, c *imap.Client)
	}{
		{
			name: "select inbox raw",
			handler: &imaptest.Handler{
				SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
					require.False(t, readonly)
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
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				res, err := c.Send("SELECT INBOX")
				require.NoError(t, err)
				require.Equal(t, []string{
					"* 172 EXISTS",
					"* 1 RECENT",
					"* OK [UNSEEN 12] Message 12 is first unseen",
					"* OK [UIDVALIDITY 3857529045] UIDs valid",
					"* OK [UIDNEXT 4392] Predicted next UID",
					"* FLAGS (\\Answered \\Flagged \\Deleted \\Seen \\Draft)",
					"A0002 OK [READ-WRITE] SELECT completed",
				}, res)
			},
		},
		{
			name: "examine inbox raw",
			handler: &imaptest.Handler{
				SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
					require.True(t, readonly)
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
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				res, err := c.Send("EXAMINE INBOX")
				require.NoError(t, err)
				require.Equal(t, []string{
					"* 172 EXISTS",
					"* 1 RECENT",
					"* OK [UNSEEN 12] Message 12 is first unseen",
					"* OK [UIDVALIDITY 3857529045] UIDs valid",
					"* OK [UIDNEXT 4392] Predicted next UID",
					"* FLAGS (\\Answered \\Flagged \\Deleted \\Seen \\Draft)",
					"A0002 OK [READ-ONLY] SELECT completed",
				}, res)
			},
		},
		{
			name: "select inbox",
			handler: &imaptest.Handler{
				SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
					require.False(t, readonly)
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
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				selected, err := c.Select("INBOX", false)
				require.NoError(t, err)
				require.Equal(t, uint32(172), selected.NumMessages)
				require.Equal(t, uint32(1), selected.NumRecent)
				require.Equal(t, []imap.Flag{imap.FlagAnswered, imap.FlagFlagged, imap.FlagDeleted, imap.FlagSeen, imap.FlagDraft}, selected.Flags)
				require.Equal(t, uint32(12), selected.FirstUnseen)
				require.Equal(t, uint32(3857529045), selected.UIDValidity)
				require.Equal(t, uint32(4392), selected.UIDNext)
			},
		},
		{
			name: "examine inbox",
			handler: &imaptest.Handler{
				SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
					require.True(t, readonly)
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
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				selected, err := c.Select("INBOX", true)
				require.NoError(t, err)
				require.Equal(t, uint32(172), selected.NumMessages)
				require.Equal(t, uint32(1), selected.NumRecent)
				require.Equal(t, []imap.Flag{imap.FlagAnswered, imap.FlagFlagged, imap.FlagDeleted, imap.FlagSeen, imap.FlagDraft}, selected.Flags)
				require.Equal(t, uint32(12), selected.FirstUnseen)
				require.Equal(t, uint32(3857529045), selected.UIDValidity)
				require.Equal(t, uint32(4392), selected.UIDNext)
			},
		},
		{
			name: "close selected mailbox",
			handler: &imaptest.Handler{
				SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
					return &imap.Selected{
						NumMessages: 172,
						NumRecent:   1,
						FirstUnseen: 12,
						UIDValidity: 3857529045,
						UIDNext:     4392,
						Flags:       []imap.Flag{imap.FlagAnswered, imap.FlagFlagged, imap.FlagDeleted, imap.FlagSeen, imap.FlagDraft},
					}, nil
				},
				UnselectFunc: func(session map[string]interface{}) error {
					return nil
				},
				ExpungeFunc: func(set *imap.IdSet, w imap.ExpungeWriter, session map[string]interface{}) error {
					return nil
				},
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				_, err = c.Select("INBOX", false)
				require.NoError(t, err)
				err = c.Close()
				require.NoError(t, err)
			},
		},
		{
			name: "not authenticated",
			handler: &imaptest.Handler{
				SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
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
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				_, err = c.Select("INBOX", false)
				require.EqualError(t, err, "imap status [BAD]: Command is only valid in authenticated state")
			},
		},
		{
			name: "no such mailbox",
			handler: &imaptest.Handler{
				SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
					return nil, fmt.Errorf("no mailbox")
				},
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				_, err = c.Select("INBOX", false)
				require.EqualError(t, err, "imap status [NO]: No such mailbox, can't access mailbox")
			},
		},
		{
			name: "unselect before select another mailbox",
			handler: &imaptest.Handler{
				SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
					if _, found := session["mailbox"]; found {
						panic("mailbox not unselected")
					}
					session["mailbox"] = mailbox
					return &imap.Selected{}, nil
				},
				UnselectFunc: func(session map[string]interface{}) error {
					delete(session, "mailbox")
					return nil
				},
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				_, err = c.Select("INBOX", false)
				require.NoError(t, err)
				_, err = c.Select("FOO", false)
				require.NoError(t, err)
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
