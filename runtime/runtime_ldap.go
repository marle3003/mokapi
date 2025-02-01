package runtime

import (
	"context"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/directory"
	engine "mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/runtime/monitor"
	"path/filepath"
	"sort"
	"time"
)

type LdapInfo struct {
	*directory.Config
	configs      map[string]*dynamic.Config
	eventEmitter engine.EventEmitter
}

type ldapHandler struct {
	ldap *monitor.Ldap
	next ldap.Handler
}

func NewLdapInfo(c *dynamic.Config, emitter engine.EventEmitter) *LdapInfo {
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

func IsLdapConfig(c *dynamic.Config) bool {
	_, ok := c.Data.(*directory.Config)
	return ok
}

func getLdapConfig(c *dynamic.Config) *directory.Config {
	return c.Data.(*directory.Config)
}
