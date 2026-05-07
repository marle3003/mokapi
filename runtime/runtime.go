package runtime

import (
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/runtime/search"
	"mokapi/safe"
	"mokapi/version"
	"sync"

	log "github.com/sirupsen/logrus"
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
	Events  *events.StoreManager
	Engine  common.EventEmitter

	m           sync.Mutex
	cfg         *static.Config
	searchIndex *SearchIndex

	reader dynamic.Reader
	hook   *LogHook

	Configs map[string]*dynamic.Config
}

func New(cfg *static.Config, reader dynamic.Reader) *App {
	m := monitor.New()

	index := newSearchIndex(cfg.Api.Search)
	em := events.NewStoreManager(index)
	configureLogging(cfg)

	em.SetStore(int(cfg.Event.Store["default"].Size), events.NewTraits())
	em.SetStore(int(cfg.Event.Store["default"].Size), events.NewTraits().WithNamespace("http"))
	em.SetStore(int(cfg.Event.Store["default"].Size), events.NewTraits().WithNamespace("kafka"))
	em.SetStore(int(cfg.Event.Store["default"].Size), events.NewTraits().WithNamespace("ldap"))
	em.SetStore(int(cfg.Event.Store["default"].Size), events.NewTraits().WithNamespace("mail"))
	em.SetStore(int(cfg.Event.Store["default"].Size), events.NewTraits().WithNamespace("job"))
	em.SetStore(int(cfg.Event.Store["default"].Size), events.NewTraits().WithNamespace("logs"))

	app := &App{
		Version:     version.BuildVersion,
		BuildTime:   version.BuildTime,
		Monitor:     m,
		Events:      em,
		Configs:     map[string]*dynamic.Config{},
		Http:        &HttpStore{cfg: cfg, index: index, events: em, reader: reader},
		Kafka:       &KafkaStore{monitor: m, cfg: cfg, index: index, events: em, reader: reader},
		Mqtt:        &MqttStore{monitor: m, cfg: cfg, index: index, sm: em, events: em, reader: reader},
		Ldap:        &LdapStore{cfg: cfg, events: em, index: index},
		Mail:        &MailStore{cfg: cfg, sm: em, index: index},
		cfg:         cfg,
		searchIndex: index,
		reader:      reader,
	}

	return app
}

func (a *App) Start(p *safe.Pool) {
	go a.searchIndex.start(p)
}

func (a *App) Stop() {
	if a.hook != nil {
		a.hook.Disable()
	}
}

func (a *App) EnableLogHook() {
	a.hook = NewLogHook(a.Events)
	log.AddHook(a.hook)
}

func (a *App) UpdateConfig(e dynamic.ConfigEvent) {
	a.m.Lock()

	if e.Event == dynamic.Delete {
		delete(a.Configs, e.Config.Info.Key())
	} else {
		a.Configs[e.Config.Info.Key()] = e.Config
		for _, r := range e.Config.Refs.List(true) {
			a.Configs[r.Info.Key()] = r
		}
	}
	a.m.Unlock()

	if a.cfg.Api.Search.Enabled {
		a.removeConfigFromIndex(e.Config)
		a.addConfigToIndex(e.Config)
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

func (a *App) Search(r search.Request) (search.Result, error) {
	return a.searchIndex.Search(r)
}
