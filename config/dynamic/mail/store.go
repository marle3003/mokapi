package mail

import (
	"fmt"
	"mokapi/imap"
	"mokapi/smtp"
	"time"
)

const mailboxSize = 100

type Store struct {
	Mailboxes map[string]*Mailbox

	canAddMailbox bool
}

func NewStore(c *Config) *Store {
	s := &Store{
		Mailboxes:     map[string]*Mailbox{},
		canAddMailbox: len(c.Mailboxes) == 0,
	}
	for _, mb := range c.Mailboxes {
		s.NewMailbox(&mb)
	}

	return s
}

func (s *Store) Update(c *Config) {
	for _, mb := range c.Mailboxes {
		if _, ok := s.Mailboxes[mb.Name]; !ok {
			s.NewMailbox(&mb)
		}
	}
}

func (s *Store) ExistsMailbox(name string) bool {
	_, b := s.Mailboxes[name]
	return b
}

func (s *Store) NewMailbox(cfg *MailboxConfig) {
	if _, found := s.Mailboxes[cfg.Name]; found {
		return
	}

	mb := &Mailbox{
		Name:            cfg.Name,
		Username:        cfg.Username,
		Password:        cfg.Password,
		nextUidValidity: uint32(time.Now().Unix()),
	}
	mb.Folders = getFolders(cfg.Folders)

	s.Mailboxes[cfg.Name] = mb
}

func (s *Store) EnsureMailbox(name string) error {
	if _, found := s.Mailboxes[name]; found {
		return nil
	}
	if !s.canAddMailbox {
		return fmt.Errorf("mailbox can not be created")
	}
	s.NewMailbox(&MailboxConfig{Name: name})
	return nil
}

func (s *Store) GetMail(id string) *smtp.Message {
	for _, b := range s.Mailboxes {
		for _, f := range b.Folders {
			for _, m := range f.Messages {
				if m.MessageId == id {
					return m.Message
				}
			}
		}
	}
	return nil
}

func getFolders(cfg []FolderConfig) map[string]*Folder {
	result := make(map[string]*Folder)
	for _, sub := range cfg {
		f := &Folder{
			Name:        sub.Name,
			uidNext:     1,
			uidValidity: uint32(time.Now().Unix()),
		}

		for _, flag := range sub.Flags {
			f.Flags = append(f.Flags, imap.MailboxFlags(flag))
		}

		f.Folders = getFolders(sub.Folders)
		result[sub.Name] = f
	}

	return result
}
