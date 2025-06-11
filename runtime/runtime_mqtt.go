package runtime

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/mqtt/store"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"path/filepath"
	"sort"
	"sync"
)

type MqttStore struct {
	infos   map[string]*MqttInfo
	monitor *monitor.Monitor
	cfg     *static.Config
	m       sync.RWMutex
}

type MqttInfo struct {
	*asyncapi3.Config
	*store.Store
	configs               map[string]*dynamic.Config
	seenTopics            map[string]bool
	updateEventAndMetrics func(k *MqttInfo)
}

type MqttHandler struct {
	Mqtt *monitor.Mqtt
	next mqtt.Handler
}

func NewMqttInfo(c *dynamic.Config, store *store.Store, updateEventAndMetrics func(info *MqttInfo)) *MqttInfo {
	hc := &MqttInfo{
		configs:               map[string]*dynamic.Config{},
		Store:                 store,
		seenTopics:            map[string]bool{},
		updateEventAndMetrics: updateEventAndMetrics,
	}
	hc.AddConfig(c)
	return hc
}

func (s *MqttStore) Get(name string) *MqttInfo {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.infos[name]
}

func (s *MqttStore) List() []*MqttInfo {
	if s == nil {
		return nil
	}

	s.m.RLock()
	defer s.m.RUnlock()

	var list []*MqttInfo
	for _, v := range s.infos {
		list = append(list, v)
	}
	return list
}

func (s *MqttStore) Add(c *dynamic.Config, emitter common.EventEmitter) (*MqttInfo, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*MqttInfo)
	}
	cfg, err := getMqttConfig(c)
	if err != nil {
		return nil, err
	}

	name := cfg.Info.Name
	ki, ok := s.infos[name]

	eventStore, hasStoreConfig := s.cfg.Event.Store[name]
	if !hasStoreConfig {
		eventStore = s.cfg.Event.Store["default"]
	}

	if !ok {
		events.ResetStores(events.NewTraits().WithNamespace("Mqtt").WithName(cfg.Info.Name))
		events.SetStore(int(eventStore.Size), events.NewTraits().WithNamespace("Mqtt").WithName(cfg.Info.Name))

		ki = NewMqttInfo(c, store.New(cfg, emitter), s.updateEventStore)
		s.infos[cfg.Info.Name] = ki
	} else {
		ki.AddConfig(c)
	}

	return ki, nil
}

func (s *MqttStore) Set(name string, ki *MqttInfo) {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*MqttInfo)
	}

	s.infos[name] = ki
}

func (s *MqttStore) Remove(c *dynamic.Config) {
	s.m.RLock()

	cfg, err := getMqttConfig(c)
	if err != nil {
		return
	}

	name := cfg.Info.Name
	ki := s.infos[name]
	ki.Remove(c)
	if len(ki.configs) == 0 {
		s.m.RUnlock()
		s.m.Lock()
		delete(s.infos, name)
		events.ResetStores(events.NewTraits().WithNamespace("Mqtt").WithName(name))
		s.m.Unlock()
	} else {
		s.m.RUnlock()
	}
}

func (c *MqttInfo) AddConfig(config *dynamic.Config) {
	key := config.Info.Url.String()
	c.configs[key] = config
	c.update()
}

func (c *MqttInfo) update() {
	if len(c.configs) == 0 {
		c.Config = nil
		c.Store = nil
		return
	}

	var keys []string
	for k := range c.configs {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		x := keys[i]
		y := keys[j]
		return filepath.Base(x) < filepath.Base(y)
	})

	cfg := &asyncapi3.Config{}
	for i, k := range keys {
		p, err := getMqttConfig(c.configs[k])
		if err != nil {
			log.Errorf("patch %v failed: %v", c.configs[k].Info.Url, err)
		}
		if i == 0 {
			*cfg = *p
		} else {
			log.Infof("applying patch for %s: %s", cfg.Info.Name, k)
			cfg.Patch(p)
		}
	}

	c.Config = cfg
	c.updateEventAndMetrics(c)
	c.Store.Update(cfg)
}

func (c *MqttInfo) Handler(Mqtt *monitor.Mqtt) mqtt.Handler {
	return &MqttHandler{Mqtt: Mqtt, next: c.Store}
}

func (c *MqttInfo) Configs() []*dynamic.Config {
	var r []*dynamic.Config
	for _, config := range c.configs {
		r = append(r, config)
	}
	return r
}

func (h *MqttHandler) ServeMessage(rw mqtt.MessageWriter, req *mqtt.Message) {
	ctx := monitor.NewMqttContext(req.Context, h.Mqtt)

	req.WithContext(ctx)
	h.next.ServeMessage(rw, req)
}

func IsMqttConfig(c *dynamic.Config) (*asyncapi3.Config, bool) {
	var cfg *asyncapi3.Config
	if old, ok := c.Data.(*asyncApi.Config); ok {
		var err error
		cfg, err = old.Convert()
		if err != nil {
			return nil, false
		}
	} else {
		cfg, ok = c.Data.(*asyncapi3.Config)
		if !ok {
			return nil, false
		}
	}

	return cfg, hasMqttBroker(cfg)
}

func hasMqttBroker(c *asyncapi3.Config) bool {
	for _, server := range c.Servers {
		if server.Value.Protocol == "mqtt" {
			return true
		}
	}
	return false
}

func (c *MqttInfo) Remove(cfg *dynamic.Config) {
	delete(c.configs, cfg.Info.Url.String())
	c.update()
}

func getMqttConfig(c *dynamic.Config) (*asyncapi3.Config, error) {
	if _, ok := c.Data.(*asyncapi3.Config); ok {
		return c.Data.(*asyncapi3.Config), nil
	} else {
		old := c.Data.(*asyncApi.Config)
		return old.Convert()
	}
}

func (c *MqttStore) updateEventStore(k *MqttInfo) {
	eventStore, hasStoreConfig := c.cfg.Event.Store[k.Config.Info.Name]
	if !hasStoreConfig {
		eventStore = c.cfg.Event.Store["default"]
	}

	for topicName, topic := range k.Config.Channels {
		if topic.Value == nil {
			continue
		}
		if topic.Value.Address != "" {
			topicName = topic.Value.Address
		}
		if _, ok := k.seenTopics[topicName]; ok {
			continue
		}
		c.monitor.Mqtt.Messages.WithLabel(k.Config.Info.Name, topicName)
		c.monitor.Mqtt.LastMessage.WithLabel(k.Config.Info.Name, topicName)
		traits := events.NewTraits().WithNamespace("mqtt").WithName(k.Config.Info.Name).With("topic", topicName)
		events.SetStore(int(eventStore.Size), traits)
		k.seenTopics[topicName] = true
	}
}
