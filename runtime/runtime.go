package runtime

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/directory"
	"mokapi/config/dynamic/mail"
	"mokapi/config/dynamic/openapi"
	common2 "mokapi/engine/common"
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

func (a *App) AddKafka(c *common.Config, emitter common2.EventEmitter) *KafkaInfo {
	if len(a.Kafka) == 0 {
		a.Kafka = make(map[string]*KafkaInfo)
	}

	cfg := c.Data.(*asyncApi.Config)
	name := cfg.Info.Name
	hc, ok := a.Kafka[name]
	if !ok {
		hc = NewKafkaInfo(c, store.New(cfg, emitter))
		a.Kafka[cfg.Info.Name] = hc
	} else {
		hc.AddConfig(c)
	}

	events.ResetStores(events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name))
	events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name))
	for name := range cfg.Channels {
		a.Monitor.Kafka.Messages.WithLabel(cfg.Info.Name, name).Add(0)
		a.Monitor.Kafka.LastMessage.WithLabel(cfg.Info.Name, name).Set(0)
		events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name).With("topic", name))
	}
	return hc
}

func (a *App) AddSmtp(c *common.Config) *SmtpInfo {
	if len(a.Smtp) == 0 {
		a.Smtp = make(map[string]*SmtpInfo)
	}

	cfg := c.Data.(*mail.Config)
	name := cfg.Info.Name
	hc, ok := a.Smtp[name]
	if !ok {
		hc = NewSmtpInfo(c)
		a.Smtp[cfg.Info.Name] = hc
	} else {
		hc.AddConfig(c)
	}

	events.ResetStores(events.NewTraits().WithNamespace("smtp").WithName(cfg.Info.Name))
	events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("smtp").WithName(cfg.Info.Name))

	return hc
}

func (a *App) AddLdap(c *common.Config, emitter common2.EventEmitter) *LdapInfo {
	if len(a.Ldap) == 0 {
		a.Ldap = make(map[string]*LdapInfo)
	}

	cfg := c.Data.(*directory.Config)
	name := cfg.Info.Name
	hc, ok := a.Ldap[name]
	if !ok {
		hc = NewLdapInfo(c, emitter)
		a.Ldap[cfg.Info.Name] = hc
	} else {
		hc.AddConfig(c)
	}

	events.ResetStores(events.NewTraits().WithNamespace("ldap").WithName(cfg.Info.Name))
	events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("ldap").WithName(cfg.Info.Name))

	return hc
}
