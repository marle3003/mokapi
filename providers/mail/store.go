package mail

import (
	"fmt"
	"mokapi/imap"
	"mokapi/smtp"
	"strings"
	"time"
)

const mailboxSize = 100

type Store struct {
	Mailboxes map[string]*Mailbox
	Settings  *Settings
}

func NewStore(c *Config) *Store {
	s := &Store{
		Mailboxes: map[string]*Mailbox{},
		Settings:  c.Settings,
	}

	for name, mb := range c.Mailboxes {
		s.NewMailbox(name, mb, c.Settings)
	}

	return s
}

func (s *Store) Update(c *Config) {
	for name, mb := range c.Mailboxes {
		if exist, ok := s.Mailboxes[name]; !ok {
			s.NewMailbox(name, mb, c.Settings)
		} else {
			exist.Username = mb.Username
			exist.Password = mb.Password
			folders := getFolders(mb.Folders)
			for _, folder := range folders {
				exist.Folders[folder.Name] = folder
			}
		}
	}
}

func (s *Store) ExistsMailbox(name string) bool {
	_, b := s.Mailboxes[name]
	return b
}

func (s *Store) NewMailbox(name string, cfg *MailboxConfig, settings *Settings) {
	if _, found := s.Mailboxes[name]; found {
		return
	}

	maxInboxMails := mailboxSize
	if settings != nil {
		maxInboxMails = settings.MaxInboxMails
	}

	mb := &Mailbox{
		Username:        cfg.Username,
		Password:        cfg.Password,
		Description:     cfg.Description,
		MaxInboxMails:   maxInboxMails,
		nextUidValidity: uint32(time.Now().Unix()),
	}
	mb.Folders = getFolders(cfg.Folders)

	s.Mailboxes[name] = mb
}

func (s *Store) EnsureMailbox(name string) error {
	if _, found := s.Mailboxes[name]; found {
		return nil
	}
	if s.Settings != nil && !s.Settings.AutoCreateMailbox {
		return fmt.Errorf("mailbox can not be created")
	}
	s.NewMailbox(name, &MailboxConfig{}, s.Settings)
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

func getFolders(cfg map[string]*FolderConfig) map[string]*Folder {
	result := make(map[string]*Folder)
	for name, sub := range cfg {
		if strings.ToUpper(name) == "INBOX" {
			name = "INBOX"
		}

		f := &Folder{
			Name:        name,
			uidNext:     1,
			uidValidity: uint32(time.Now().Unix()),
		}

		for _, flag := range sub.Flags {
			f.Flags = append(f.Flags, imap.MailboxFlags(flag))
		}

		f.Folders = getFolders(sub.Folders)
		result[name] = f
	}

	return result
}
