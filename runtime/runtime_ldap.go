package runtime

import (
	"context"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/providers/directory"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type LdapStore struct {
	infos map[string]*LdapInfo
	cfg   *static.Config
	m     sync.RWMutex
}

type LdapInfo struct {
	*directory.Config
	configs      map[string]*dynamic.Config
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

		events.ResetStores(events.NewTraits().WithNamespace("ldap").WithName(cfg.Info.Name))
		events.SetStore(int(store.Size), events.NewTraits().WithNamespace("ldap").WithName(cfg.Info.Name))
	} else {
		li.AddConfig(c)
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
	s.m.RLock()

	cfg := c.Data.(*directory.Config)
	name := cfg.Info.Name
	li := s.infos[name]
	li.Remove(c)
	if len(li.configs) == 0 {
		s.m.RUnlock()
		s.m.Lock()
		delete(s.infos, name)
		events.ResetStores(events.NewTraits().WithNamespace("ldap").WithName(name))
		s.m.Unlock()
	} else {
		s.m.RUnlock()
	}
}

func NewLdapInfo(c *dynamic.Config, emitter common.EventEmitter) *LdapInfo {
	li := &LdapInfo{
		configs:      map[string]*dynamic.Config{},
		eventEmitter: emitter,
	}
	li.AddConfig(c)
	return li
}

func (c *LdapInfo) AddConfig(config *dynamic.Config) {
	c.configs[config.Info.Url.String()] = config
	c.update()
}

func (c *LdapInfo) update() {
	if len(c.configs) == 0 {
		c.Config = nil
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

	r := &directory.Config{}
	*r = *getLdapConfig(c.configs[keys[0]])
	for _, k := range keys[1:] {
		p := getLdapConfig(c.configs[k])
		log.Infof("applying patch for %s: %s", c.Info.Name, k)
		r.Patch(p)
	}

	c.Config = r
}

func (c *LdapInfo) Handler(ldap *monitor.Ldap) ldap.Handler {
	return &ldapHandler{ldap: ldap, next: directory.NewHandler(c.Config, c.eventEmitter)}
}

func (c *LdapInfo) Configs() []*dynamic.Config {
	var r []*dynamic.Config
	for _, config := range c.configs {
		r = append(r, config)
	}
	return r
}

func (h *ldapHandler) ServeLDAP(rw ldap.ResponseWriter, r *ldap.Request) {
	r.Context = monitor.NewLdapContext(r.Context, h.ldap)
	r.Context = context.WithValue(r.Context, "time", time.Now())

	h.next.ServeLDAP(rw, r)

}

func (c *LdapInfo) Remove(cfg *dynamic.Config) {
	delete(c.configs, cfg.Info.Url.String())
	c.update()
}

func IsLdapConfig(c *dynamic.Config) (*directory.Config, bool) {
	li, ok := c.Data.(*directory.Config)
	return li, ok
}

func getLdapConfig(c *dynamic.Config) *directory.Config {
	return c.Data.(*directory.Config)
}
