package runtime

import (
	"context"
	cfg "mokapi/config/dynamic/common"
	"mokapi/config/dynamic/directory"
	engine "mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/runtime/monitor"
	"time"
)

type LdapInfo struct {
	*directory.Config
	Name         string
	configs      map[string]*directory.Config
	eventEmitter engine.EventEmitter
}

type ldapHandler struct {
	ldap *monitor.Ldap
	next ldap.Handler
}

func NewLdapInfo(c *cfg.Config, emitter engine.EventEmitter) *LdapInfo {
	li := &LdapInfo{
		configs:      map[string]*directory.Config{},
		eventEmitter: emitter,
	}
	li.AddConfig(c)
	return li
}

func (c *LdapInfo) AddConfig(config *cfg.Config) {
	lc := config.Data.(*directory.Config)
	if len(c.Name) == 0 {
		c.Name = lc.Info.Name
	}

	key := config.Info.Url.String()
	c.configs[key] = lc
	c.update()
}

func (c *LdapInfo) update() {
	cfg := &directory.Config{}
	cfg.Info.Name = c.Name
	for _, p := range c.configs {
		cfg.Patch(p)
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

func IsLdapConfig(c *cfg.Config) bool {
	_, ok := c.Data.(*directory.Config)
	return ok
}
