package mail

import (
	"mokapi/imap"
	"mokapi/smtp"
	"strings"
	"sync"
	"time"
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
	Flags      []imap.MailboxFlags
	Messages   []*Mail
	Subscribed bool
	Folders    map[string]*Folder

	// next available UID for new messages
	uidNext uint32
	// UIDVALIDITY is a per-folder identifier assigned by the server when the folder (mailbox) is created.
	// It helps IMAP clients determine whether previously stored UIDs are still valid.
	// If UIDVALIDITY changes, it means that all existing UIDs in that folder are no longer valid, and the client must discard any cached UIDs.
	uidValidity uint32

	recentUid uint32
}

func (mb *Mailbox) Append(m *smtp.Message) {
	mb.m.Lock()
	defer mb.m.Unlock()

	mb.EnsureInbox()
	f := mb.Folders["INBOX"]
	f.Append(m)
}

// Offset to start UID from year 2000 instead of 1970 (Unix epoch)
const epochOffset = 1740937638

func (mb *Mailbox) EnsureInbox() {
	if mb.Folders == nil {
		mb.Folders = make(map[string]*Folder)
	}
	if _, ok := mb.Folders["INBOX"]; !ok {
		f := &Folder{Name: "INBOX", Subscribed: true, uidNext: uint32(time.Now().Unix() - epochOffset)}
		f.uidValidity = mb.nextUidValidity
		mb.nextUidValidity++

		mb.Folders["INBOX"] = f
	}
}

func (mb *Mailbox) List(pattern string) []*Folder {
	var result []*Folder

	if pattern == "" {
		for _, child := range mb.Folders {
			result = append(result, child)
		}
		return result
	}
	parts := strings.Split(pattern, "/")

	for _, child := range mb.Folders {
		if parts[0] == child.Name {
			if len(parts) > 1 {
				result = append(result, child.List(strings.Join(parts[1:], "/"))...)
			} else {
				for _, sub := range child.Folders {
					result = append(result, sub)
				}
			}
		}
	}

	return result
}

func (f *Folder) Append(m *smtp.Message) {
	if len(f.Messages) == mailboxSize {
		f.Messages = f.Messages[0 : len(f.Messages)-1]
	}
	uid := f.uidNext
	f.uidNext++
	f.Messages = append(f.Messages, &Mail{
		Message: m,
		UId:     uid,
		SeqNum:  uid,
		Flags:   []imap.Flag{imap.FlagRecent},
	})
}

func (f *Folder) Copy(m *Mail) {
	if len(f.Messages) == mailboxSize {
		f.Messages = f.Messages[0 : len(f.Messages)-1]
	}
	uid := f.uidNext
	f.uidNext++
	m.UId = uid
	m.SeqNum = uid
	f.Messages = append(f.Messages, m)
}

func (f *Folder) NumRecent() int {
	c := 0
	for _, m := range f.Messages {
		if m.UId <= f.recentUid {
			m.RemoveFlag(imap.FlagRecent)
		} else if m.HasFlag(imap.FlagRecent) {
			c++
			f.recentUid = m.UId
		}
	}

	return c
}

func (f *Folder) FirstUnseen() int {
	for i, m := range f.Messages {
		if !m.HasFlag(imap.FlagSeen) {
			return i + 1
		}
	}
	return -1
}

func (f *Folder) List(pattern string) []*Folder {
	if pattern == "" {
		return nil
	}
	parts := strings.Split(pattern, "/")
	var result []*Folder

	if len(parts) == 1 {
		switch parts[0] {
		case "*":
			result = append(result, f)
			for _, child := range f.Folders {
				result = append(result, child.List("*")...)
			}
		case "%":
			result = append(result, f)
		default:
			if parts[0] == f.Name {
				result = append(result, f)
			}
		}
	} else {
		if parts[0] != f.Name {
			return nil
		}
		for _, child := range f.Folders {
			result = append(result, child.List(strings.Join(parts[1:], "/"))...)
		}
	}

	return result
}
