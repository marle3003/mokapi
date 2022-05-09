package runtime

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/config/dynamic/openapi"
	"mokapi/runtime/monitor"
)

type App struct {
	Http  map[string]*HttpInfo
	Ldap  map[string]*LdapInfo
	Kafka map[string]*KafkaInfo
	Smtp  map[string]*SmtpInfo

	Monitor *monitor.Monitor
}

func New() *App {
	return &App{
		Monitor: monitor.New(),
	}
}

func (a *App) AddHttp(c *openapi.Config) {
	if len(a.Http) == 0 {
		a.Http = make(map[string]*HttpInfo)
	}
	a.Http[c.Info.Name] = &HttpInfo{Config: c}
}

func (a *App) AddKafka(c *asyncApi.Config, store *store.Store) {
	if len(a.Kafka) == 0 {
		a.Kafka = make(map[string]*KafkaInfo)
	}
	a.Kafka[c.Info.Name] = &KafkaInfo{Config: c, Store: store}

	for n := range c.Channels {
		a.Monitor.Kafka.Messages.WithLabel(c.Info.Name, n).Add(0)
		a.Monitor.Kafka.LastMessage.WithLabel(c.Info.Name, n).Set(0)
	}
}
