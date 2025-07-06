package mail

import (
	"fmt"
	"mokapi/imap"
	"mokapi/smtp"
	"slices"
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

	mb *Mailbox

	// next available UID for new messages
	uidNext uint32
	// UIDVALIDITY is a per-folder identifier assigned by the server when the folder (mailbox) is created.
	// It helps IMAP clients determine whether previously stored UIDs are still valid.
	// If UIDVALIDITY changes, it means that all existing UIDs in that folder are no longer valid, and the client must discard any cached UIDs.
	uidValidity uint32

	recentUid uint32
}

func (mb *Mailbox) Append(m *smtp.Message) {
	mb.EnsureInbox()
	f := mb.Folders["INBOX"]
	f.Append(m)
}

// Offset to start UID from year 2000 instead of 1970 (Unix epoch)
const epochOffset = 1740937638

func (mb *Mailbox) AddFolder(child *Folder) {
	if mb.Folders == nil {
		mb.Folders = make(map[string]*Folder)
	}

	child.mb = mb
	child.uidNext = uint32(time.Now().Unix() - epochOffset)
	child.uidValidity = mb.getNextUidValidity()

	mb.Folders[child.Name] = child
}

func (mb *Mailbox) ensurePath(path string) *Folder {
	if mb.Folders == nil {
		mb.Folders = make(map[string]*Folder)
	}

	parts := strings.Split(path, "/")
	var current *Folder
	for _, part := range parts {
		var next *Folder
		if current == nil {
			next = mb.Folders[part]
		} else {
			next = current.Folders[part]
		}
		if next == nil {
			next = &Folder{
				Name:        part,
				uidNext:     uint32(time.Now().Unix() - epochOffset),
				uidValidity: mb.getNextUidValidity(),
				mb:          mb,
			}
			if current == nil {
				mb.Folders[part] = next
			} else {
				if current.Folders == nil {
					current.Folders = make(map[string]*Folder)
				}
				current.Folders[part] = next
			}
		}
		current = next
	}
	return current
}

func (mb *Mailbox) Select(path string) *Folder {
	parts := strings.Split(path, "/")
	var current *Folder
	for _, part := range parts {
		// Inbox is a special, mandatory mailbox that is case-insensitive
		if strings.ToUpper(part) == "INBOX" {
			current = mb.Folders["INBOX"]
			continue
		}

		if current == nil {
			current = mb.Folders[part]
		} else {
			current = current.Folders[part]
		}
		if current == nil {
			return nil
		}
	}
	return current
}

func (mb *Mailbox) DeleteFolder(path string) error {
	if strings.ToUpper(path) == "INBOX" {
		return fmt.Errorf("INBOX cannot be deleted")
	}

	parts := strings.Split(path, "/")
	var parent *Folder
	for _, part := range parts[:len(parts)-1] {
		if parent == nil {
			parent = mb.Folders[part]
		} else {
			parent = parent.Folders[part]
		}
		if parent == nil {
			return fmt.Errorf("mailbox \"%s\" not found", part)
		}
	}
	name := parts[len(parts)-1]
	if parent != nil {
		return deleteFolder(parent.Folders, name)
	}
	return deleteFolder(mb.Folders, name)
}

func deleteFolder(folders map[string]*Folder, name string) error {
	if toDelete, ok := folders[name]; !ok {
		return fmt.Errorf("mailbox \"%s\" not found", name)
	} else if len(toDelete.Folders) > 0 {
		return fmt.Errorf("name \"%s\" has inferior hierarchical names", name)
	} else {
		delete(folders, name)
	}
	return nil
}

func (mb *Mailbox) RenameFolder(existingName, newName string) error {
	if strings.ToUpper(existingName) == "INBOX" {
		f := mb.ensurePath(newName)
		inbox := mb.Folders["INBOX"]
		f.Messages = inbox.Messages
		inbox.Messages = nil
		return nil
	}

	parts := strings.Split(existingName, "/")
	folders := mb.Folders
	for _, part := range parts[:len(parts)-1] {
		folder := folders[part]
		if folder == nil {
			return fmt.Errorf("mailbox \"%s\" not found", part)
		}
		folders = folders[part].Folders
	}
	name := parts[len(parts)-1]
	toRename, ok := folders[name]
	if !ok {
		return fmt.Errorf("mailbox \"%s\" not found", name)
	}

	parts = strings.Split(newName, "/")
	folders2 := mb.Folders
	for _, part := range parts[:len(parts)-1] {
		folder := folders2[part]
		if folder == nil {
			return fmt.Errorf("mailbox \"%s\" not found", part)
		}
		if folders[part].Folders == nil {
			folders[part].Folders = make(map[string]*Folder)
		}
		folders2 = folders[part].Folders
	}

	name2 := parts[len(parts)-1]
	if _, ok := folders2[name2]; ok {
		return fmt.Errorf("mailbox \"%s\" already exists", newName)
	}

	folders2[name2] = toRename
	toRename.Name = name2
	delete(folders, name)

	return nil
}

