package mail

import (
	"fmt"
	"maps"
	"mokapi/smtp"
	"slices"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// DefaultMaxInboxMails is the documented INBOX limit when a config does not set maxInboxMails.
const DefaultMaxInboxMails = 100

// Store keeps the mailboxes of one mail service. It outlives config reloads: Update changes the
// mailboxes to the new config without losing their mails. The mailboxes are only reachable
// through the methods of the store, which guard the map with its own lock.
type Store struct {
	Name string

	m         sync.RWMutex
	mailboxes map[string]*Mailbox
}

func NewStore(c *Config) *Store {
	s := &Store{
		Name:      c.Info.Name,
		mailboxes: map[string]*Mailbox{},
	}
	s.Update(c)
	return s
}

// Update brings the mailboxes in line with the config: new mailboxes are created, existing ones
// keep their mails and folders, configured mailboxes that are no longer in the config are
// removed. Mailboxes created at runtime by EnsureMailbox stay.
func (s *Store) Update(c *Config) {
	s.m.Lock()
	defer s.m.Unlock()

	limit := maxInboxMails(c.Settings)
	for name, cfg := range c.Mailboxes {
		if cfg == nil {
			cfg = &MailboxConfig{}
		}
		if mb, ok := s.mailboxes[name]; ok {
			mb.update(cfg, limit)
		} else {
			s.mailboxes[name] = newMailbox(name, cfg, limit)
		}
	}

	for name, mb := range s.mailboxes {
		if _, ok := c.Mailboxes[name]; !ok && !mb.autoCreated {
			delete(s.mailboxes, name)
		}
	}
}

func (s *Store) Mailbox(name string) (*Mailbox, bool) {
	if s == nil {
		return nil, false
	}
	s.m.RLock()
	defer s.m.RUnlock()
	mb, ok := s.mailboxes[name]
	return mb, ok
}

// Mailboxes returns a copy of the list, sorted by name.
func (s *Store) Mailboxes() []*Mailbox {
	if s == nil {
		return nil
	}
	s.m.RLock()
	defer s.m.RUnlock()
	list := slices.Collect(maps.Values(s.mailboxes))
	slices.SortFunc(list, func(a, b *Mailbox) int { return strings.Compare(a.Name, b.Name) })
	return list
}

// EnsureMailbox creates the mailbox if it does not exist, unless the settings forbid it.
func (s *Store) EnsureMailbox(name string, settings *Settings) error {
	s.m.Lock()
	if _, found := s.mailboxes[name]; found {
		s.m.Unlock()
		return nil
	}
	if settings != nil && !settings.AutoCreateMailbox {
		s.m.Unlock()
		return fmt.Errorf("mailbox can not be created")
	}
	mb := newMailbox(name, &MailboxConfig{}, maxInboxMails(settings))
	mb.autoCreated = true
	s.mailboxes[name] = mb
	s.m.Unlock()

	log.Infof("adding new mailbox '%v' to '%v'", name, s.Name)
	return nil
}

func (s *Store) GetMail(id string) *smtp.Message {
	for _, mb := range s.Mailboxes() {
		if m := mb.message(id); m != nil {
			return m
		}
	}
	return nil
}

func maxInboxMails(settings *Settings) int {
	if settings == nil {
		return DefaultMaxInboxMails
	}
	return settings.MaxInboxMails
}

func newMailbox(name string, cfg *MailboxConfig, maxInboxMails int) *Mailbox {
	mb := &Mailbox{
		Name:            name,
		Folders:         map[string]*Folder{},
		username:        cfg.Username,
		password:        cfg.Password,
		description:     cfg.Description,
		maxInboxMails:   maxInboxMails,
		nextUidValidity: uint32(time.Now().Unix()),
	}
	mb.mergeFolders(mb.Folders, cfg.Folders, true)
	return mb
}
