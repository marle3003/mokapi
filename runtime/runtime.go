package runtime

import (
	"github.com/blevesearch/bleve/v2"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/version"
	"sync"
)

type App struct {
	Version   string
	BuildTime string
	http      *HttpStore
	Ldap      *LdapStore
	Kafka     *KafkaStore
	Mqtt      *MqttStore
	Mail      *MailStore

	Monitor *monitor.Monitor
	Events  *events.StoreManager

	m     sync.Mutex
	cfg   *static.Config
	index bleve.Index

	Configs map[string]*dynamic.Config
}

func New(cfg *static.Config) *App {
	m := monitor.New()

	index := newIndex(cfg)
	em := events.NewStoreManager(index)

	em.SetStore(int(cfg.Event.Store["default"].Size), events.NewTraits().WithNamespace("http"))
	em.SetStore(int(cfg.Event.Store["default"].Size), events.NewTraits().WithNamespace("kafka"))
	em.SetStore(int(cfg.Event.Store["default"].Size), events.NewTraits().WithNamespace("ldap"))
	em.SetStore(int(cfg.Event.Store["default"].Size), events.NewTraits().WithNamespace("mail"))
	em.SetStore(int(cfg.Event.Store["default"].Size), events.NewTraits().WithNamespace("job"))

	app := &App{
		Version:   version.BuildVersion,
		BuildTime: version.BuildTime,
		Monitor:   m,
		Events:    em,
		Configs:   map[string]*dynamic.Config{},
		http:      NewHttpStore(cfg, index, em),
		Kafka:     &KafkaStore{monitor: m, cfg: cfg, index: index, events: em},
		Mqtt:      &MqttStore{monitor: m, cfg: cfg, sm: em},
		Ldap:      &LdapStore{cfg: cfg, events: em, index: index},
		Mail:      &MailStore{cfg: cfg, sm: em, index: index},
		cfg:       cfg,
		index:     index,
	}

	return app
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

	if a.cfg.Api.Search.Enabled {
		removeConfigFromIndex(a.index, e.Config)
		if err := addConfigToIndex(a.index, e.Config); err != nil {
			log.Errorf("add '%s' to search index failed", e.Config.Info.Path())
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

func (a *App) AddHttp(c *dynamic.Config) *HttpInfo {
	return a.http.Add(c)
}

func (a *App) GetHttp(name string) *HttpInfo {
	return a.http.Get(name)
}

func (a *App) RemoveHttp(c *dynamic.Config) {
	a.http.Remove(c)
}

func (a *App) ListHttp() []*HttpInfo {
	return a.http.List()
}
