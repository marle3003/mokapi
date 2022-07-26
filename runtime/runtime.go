package runtime

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka/store"
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

func (a *App) AddHttp(c *openapi.Config) {
	if len(a.Http) == 0 {
		a.Http = make(map[string]*HttpInfo)
	}
	a.Http[c.Info.Name] = &HttpInfo{Config: c}

	events.ResetStores(events.NewTraits().WithNamespace("http").WithName(c.Info.Name))
	events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("http").WithName(c.Info.Name))
	for path := range c.Paths.Value {
		events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("http").WithName(c.Info.Name).With("path", path))
	}
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
