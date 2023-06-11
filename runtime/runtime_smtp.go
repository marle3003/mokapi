package runtime

import (
	"context"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/mail"
	engine "mokapi/engine/common"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"time"
)

type SmtpInfo struct {
	*mail.Config
	*mail.Store
	Name    string
	configs map[string]*mail.Config
}

func NewSmtpInfo(c *common.Config) *SmtpInfo {
	si := &SmtpInfo{
		configs: map[string]*mail.Config{},
	}
	si.AddConfig(c)
	return si
}

func (c *SmtpInfo) AddConfig(config *common.Config) {
	lc := config.Data.(*mail.Config)
	if len(c.Name) == 0 {
		c.Name = lc.Info.Name
	}

	key := config.Info.Url.String()
	c.configs[key] = lc
	c.update()
}

func (c *SmtpInfo) update() {
	cfg := &mail.Config{}
	cfg.Info.Name = c.Name
	for _, p := range c.configs {
		cfg.Patch(p)
	}

	c.Config = cfg

	if c.Store == nil {
		c.Store = mail.NewStore(cfg)
	} else {
		c.Store.Update(cfg)
	}
}

func (c *SmtpInfo) Handler(smtp *monitor.Smtp, emitter engine.EventEmitter) smtp.Handler {
	return &smtpHandler{
		smtp: smtp,
		next: mail.NewHandler(c.Config, c.Store, emitter),
	}
}

type smtpHandler struct {
	smtp *monitor.Smtp
	next smtp.Handler
}

func (h *smtpHandler) ServeSMTP(rw smtp.ResponseWriter, r smtp.Request) {
	r.WithContext(monitor.NewSmtpContext(r.Context(), h.smtp))
	r.WithContext(context.WithValue(r.Context(), "time", time.Now()))

	h.next.ServeSMTP(rw, r)
}

func IsSmtpConfig(c *common.Config) bool {
	_, ok := c.Data.(*mail.Config)
	return ok
}
