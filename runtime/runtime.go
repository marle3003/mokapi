package runtime

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/directory"
	"mokapi/config/dynamic/mail"
	"mokapi/config/dynamic/openapi"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/version"
)

const sizeEventStore = 20

type App struct {
	Version string
	Http    map[string]*HttpInfo
	Ldap    map[string]*LdapInfo
	Kafka   map[string]*KafkaInfo
	Smtp    map[string]*SmtpInfo

	Monitor *monitor.Monitor
}

func New() *App {
	return &App{
		Version: version.BuildVersion,
		Monitor: monitor.New(),
	}
}

func (a *App) AddHttp(c *common.Config) *HttpInfo {
	if len(a.Http) == 0 {
		a.Http = make(map[string]*HttpInfo)
	}
	cfg := c.Data.(*openapi.Config)
	name := cfg.Info.Name
	hc, ok := a.Http[name]
	if !ok {
		hc = NewHttpInfo(c)
		a.Http[cfg.Info.Name] = hc
	} else {
		hc.AddConfig(c)
	}

	events.ResetStores(events.NewTraits().WithNamespace("http").WithName(name))
	events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("http").WithName(name))
	for path := range cfg.Paths.Value {
		events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("http").WithName(name).With("path", path))
	}

	return hc
}

func (a *App) AddKafka(c *asyncApi.Config, store *store.Store) {
	if len(a.Kafka) == 0 {
		a.Kafka = make(map[string]*KafkaInfo)
	}

	a.Kafka[c.Info.Name] = &KafkaInfo{Config: c, Store: store}

	events.ResetStores(events.NewTraits().WithNamespace("kafka").WithName(c.Info.Name))
	events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("kafka").WithName(c.Info.Name))
	for name := range c.Channels {
		a.Monitor.Kafka.Messages.WithLabel(c.Info.Name, name).Add(0)
		a.Monitor.Kafka.LastMessage.WithLabel(c.Info.Name, name).Set(0)
		events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("kafka").WithName(c.Info.Name).With("topic", name))
	}
}

func (a *App) AddSmtp(c *mail.Config, store *mail.Store) {
	if len(a.Smtp) == 0 {
		a.Smtp = make(map[string]*SmtpInfo)
	}

	a.Smtp[c.Info.Name] = &SmtpInfo{Config: c, Store: store}

	events.ResetStores(events.NewTraits().WithNamespace("smtp").WithName(c.Info.Name))
	events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("smtp").WithName(c.Info.Name))
}

func (a *App) AddLdap(c *directory.Config) {
	if len(a.Ldap) == 0 {
		a.Ldap = make(map[string]*LdapInfo)
	}
	a.Ldap[c.Info.Name] = &LdapInfo{Config: c}

	events.ResetStores(events.NewTraits().WithNamespace("ldap").WithName(c.Info.Name))
	events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("ldap").WithName(c.Info.Name))
}
