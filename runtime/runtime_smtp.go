package runtime

import (
	"context"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/mail"
	"mokapi/config/static"
	engine "mokapi/engine/common"
	"mokapi/imap"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type MailHandler interface {
	smtp.Handler
	imap.Handler
}

type MailStore struct {
	infos map[string]*MailInfo
	m     sync.RWMutex
	cfg   *static.Config
}

type MailInfo struct {
	*mail.Config
	*mail.Store
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

		events.ResetStores(events.NewTraits().WithNamespace("smtp").WithName(cfg.Info.Name))
		events.SetStore(int(store.Size), events.NewTraits().WithNamespace("smtp").WithName(cfg.Info.Name))
	} else {
		mi.AddConfig(c)
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
	s.m.RLock()

	cfg := c.Data.(*mail.Config)
	name := cfg.Info.Name
	mi := s.infos[name]
	mi.Remove(c)
	if len(mi.configs) == 0 {
		s.m.RUnlock()
		s.m.Lock()
		delete(s.infos, name)
		events.ResetStores(events.NewTraits().WithNamespace("smtp").WithName(name))
		s.m.Unlock()
	} else {
		s.m.RUnlock()
	}
}

func NewMailInfo(c *dynamic.Config) *MailInfo {
	si := &MailInfo{
		configs: map[string]*dynamic.Config{},
	}
	si.AddConfig(c)
	return si
}

func (c *MailInfo) AddConfig(config *dynamic.Config) {
	c.configs[config.Info.Url.String()] = config
	c.update()
}

func (c *MailInfo) update() {
	if len(c.configs) == 0 {
		c.Config = nil
		c.Store = nil
		return
	}

	var keys []string
	for k := range c.configs {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		x := keys[i]
		y := keys[j]
		return filepath.Base(x) < filepath.Base(y)
	})

	cfg := &mail.Config{}
	*cfg = *getSmtpConfig(c.configs[keys[0]])
	for _, k := range keys[1:] {
		p := getSmtpConfig(c.configs[k])
		log.Infof("applying patch for %s: %s", c.Info.Name, k)
		cfg.Patch(p)
	}

	if len(cfg.Server) > 0 {
		log.Warnf("deprecated server configuration. Please use servers configuration")
		cfg.Servers = append(cfg.Servers, mail.Server{Url: cfg.Server})
	}

	c.Config = cfg
	if c.Store == nil {
		c.Store = mail.NewStore(cfg)
	} else {
		c.Store.Update(cfg)
	}
}

func (c *MailInfo) Handler(smtp *monitor.Smtp, emitter engine.EventEmitter) MailHandler {
	return &mailHandler{
		monitor: smtp,
		next:    mail.NewHandler(c.Config, c.Store, emitter),
	}
}

func (c *MailInfo) Configs() []*dynamic.Config {
	var r []*dynamic.Config
	for _, config := range c.configs {
		r = append(r, config)
	}
	return r
}

func (c *MailInfo) Remove(cfg *dynamic.Config) {
	delete(c.configs, cfg.Info.Url.String())
	c.update()
}

type mailHandler struct {
	monitor *monitor.Smtp
	next    *mail.Handler
}

func (h *mailHandler) ServeSMTP(rw smtp.ResponseWriter, r smtp.Request) {
	r.WithContext(monitor.NewSmtpContext(r.Context(), h.monitor))
	r.WithContext(context.WithValue(r.Context(), "time", time.Now()))

	h.next.ServeSMTP(rw, r)
}

func IsSmtpConfig(c *dynamic.Config) (*mail.Config, bool) {
	cfg, ok := c.Data.(*mail.Config)
	return cfg, ok
}

func (h *mailHandler) Login(username, password string, ctx context.Context) error {
	return h.next.Login(username, password, ctx)
}

func (h *mailHandler) Select(mailbox string, ctx context.Context) (*imap.Selected, error) {
	return h.next.Select(mailbox, ctx)
}

func (h *mailHandler) Unselect(ctx context.Context) error {
	return h.next.Unselect(ctx)
}

func (h *mailHandler) List(ref, pattern string, flags []imap.MailboxFlags, ctx context.Context) ([]imap.ListEntry, error) {
	return h.next.List(ref, pattern, flags, ctx)
}

func (h *mailHandler) Fetch(req *imap.FetchRequest, res imap.FetchResponse, ctx context.Context) error {
	return h.next.Fetch(req, res, ctx)
}

func (h *mailHandler) Store(req *imap.StoreRequest, res imap.FetchResponse, ctx context.Context) error {
	return h.next.Store(req, res, ctx)
}

func (h *mailHandler) Expunge(set *imap.IdSet, w imap.ExpungeWriter, ctx context.Context) error {
	return h.next.Expunge(set, w, ctx)
}

func (h *mailHandler) Create(name string, opt *imap.CreateOptions, ctx context.Context) error {
	return h.next.Create(name, opt, ctx)
}

func (h *mailHandler) Copy(set *imap.IdSet, dest string, w imap.CopyWriter, ctx context.Context) error {
	return h.next.Copy(set, dest, w, ctx)
}

func (h *mailHandler) Move(set *imap.IdSet, dest string, w imap.MoveWriter, ctx context.Context) error {
	return h.next.Move(set, dest, w, ctx)
}

func (h *mailHandler) Status(req *imap.StatusRequest, ctx context.Context) (imap.StatusResult, error) {
	return h.next.Status(req, ctx)
}

func (h *mailHandler) Subscribe(mailbox string, ctx context.Context) error {
	return h.next.Subscribe(mailbox, ctx)
}

func (h *mailHandler) Unsubscribe(mailbox string, ctx context.Context) error {
	return h.next.Unsubscribe(mailbox, ctx)
}

func getSmtpConfig(c *dynamic.Config) *mail.Config {
	return c.Data.(*mail.Config)
}
