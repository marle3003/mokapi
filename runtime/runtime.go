package runtime

import (
	"mokapi/config/dynamic"
	"mokapi/runtime/monitor"
	"mokapi/version"
	"sync"
)

const sizeEventStore = 20

type App struct {
	Version   string
	BuildTime string
	Http      *HttpStore
	Ldap      *LdapStore
	Kafka     *KafkaStore
	Mail      *MailStore

	Monitor *monitor.Monitor
	m       sync.Mutex

	Configs map[string]*dynamic.Config
}

func New() *App {
	m := monitor.New()
	return &App{
		Version:   version.BuildVersion,
		BuildTime: version.BuildTime,
		Monitor:   m,
		Configs:   map[string]*dynamic.Config{},
		Http:      &HttpStore{},
		Kafka:     &KafkaStore{monitor: m},
		Ldap:      &LdapStore{},
		Mail:      &MailStore{},
	}
}

func (a *App) UpdateConfig(e dynamic.ConfigEvent) {
	a.m.Lock()
	defer a.m.Unlock()

	if e.Event == dynamic.Delete {
		delete(a.Configs, e.Config.Info.Key())
	} else {
		a.Configs[e.Config.Info.Key()] = e.Config
		for _, r := range e.Config.Refs.List(true) {
			a.Configs[r.Info.Key()] = r
		}
	}
}

func (a *App) FindConfig(key string) *dynamic.Config {
	c, ok := a.Configs[key]
	if ok {
		return c
	}

	for _, c = range a.Configs {
		for _, ref := range c.Refs.List(true) {
			if ref.Info.Key() == key {
				return ref
			}
		}
	}

	return nil
}
