package mail

import (
	"mokapi/imap"
	"mokapi/smtp"
)

type Mail struct {
	*smtp.Message
	UId uint32

	// A Message Sequence Number (MSN) is a temporary, per-session identifier for messages within an IMAP folder
	// (mailbox). It starts from 1 for the first message and increases sequentially for each message in the folder.
	// Unlike UIDs, which are permanent and unique per folder, Message Sequence Numbers can change between sessions
	// as messages are added, removed, or expunged.
	SeqNum uint32
	Flags  []imap.Flag
}

func (m *Mail) HasFlag(flag imap.Flag) bool {
	for _, f := range m.Flags {
		if f == flag {
			return true
		}
	}
	return false
}
