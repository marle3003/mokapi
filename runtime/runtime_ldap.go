package runtime

import (
	"context"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/providers/directory"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/runtime/search"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type LdapStore struct {
	infos  map[string]*LdapInfo
	cfg    *static.Config
	events *events.StoreManager
	m      sync.RWMutex
	index  search.Index
}

// LdapInfo publishes the merged config as an immutable snapshot. A reload replaces the whole
// config including the directory entries, so a reader takes one snapshot via Config and uses only
// that. The result of Config must not be modified, except for the entries, which have their own
// lock.
type LdapInfo struct {
	configSnapshot[directory.Config]
	eventEmitter common.EventEmitter
}

type ldapHandler struct {
	ldap *monitor.Ldap
	next ldap.Handler
}

func (s *LdapStore) Get(name string) *LdapInfo {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.infos[name]
}

func (s *LdapStore) List() []*LdapInfo {
	if s == nil {
		return nil
	}

	s.m.RLock()
	defer s.m.RUnlock()

	var list []*LdapInfo
	for _, v := range s.infos {
		list = append(list, v)
	}
	return list
}

func (s *LdapStore) Add(c *dynamic.Config, emitter common.EventEmitter) *LdapInfo {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*LdapInfo)
	}
	cfg := c.Data.(*directory.Config)
	name := cfg.Info.Name
	li, ok := s.infos[name]

	store, hasStoreConfig := s.cfg.Event.Store[name]
	if !hasStoreConfig {
		store = s.cfg.Event.Store["default"]
	}

	if !ok {
		li = NewLdapInfo(c, emitter)
		s.infos[cfg.Info.Name] = li

		s.events.ResetStores(events.NewTraits().WithNamespace("ldap").WithName(cfg.Info.Name))
		s.events.SetStore(int(store.Size), events.NewTraits().WithNamespace("ldap").WithName(cfg.Info.Name))
	} else {
		li.add(c, mergeLdapConfigs)
	}

	if s.cfg.Api.Search.Enabled {
		s.addToIndex(li.Config())
	}

	return li
}

func (s *LdapStore) Set(name string, li *LdapInfo) {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*LdapInfo)
	}

	s.infos[name] = li
}

func (s *LdapStore) Remove(c *dynamic.Config) {
	s.m.Lock()
	defer s.m.Unlock()

	cfg := c.Data.(*directory.Config)
	name := cfg.Info.Name
	li := s.infos[name]
	if li == nil {
		return
	}
	if cfg := li.Config(); cfg != nil && s.cfg.Api.Search.Enabled {
		s.removeFromIndex(cfg)
	}

	li.remove(c, mergeLdapConfigs)

	if remaining := li.Config(); remaining == nil {
		delete(s.infos, name)
		s.events.ResetStores(events.NewTraits().WithNamespace("ldap").WithName(name))
	} else if s.cfg.Api.Search.Enabled {
		s.addToIndex(remaining)
	}
}

func NewLdapInfo(c *dynamic.Config, emitter common.EventEmitter) *LdapInfo {
	li := &LdapInfo{
		eventEmitter: emitter,
	}
	li.add(c, mergeLdapConfigs)
	return li
}

// mergeLdapConfigs builds a new snapshot: the first file (ordered by file name) patched with the others.
// The merge works on a clone, so the parsed files are never changed.
func mergeLdapConfigs(files []*dynamic.Config) *directory.Config {
	r := getLdapConfig(files[0]).Clone()
	for _, f := range files[1:] {
		log.Infof("applying patch for %s: %s", r.Info.Name, f.Info.Url)
		r.Patch(getLdapConfig(f))
	}
	return r
}

func (c *LdapInfo) Handler(ldap *monitor.Ldap, eh events.Handler) ldap.Handler {
	return &ldapHandler{ldap: ldap, next: directory.NewHandler(c.Config(), c.eventEmitter, eh)}
}

func (h *ldapHandler) ServeLDAP(rw ldap.ResponseWriter, r *ldap.Request) {
	r.Context = monitor.NewLdapContext(r.Context, h.ldap)
	r.Context = context.WithValue(r.Context, "time", time.Now())

	h.next.ServeLDAP(rw, r)
}

func (h *ldapHandler) Unbind(ctx context.Context) {
	ctx = monitor.NewLdapContext(ctx, h.ldap)
	ctx = context.WithValue(ctx, "time", time.Now())

	h.next.Unbind(ctx)
}

func IsLdapConfig(c *dynamic.Config) (*directory.Config, bool) {
	li, ok := c.Data.(*directory.Config)
	return li, ok
}

func getLdapConfig(c *dynamic.Config) *directory.Config {
	return c.Data.(*directory.Config)
}

func (s *LdapStore) Len() int {
	if s == nil {
		return 0
	}

	s.m.RLock()
	defer s.m.RUnlock()
	return len(s.infos)
}
