package mail

import (
	"fmt"
	"mokapi/smtp"
	"sync"
	"time"
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
	Messages []*Mail

	messageSequenceNumber uint32
	uidValidity           uint32
	m                     sync.Mutex
}

type Mail struct {
	*smtp.Message
	UId uint32
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
		Name:                  name,
		Username:              username,
		Password:              password,
		messageSequenceNumber: 1,
		// max date is February 2106
		uidValidity: uint32(time.Now().Unix()),
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
				return m.Message
			}
		}
	}
	return nil
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
	})
}
