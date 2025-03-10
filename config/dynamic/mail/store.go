package mail

import (
	"fmt"
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
		s.NewMailbox(mb.Name, mb.Username, mb.Password)
	}

	return s
}

func (s *Store) Update(c *Config) {
	for _, mb := range c.Mailboxes {
		if _, ok := s.Mailboxes[mb.Name]; !ok {
			s.NewMailbox(mb.Name, mb.Username, mb.Password)
		}
	}
}

func (s *Store) ExistsMailbox(name string) bool {
	_, b := s.Mailboxes[name]
	return b
}

func (s *Store) NewMailbox(name, username, password string) {
	if _, found := s.Mailboxes[name]; found {
		return
	}
	s.Mailboxes[name] = &Mailbox{
		Name:            name,
		Username:        username,
		Password:        password,
		nextUidValidity: uint32(time.Now().Unix()),
	}
}

func (s *Store) EnsureMailbox(name string) error {
	if _, found := s.Mailboxes[name]; found {
		return nil
	}
	if !s.canAddMailbox {
		return fmt.Errorf("mailbox can not be created")
	}
	s.NewMailbox(name, "", "")
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
