package runtime

import (
	"context"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	engine "mokapi/engine/common"
	"mokapi/imap"
	"mokapi/providers/mail"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/runtime/search"
	"mokapi/smtp"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

type MailHandler interface {
	smtp.Handler
	imap.Handler
}

// ReloadableMailHandler serves SMTP and IMAP requests and switches to the current config snapshot
// of its service on Reload, so running servers apply changed rules and settings.
type ReloadableMailHandler interface {
	MailHandler
	Reload()
}

type MailStore struct {
	infos map[string]*MailInfo
	m     sync.RWMutex
	cfg   *static.Config
	sm    *events.StoreManager
	index search.Index
}

// MailInfo publishes the merged config as an immutable snapshot, so a reader takes one snapshot
// via Config and uses only that. The Store holds the mailboxes and their mails. It is created
// once and follows each reload through Store.Update, so received mails survive.
type MailInfo struct {
	*mail.Store
	config  atomic.Pointer[mail.Config]
	m       sync.RWMutex // guards configs
	configs map[string]*dynamic.Config
}

func (s *MailStore) Get(name string) *MailInfo {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.infos[name]
}

func (s *MailStore) List() []*MailInfo {
	if s == nil {
		return nil
	}

	s.m.RLock()
	defer s.m.RUnlock()

	var list []*MailInfo
	for _, v := range s.infos {
		list = append(list, v)
	}
	return list
}

func (s *MailStore) Add(c *dynamic.Config) *MailInfo {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*MailInfo)
	}
	cfg := c.Data.(*mail.Config)
	name := cfg.Info.Name
	mi, ok := s.infos[name]

	store, hasStoreConfig := s.cfg.Event.Store[name]
	if !hasStoreConfig {
		store = s.cfg.Event.Store["default"]
	}

	if !ok {
		mi = NewMailInfo(c)
		s.infos[cfg.Info.Name] = mi

		s.sm.ResetStores(events.NewTraits().WithNamespace("mail").WithName(cfg.Info.Name))
		s.sm.SetStore(int(store.Size), events.NewTraits().WithNamespace("mail").WithName(cfg.Info.Name))
	} else {
		mi.addConfig(c)
	}

	if s.cfg.Api.Search.Enabled {
		s.addToIndex(mi.Config())
	}

	return mi
}

func (s *MailStore) Set(name string, mi *MailInfo) {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*MailInfo)
	}

	s.infos[name] = mi
}

func (s *MailStore) Remove(c *dynamic.Config) {
	s.m.Lock()
	defer s.m.Unlock()

	cfg := c.Data.(*mail.Config)
	name := cfg.Info.Name
	mi := s.infos[name]
	if mi == nil {
		return
	}
	if cfg := mi.Config(); cfg != nil && s.cfg.Api.Search.Enabled {
		s.removeFromIndex(cfg)
	}

	mi.removeConfig(c)

	if remaining := mi.Config(); remaining == nil {
		delete(s.infos, name)
		s.sm.ResetStores(events.NewTraits().WithNamespace("mail").WithName(name))
	} else if s.cfg.Api.Search.Enabled {
		s.addToIndex(remaining)
	}
}

func NewMailInfo(c *dynamic.Config) *MailInfo {
	mi := &MailInfo{
		Store:   mail.NewStore(&mail.Config{Info: getSmtpConfig(c).Info}),
		configs: map[string]*dynamic.Config{},
	}
	mi.addConfig(c)
	return mi
}

func (c *MailInfo) addConfig(config *dynamic.Config) {
	c.m.Lock()
	c.configs[config.Info.Url.String()] = config
	c.m.Unlock()
	c.update()
}

func (c *MailInfo) removeConfig(config *dynamic.Config) {
	c.m.Lock()
	delete(c.configs, config.Info.Url.String())
	c.m.Unlock()
	c.update()
}

// Config returns the current merged config, or nil after the last file was removed.
// The result must not be modified.
func (c *MailInfo) Config() *mail.Config {
	return c.config.Load()
}

// update publishes a new snapshot: the first file (ordered by file name) patched with the others.
// The merge works on a clone, so the parsed files are never changed. The store follows the new
// snapshot, but stays after the last file is removed.
func (c *MailInfo) update() {
	files := c.Configs()
	if len(files) == 0 {
		c.config.Store(nil)
		return
	}

	sort.Slice(files, func(i, j int) bool {
		return filepath.Base(files[i].Info.Url.String()) < filepath.Base(files[j].Info.Url.String())
	})

	r := getSmtpConfig(files[0]).Clone()
	for _, f := range files[1:] {
		log.Infof("applying patch for %s: %s", r.Info.Name, f.Info.Url)
		r.Patch(getSmtpConfig(f))
	}

	c.config.Store(r)
	c.Store.Update(r)
}

