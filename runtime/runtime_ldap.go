package runtime

import (
	"context"
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
	configs      map[string]*directory.Config
	eventEmitter engine.EventEmitter
}

type ldapHandler struct {
	ldap *monitor.Ldap
	next ldap.Handler
}

func NewLdapInfo(c *dynamic.Config, emitter engine.EventEmitter) *LdapInfo {
	li := &LdapInfo{
		configs:      map[string]*directory.Config{},
		eventEmitter: emitter,
	}
	li.AddConfig(c)
	return li
}

func (c *LdapInfo) AddConfig(config *dynamic.Config) {
	lc := config.Data.(*directory.Config)
	key := config.Info.Url.String()
	c.configs[key] = lc
	c.update()
}

func (c *LdapInfo) update() {
	var keys []string
	for k := range c.configs {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		x := keys[i]
		y := keys[j]
		return filepath.Base(x) < filepath.Base(y)
	})

	cfg := c.configs[keys[0]]
	for _, k := range keys[1:] {
		cfg.Patch(c.configs[k])
	}

	c.Config = cfg
}

func (c *LdapInfo) Handler(ldap *monitor.Ldap) ldap.Handler {
	return &ldapHandler{ldap: ldap, next: directory.NewHandler(c.Config, c.eventEmitter)}
}

func (h *ldapHandler) ServeLDAP(rw ldap.ResponseWriter, r *ldap.Request) {
	r.Context = monitor.NewLdapContext(r.Context, h.ldap)
	r.Context = context.WithValue(r.Context, "time", time.Now())

	h.next.ServeLDAP(rw, r)

}

func IsLdapConfig(c *dynamic.Config) bool {
	_, ok := c.Data.(*directory.Config)
	return ok
}