func (mb *Mailbox) getNextUidValidity() uint32 {
	uidValidity := mb.nextUidValidity
	mb.nextUidValidity++
	return uidValidity
}

func (mb *Mailbox) EnsureInbox() {
	mb.m.Lock()
	_, ok := mb.Folders["INBOX"]
	mb.m.Unlock()
	if !ok {
		mb.AddFolder(&Folder{Name: "INBOX"})
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
				result = append(result, child.List(strings.Join(parts[1:], "/"), nil)...)
			} else {
				for _, sub := range child.Folders {
					result = append(result, sub)
				}
			}
		}
	}

	return result
}

func (f *Folder) UidValidity() uint32 {
	if f.uidValidity == 0 {
		f.uidValidity = uint32(time.Now().Unix() - epochOffset)
	}
	return f.uidValidity
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
		Flags:   []imap.Flag{imap.FlagRecent},
	})
}

func (f *Folder) Copy(m *Mail) *Mail {
	c := *m
	uid := f.uidNext
	f.uidNext++
	c.UId = uid
	f.Messages = append(f.Messages, &c)
	return &c
}

func (f *Folder) Remove(m *Mail) {
	var result []*Mail
	for _, v := range f.Messages {
		if v.UId != m.UId {
			result = append(result, v)
		}
	}
	f.Messages = result
}

func (f *Folder) Status() imap.StatusResult {
	result := imap.StatusResult{
		UIDNext:     f.uidNext,
		UIDValidity: f.uidValidity,
	}
	for _, m := range f.Messages {
		result.Messages++
		if m.HasFlag(imap.FlagRecent) {
			result.Recent++
		}
		if !m.HasFlag(imap.FlagSeen) {
			result.Unseen++
		}
	}
	return result
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

func (f *Folder) AddFolder(child *Folder) {
	if f.Folders == nil {
		f.Folders = make(map[string]*Folder)
	}

	child.uidNext = uint32(time.Now().Unix() - epochOffset)
	child.uidValidity = f.mb.getNextUidValidity()
	child.mb = f.mb

	f.Folders[child.Name] = child
}

func (f *Folder) List(pattern string, flags []imap.MailboxFlags) []*Folder {
	if pattern == "" {
		return nil
	}
	parts := strings.Split(pattern, "/")
	var result []*Folder

	if len(parts) == 1 {
		switch parts[0] {
		case "*":
			if f.HasFlags(flags...) {
				result = append(result, f)
			}
			for _, child := range f.Folders {
				if !child.HasFlags(flags...) {
					continue
				}
				result = append(result, child.List("*", flags)...)
			}
		case "%":
			if f.HasFlags(flags...) {
				result = append(result, f)
			}
		default:
			if parts[0] == f.Name && f.HasFlags(flags...) {
				result = append(result, f)
			}
		}
	} else {
		if parts[0] != f.Name {
			return nil
		}
		for _, child := range f.Folders {
			result = append(result, child.List(strings.Join(parts[1:], "/"), flags)...)
		}
	}

	return result
}

func (f *Folder) RemoveFlag(flag imap.MailboxFlags) {
	f.Flags = slices.DeleteFunc(f.Flags, func(f imap.MailboxFlags) bool {
		return f == flag
	})
}

func (f *Folder) HasFlags(flags ...imap.MailboxFlags) bool {
	flagSet := make(map[imap.MailboxFlags]struct{}, len(f.Flags))
	for _, flag := range f.Flags {
		flagSet[flag] = struct{}{}
	}

	for _, flag := range flags {
		if _, exists := flagSet[flag]; !exists {
			return false
		}
	}
	return true
}
