package mail

import (
	"fmt"
	"mokapi/smtp"
)

const mailboxSize = 100

type Store struct {
	Mailboxes map[string]*Mailbox

	canAddMailbox bool
}

type Mailbox struct {
	Name     string
	Username string
	Password string
	Messages []*smtp.Message
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

func (s *Store) ExistsMailbox(name string) bool {
	_, b := s.Mailboxes[name]
	return b
}

func (s *Store) NewMailbox(name, username, password string) {
	if _, found := s.Mailboxes[name]; found {
		return
	}
	s.Mailboxes[name] = &Mailbox{
		Name:     name,
		Username: username,
		Password: password,
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
		for _, m := range b.Messages {
			if m.MessageId == id {
				return m
			}
		}
	}
	return nil
}

func (mb *Mailbox) Append(m *smtp.Message) {
	if len(mb.Messages) == mailboxSize {
		mb.Messages = mb.Messages[0 : len(mb.Messages)-1]
	}
	mb.Messages = append(mb.Messages, m)
}
