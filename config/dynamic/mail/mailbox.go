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
	Folders  map[string]*Folder

	nextUidValidity uint32
	m               sync.Mutex
}

type Folder struct {
	Name       string
	Flags      []imap.Flag
	Messages   []*Mail
	Subscribed bool

	// next available UID for new messages
	uidNext uint32
	// UIDVALIDITY is a per-folder identifier assigned by the server when the folder (mailbox) is created.
	// It helps IMAP clients determine whether previously stored UIDs are still valid.
	// If UIDVALIDITY changes, it means that all existing UIDs in that folder are no longer valid, and the client must discard any cached UIDs.
	uidValidity uint32
}

func (mb *Mailbox) Append(m *smtp.Message) {
	mb.m.Lock()
	defer mb.m.Unlock()

	mb.EnsureInbox()
	f := mb.Folders["INBOX"]

	if len(f.Messages) == mailboxSize {
		f.Messages = f.Messages[0 : len(f.Messages)-1]
	}
	uid := f.uidNext
	f.uidNext++
	f.Messages = append(f.Messages, &Mail{
		Message: m,
		UId:     uid,
		SeqNum:  uid,
	})
}

func (mb *Mailbox) EnsureInbox() {
	if mb.Folders == nil {
		mb.Folders = make(map[string]*Folder)
	}
	if _, ok := mb.Folders["INBOX"]; !ok {
		f := &Folder{Name: "INBOX", Subscribed: true}
		f.uidValidity = mb.nextUidValidity
		mb.nextUidValidity++

		mb.Folders["INBOX"] = f
	}
}

func (f *Folder) NumRecent() int {
	c := 0
	for _, m := range f.Messages {
		if m.HasFlag(imap.FlagRecent) {
			c += 1
		}
	}
	return c
}

func (f *Folder) FirstUnseen() *Mail {
	for _, m := range f.Messages {
		if !m.HasFlag(imap.FlagSeen) {
			return m
		}
	}
	return nil
}