// Handler returns a handler for the current snapshot. A running server keeps it, a reload calls
// Reload to continue with the new snapshot.
func (c *MailInfo) Handler(smtp *monitor.Mail, emitter engine.MailEventEmitter, eh events.Handler) ReloadableMailHandler {
	h := &mailHandler{
		info:    c,
		monitor: smtp,
		emitter: emitter,
		events:  eh,
	}
	h.Reload()
	return h
}

func (c *MailInfo) Configs() []*dynamic.Config {
	c.m.RLock()
	defer c.m.RUnlock()

	var r []*dynamic.Config
	for _, config := range c.configs {
		r = append(r, config)
	}
	return r
}

type mailHandler struct {
	info    *MailInfo
	monitor *monitor.Mail
	emitter engine.MailEventEmitter
	events  events.Handler
	next    atomic.Pointer[mail.Handler]
}

func (h *mailHandler) Reload() {
	h.next.Store(mail.NewHandler(h.info.Config(), h.info.Store, h.emitter, h.events))
}

func (h *mailHandler) ServeSMTP(rw smtp.ResponseWriter, r smtp.Request) {
	r.WithContext(monitor.NewSmtpContext(r.Context(), h.monitor))
	r.WithContext(context.WithValue(r.Context(), "time", time.Now()))

	h.next.Load().ServeSMTP(rw, r)
}

func IsSmtpConfig(c *dynamic.Config) (*mail.Config, bool) {
	cfg, ok := c.Data.(*mail.Config)
	return cfg, ok
}

func (h *mailHandler) Login(username, password string, ctx context.Context) error {
	return h.next.Load().Login(username, password, ctx)
}

func (h *mailHandler) Select(mailbox string, readonly bool, ctx context.Context) (*imap.Selected, error) {
	return h.next.Load().Select(mailbox, readonly, ctx)
}

func (h *mailHandler) Unselect(ctx context.Context) error {
	return h.next.Load().Unselect(ctx)
}

func (h *mailHandler) List(ref, pattern string, flags []imap.MailboxFlags, ctx context.Context) ([]imap.ListEntry, error) {
	return h.next.Load().List(ref, pattern, flags, ctx)
}

func (h *mailHandler) Fetch(req *imap.FetchRequest, res imap.FetchResponse, ctx context.Context) error {
	return h.next.Load().Fetch(req, res, ctx)
}

func (h *mailHandler) Store(req *imap.StoreRequest, res imap.FetchResponse, ctx context.Context) error {
	return h.next.Load().Store(req, res, ctx)
}

func (h *mailHandler) Expunge(set *imap.IdSet, w imap.ExpungeWriter, ctx context.Context) error {
	return h.next.Load().Expunge(set, w, ctx)
}

func (h *mailHandler) Create(name string, opt *imap.CreateOptions, ctx context.Context) error {
	return h.next.Load().Create(name, opt, ctx)
}

func (h *mailHandler) Delete(name string, ctx context.Context) error {
	return h.next.Load().Delete(name, ctx)
}

func (h *mailHandler) Rename(existingName, newName string, ctx context.Context) error {
	return h.next.Load().Rename(existingName, newName, ctx)
}

func (h *mailHandler) Copy(set *imap.IdSet, dest string, w imap.CopyWriter, ctx context.Context) error {
	return h.next.Load().Copy(set, dest, w, ctx)
}

func (h *mailHandler) Move(set *imap.IdSet, dest string, w imap.MoveWriter, ctx context.Context) error {
	return h.next.Load().Move(set, dest, w, ctx)
}

func (h *mailHandler) Status(req *imap.StatusRequest, ctx context.Context) (imap.StatusResult, error) {
	return h.next.Load().Status(req, ctx)
}

func (h *mailHandler) Subscribe(mailbox string, ctx context.Context) error {
	return h.next.Load().Subscribe(mailbox, ctx)
}

func (h *mailHandler) Unsubscribe(mailbox string, ctx context.Context) error {
	return h.next.Load().Unsubscribe(mailbox, ctx)
}

func (h *mailHandler) Search(req *imap.SearchRequest, ctx context.Context) (*imap.SearchResponse, error) {
	return h.next.Load().Search(req, ctx)
}

func (h *mailHandler) Append(mailbox string, msg *smtp.Message, opt imap.AppendOptions, ctx context.Context) error {
	return h.next.Load().Append(mailbox, msg, opt, ctx)
}

func (h *mailHandler) Idle(w imap.UpdateWriter, done chan struct{}, ctx context.Context) error {
	return h.next.Load().Idle(w, done, ctx)
}

func getSmtpConfig(c *dynamic.Config) *mail.Config {
	return c.Data.(*mail.Config)
}

func (s *MailStore) Len() int {
	if s == nil {
		return 0
	}

	s.m.RLock()
	defer s.m.RUnlock()
	return len(s.infos)
}
