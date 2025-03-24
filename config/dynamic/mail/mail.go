package mail

import (
	"mokapi/imap"
	"mokapi/smtp"
	"slices"
)

type Mail struct {
	*smtp.Message
	UId   uint32
	Flags []imap.Flag
}

func (m *Mail) HasFlag(flag imap.Flag) bool {
	for _, f := range m.Flags {
		if f == flag {
			return true
		}
	}
	return false
}

func (m *Mail) RemoveFlag(flag imap.Flag) {
	m.Flags = slices.DeleteFunc(m.Flags, func(f imap.Flag) bool {
		return f == flag
	})
}
