package imap_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"testing"
)

func _TestClient(t *testing.T) {
	c := imap.NewClient("imap.gmx.net:143")
	caps, err := c.Dial()
	require.NoError(t, err)
	require.Equal(t, []string{"IMAP4rev1", "CHILDREN", "ENABLE", "ID", "IDLE", "LIST-EXTENDED", "LIST-STATUS", "LITERAL-", "MOVE", "NAMESPACE", "SASL-IR", "SORT", "SPECIAL-USE", "THREAD=ORDEREDSUBJECT", "UIDPLUS", "UNSELECT", "WITHIN", "STARTTLS", "LOGINDISABLED"}, caps)
	require.NoError(t, err)
	err = c.StartTLS()
	require.NoError(t, err)
	err = c.PlainAuth("", "sarah.lehmann1@gmx.ch", "baw8cwq0gmu-AXQ9hnp")
	require.NoError(t, err)

	list, err := c.List("", "*")
	require.NoError(t, err)
	require.Contains(t, list, imap.ListEntry{Flags: []imap.MailboxFlags{imap.HasNoChildren}, Delimiter: "/", Name: "INBOX"})
	require.Contains(t, list, imap.ListEntry{Flags: []imap.MailboxFlags{imap.Trash, imap.HasNoChildren}, Delimiter: "/", Name: "Gelöscht"})

	selected, err := c.Select("INBOX", false)
	require.NoError(t, err)
	require.Equal(t, []imap.Flag{imap.FlagAnswered, imap.FlagFlagged, imap.FlagDeleted, imap.FlagSeen, imap.FlagDraft}, selected.Flags)
	require.Equal(t, uint32(1), selected.NumMessages)
	require.Equal(t, uint32(0), selected.NumRecent)
	require.Equal(t, uint32(1), selected.FirstUnseen)
	require.Equal(t, uint32(0x67bf6b38), selected.UIDValidity)
	require.Equal(t, uint32(0x2), selected.UIDNext)

	//require.Equal(t, "* 726 EXISTS", lines[2])
	//require.Equal(t, "* 0 RECENT", lines[3])
	//require.Equal(t, "* OK [UNSEEN 745] First unseen.", lines[4])
	//require.Equal(t, "* OK [UIDVALIDITY 1619957420] UIDs valid", lines[5])
	//require.Equal(t, "* OK [UIDNEXT 6399] Predicted next UID", lines[6])
	//require.Equal(t, "* OK [HIGHESTMODSEQ 13602] Highest", lines[7])
	//require.Equal(t, "A3 OK [READ-WRITE] Select completed (0.001 + 0.000 secs).", lines[8])

	//lines, err = c.Send("LSUB \"\" *")
	//require.Equal(t, "", lines)

	lines, err := c.Send("FETCH 1 (INTERNALDATE UID RFC822.SIZE FLAGS BODY.PEEK[HEADER.FIELDS (date subject from to cc message-id in-reply-to references content-type x-priority x-uniform-type-identifier x-universally-unique-identifier list-id list-unsubscribe bimi-indicator bimi-location x-bimi-indicator-hash authentication-results dkim-signature)])")
	require.NoError(t, err)
	_ = lines
	//lines, err = c.Send("FETCH 1 (FLAGS BODYSTRUCTURE)")
	cmd, err := c.Fetch(num(1), imap.FetchOptions{
		Flags: true,
		Body: []imap.FetchBodySection{
			{
				Specifier: "HEADER",
				Peek:      true,
				Fields:    []string{"date", "from", "to"},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, cmd)

}
