package mail

import (
	"mokapi/imap"
	"mokapi/smtp"
)

type Mail struct {
	*smtp.Message
	UId    uint32
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
