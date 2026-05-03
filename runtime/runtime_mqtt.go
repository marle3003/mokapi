package runtime

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/mqtt/store"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/runtime/search"
	"mokapi/sortedmap"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

type MqttStore struct {
	infos   map[string]*MqttInfo
	monitor *monitor.Monitor
	cfg     *static.Config
	sm      *events.StoreManager
	m       sync.RWMutex
	events  *events.StoreManager
	index   search.Index
	reader  dynamic.Reader
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

func newMqttInfo(c *dynamic.Config, store *store.Store, updateEventAndMetrics func(info *MqttInfo)) *MqttInfo {
	hc := &MqttInfo{
		configs:               map[string]*dynamic.Config{},
		Store:                 store,
		seenTopics:            map[string]bool{},
		updateEventAndMetrics: updateEventAndMetrics,
	}
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
		s.sm.ResetStores(events.NewTraits().WithNamespace("mqtt").WithName(cfg.Info.Name))
		s.sm.SetStore(int(eventStore.Size), events.NewTraits().WithNamespace("mqtt").WithName(cfg.Info.Name))

		ki = newMqttInfo(c, store.New(cfg, emitter, s.events, s.monitor.Mqtt), s.updateEventStore)
		s.infos[cfg.Info.Name] = ki
	}
	ki.addConfig(c, s.reader)

	if s.cfg.Api.Search.Enabled {
		s.addToIndex(ki.Config)
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
	mi := s.infos[name]

	if s.cfg.Api.Search.Enabled {
		s.removeFromIndex(mi.Config)
	}

	delete(mi.configs, c.Info.Url.String())
	mi.update(s.reader)

	if len(mi.configs) == 0 {
		s.m.RUnlock()
		s.m.Lock()
		delete(s.infos, name)
		s.sm.ResetStores(events.NewTraits().WithNamespace("Mqtt").WithName(name))
		s.m.Unlock()
	} else {
		s.m.RUnlock()
	}
}

func (c *MqttInfo) addConfig(config *dynamic.Config, reader dynamic.Reader) {
	key := config.Info.Url.String()
	c.configs[key] = config
	c.update(reader)
}

func (c *MqttInfo) update(reader dynamic.Reader) {
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

	if len(c.configs) > 1 {
		err := cfg.Parse(&dynamic.Config{Data: cfg}, reader)
		if err != nil {
			log.Errorf("failed to parse config: %s", err)
		}
	}

	if cfg.Servers.Len() == 0 {
		log.Infof("no servers defined in AsyncAPI spec — using default Mokapi broker for cluster '%s'", cfg.Info.Name)
		if cfg.Servers == nil {
			cfg.Servers = &sortedmap.LinkedHashMap[string, *asyncapi3.ServerRef]{}
		}
		cfg.Servers.Set("mokapi", &asyncapi3.ServerRef{
			Value: &asyncapi3.Server{
				Host:     ":1883",
				Protocol: "mqtt",
				Title:    "Mokapi Default Broker",
				Summary:  "Automatically added broker because no servers are defined in the AsyncAPI spec",
			},
		})
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

func HasMqttServer(c *dynamic.Config) (*asyncapi3.Config, bool) {
	cfg, ok := IsAsyncApiConfig(c)
	if !ok {
		return nil, false
	}
	for it := cfg.Servers.Iter(); it.Next(); {
		s := it.Value()
		if s.Value == nil {
			continue
		}
		if strings.ToLower(s.Value.Protocol) == "mqtt" {
			return cfg, true
		}
	}
	return cfg, false
}

func getMqttConfig(c *dynamic.Config) (*asyncapi3.Config, error) {
	if _, ok := c.Data.(*asyncapi3.Config); ok {
		return c.Data.(*asyncapi3.Config), nil
	} else {
		old := c.Data.(*asyncApi.Config)
		return old.Convert()
	}
}

func (s *MqttStore) updateEventStore(k *MqttInfo) {
	eventStore, hasStoreConfig := s.cfg.Event.Store[k.Config.Info.Name]
	if !hasStoreConfig {
		eventStore = s.cfg.Event.Store["default"]
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
		s.monitor.Mqtt.Messages.WithLabel(k.Config.Info.Name, topicName)
		s.monitor.Mqtt.LastMessage.WithLabel(k.Config.Info.Name, topicName)
		traits := events.NewTraits().WithNamespace("mqtt").WithName(k.Config.Info.Name).With("topic", topicName)
		s.sm.SetStore(int(eventStore.Size), traits)
		k.seenTopics[topicName] = true
	}
}
