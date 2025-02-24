package runtime

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/mail"
	"mokapi/engine/common"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/providers/directory"
	"mokapi/providers/openapi"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/version"
	"sync"
)

const sizeEventStore = 20

type App struct {
	Version   string
	BuildTime string
	Http      map[string]*HttpInfo
	Ldap      map[string]*LdapInfo
	Kafka     map[string]*KafkaInfo
	Smtp      map[string]*SmtpInfo

	Monitor *monitor.Monitor
	m       sync.Mutex

	Configs map[string]*dynamic.Config
}

func New() *App {
	return &App{
		Version:   version.BuildVersion,
		BuildTime: version.BuildTime,
		Monitor:   monitor.New(),
		Configs:   map[string]*dynamic.Config{},
	}
}

func (a *App) GetHttp(c *dynamic.Config) (string, *HttpInfo) {
	a.m.Lock()
	defer a.m.Unlock()

	if len(a.Http) == 0 {
		return "", nil
	}
	cfg := c.Data.(*openapi.Config)
	return cfg.Info.Name, a.Http[cfg.Info.Name]
}

func (a *App) AddHttp(c *dynamic.Config) *HttpInfo {
	a.m.Lock()
	defer a.m.Unlock()

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
	for path := range cfg.Paths {
		events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("http").WithName(name).With("path", path))
	}

	return hc
}

func (a *App) GetKafka(c *dynamic.Config) (string, *KafkaInfo) {
	a.m.Lock()
	defer a.m.Unlock()

	if len(a.Kafka) == 0 {
		return "", nil
	}
	cfg, err := getKafkaConfig(c)
	if err != nil {
		return "", nil
	}
	return cfg.Info.Name, a.Kafka[cfg.Info.Name]
}

func (a *App) AddKafka(c *dynamic.Config, emitter common.EventEmitter) (*KafkaInfo, error) {
	a.m.Lock()
	defer a.m.Unlock()

	if len(a.Kafka) == 0 {
		a.Kafka = make(map[string]*KafkaInfo)
	}

	cfg, err := getKafkaConfig(c)
	if err != nil {
		return nil, err
	}

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
	return hc, nil
}

func (a *App) GetSmtp(c *dynamic.Config) (string, *SmtpInfo) {
	a.m.Lock()
	defer a.m.Unlock()

	if len(a.Smtp) == 0 {
		return "", nil
	}
	cfg := c.Data.(*mail.Config)
	return cfg.Info.Name, a.Smtp[cfg.Info.Name]
}

func (a *App) AddSmtp(c *dynamic.Config) *SmtpInfo {
	a.m.Lock()
	defer a.m.Unlock()

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

func (a *App) GetLdap(c *dynamic.Config) (string, *LdapInfo) {
	a.m.Lock()
	defer a.m.Unlock()

	if len(a.Ldap) == 0 {
		return "", nil
	}
	cfg := c.Data.(*directory.Config)
	return cfg.Info.Name, a.Ldap[cfg.Info.Name]
}

func (a *App) AddLdap(c *dynamic.Config, emitter common.EventEmitter) *LdapInfo {
	a.m.Lock()
	defer a.m.Unlock()

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

func (a *App) RemoveHttp(c *dynamic.Config) {
	a.m.Lock()
	defer a.m.Unlock()

	cfg := c.Data.(*openapi.Config)
	name := cfg.Info.Name
	hc := a.Http[name]
	hc.Remove(c)
	if len(hc.configs) == 0 {
		delete(a.Http, name)
		events.ResetStores(events.NewTraits().WithNamespace("http").WithName(name))
	}
}

func (a *App) RemoveKafka(c *dynamic.Config) {
	a.m.Lock()
	defer a.m.Unlock()

	cfg, err := getKafkaConfig(c)
	if err != nil {
		return
	}

	name := cfg.Info.Name
	hc := a.Kafka[name]
	hc.Remove(c)

	if len(hc.configs) == 0 {
		delete(a.Kafka, name)
		events.ResetStores(events.NewTraits().WithNamespace("kafka").WithName(name))
	}
}

func (a *App) RemoveLdap(c *dynamic.Config) {
	a.m.Lock()
	defer a.m.Unlock()

	cfg := c.Data.(*directory.Config)
	name := cfg.Info.Name
	hc := a.Ldap[name]
	hc.Remove(c)
	if len(hc.configs) == 0 {
		delete(a.Ldap, name)
		events.ResetStores(events.NewTraits().WithNamespace("ldap").WithName(name))
	}
}

func (a *App) RemoveSmtp(c *dynamic.Config) {
	a.m.Lock()
	defer a.m.Unlock()

	cfg := c.Data.(*mail.Config)
	name := cfg.Info.Name
	hc := a.Smtp[name]
	hc.Remove(c)
	if len(hc.configs) == 0 {
		delete(a.Smtp, name)
		events.ResetStores(events.NewTraits().WithNamespace("smtp").WithName(name))
	}
}
