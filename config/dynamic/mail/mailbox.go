package mail

import (
	"mokapi/imap"
	"mokapi/smtp"
	"sync"
)

type Mailbox struct {
	Name     string
	Username string
	Password string
	Messages []*Mail

	messageSequenceNumber uint32
	uidValidity           uint32
	m                     sync.Mutex
}

func (mb *Mailbox) Append(m *smtp.Message) {
	mb.m.Lock()
	defer mb.m.Unlock()

	if len(mb.Messages) == mailboxSize {
		mb.Messages = mb.Messages[0 : len(mb.Messages)-1]
	}
	uid := mb.messageSequenceNumber
	mb.messageSequenceNumber++
	mb.Messages = append(mb.Messages, &Mail{
		Message: m,
		UId:     uid,
		SeqNum:  uid,
	})
}

func (mb *Mailbox) NumRecent() int {
	c := 0
	for _, m := range mb.Messages {
		if m.HasFlag(imap.FlagRecent) {
			c += 1
		}
	}
	return c
}

func (mb *Mailbox) FirstUnseen() *Mail {
	for _, m := range mb.Messages {
		if !m.HasFlag(imap.FlagSeen) {
			return m
		}
	}
	return nil
}
