package runtime

import (
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/runtime/monitor"
	"mokapi/version"
	"sync"
)

type App struct {
	Version   string
	BuildTime string
	Http      *HttpStore
	Ldap      *LdapStore
	Kafka     *KafkaStore
	Mqtt      *MqttStore
	Mail      *MailStore

	Monitor *monitor.Monitor
	m       sync.Mutex
	cfg     *static.Config

	Configs map[string]*dynamic.Config
}

func New(cfg *static.Config) *App {
	m := monitor.New()
	return &App{
		Version:   version.BuildVersion,
		BuildTime: version.BuildTime,
		Monitor:   m,
		Configs:   map[string]*dynamic.Config{},
		Http:      &HttpStore{cfg: cfg},
		Kafka:     &KafkaStore{monitor: m, cfg: cfg},
		Mqtt:      &MqttStore{monitor: m, cfg: cfg},
		Ldap:      &LdapStore{cfg: cfg},
		Mail:      &MailStore{cfg: cfg},
		cfg:       cfg,
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
