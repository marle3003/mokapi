package runtime

import (
	"context"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/mail"
	engine "mokapi/engine/common"
	"mokapi/imap"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"path/filepath"
	"sort"
	"time"
)

type MailHandler interface {
	smtp.Handler
	imap.Handler
}

type SmtpInfo struct {
	*mail.Config
	*mail.Store
	configs map[string]*mail.Config
}

func NewSmtpInfo(c *dynamic.Config) *SmtpInfo {
	si := &SmtpInfo{
		configs: map[string]*mail.Config{},
	}
	si.AddConfig(c)
	return si
}

func (c *SmtpInfo) AddConfig(config *dynamic.Config) {
	lc := config.Data.(*mail.Config)
	key := config.Info.Url.String()
	c.configs[key] = lc
	c.update()
}

func (c *SmtpInfo) update() {
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
	if c.Store == nil {
		c.Store = mail.NewStore(cfg)
	} else {
		c.Store.Update(cfg)
	}
}

func (c *SmtpInfo) Handler(smtp *monitor.Smtp, emitter engine.EventEmitter) MailHandler {
	return &mailHandler{
		monitor: smtp,
		next:    mail.NewHandler(c.Config, c.Store, emitter),
	}
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

func IsSmtpConfig(c *dynamic.Config) bool {
	_, ok := c.Data.(*mail.Config)
	return ok
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

func (h *mailHandler) List(ref, pattern string, ctx context.Context) ([]imap.ListEntry, error) {
	return h.next.List(ref, pattern, ctx)
}

func (h *mailHandler) Fetch(req *imap.FetchRequest, res imap.FetchResponse, ctx context.Context) error {
	return h.next.Fetch(req, res, ctx)
}
