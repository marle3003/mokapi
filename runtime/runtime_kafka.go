package runtime

import (
	"github.com/blevesearch/bleve/v2"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/kafka"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"path/filepath"
	"sort"
	"sync"
)

type KafkaStore struct {
	infos   map[string]*KafkaInfo
	monitor *monitor.Monitor
	cfg     *static.Config
	events  *events.StoreManager
	index   bleve.Index
	m       sync.RWMutex
}

type KafkaInfo struct {
	*asyncapi3.Config
	*store.Store
	configs               map[string]*dynamic.Config
	seenTopics            map[string]bool
	updateEventAndMetrics func(k *KafkaInfo)
}

type KafkaHandler struct {
	kafka *monitor.Kafka
	next  kafka.Handler
}

func NewKafkaInfo(c *dynamic.Config, store *store.Store, updateEventAndMetrics func(info *KafkaInfo)) *KafkaInfo {
	hc := &KafkaInfo{
		configs:               map[string]*dynamic.Config{},
		Store:                 store,
		seenTopics:            map[string]bool{},
		updateEventAndMetrics: updateEventAndMetrics,
	}
	hc.AddConfig(c)
	return hc
}

func (s *KafkaStore) Get(name string) *KafkaInfo {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.infos[name]
}

func (s *KafkaStore) List() []*KafkaInfo {
	if s == nil {
		return nil
	}

	s.m.RLock()
	defer s.m.RUnlock()

	var list []*KafkaInfo
	for _, v := range s.infos {
		list = append(list, v)
	}
	return list
}

func (s *KafkaStore) Add(c *dynamic.Config, emitter common.EventEmitter) (*KafkaInfo, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*KafkaInfo)
	}
	cfg, err := getKafkaConfig(c)
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
		s.events.ResetStores(events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name))
		s.events.SetStore(int(eventStore.Size), events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name))

		ki = NewKafkaInfo(c, store.New(cfg, emitter, s.events), s.updateEventStore)
		s.infos[cfg.Info.Name] = ki
	} else {
		ki.AddConfig(c)
	}

	if s.cfg.Api.Search.Enabled {
		s.addToIndex(ki.Config)
	}

	return ki, nil
}

func (s *KafkaStore) Set(name string, ki *KafkaInfo) {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*KafkaInfo)
	}

	s.infos[name] = ki
}

func (s *KafkaStore) Remove(c *dynamic.Config) {
	s.m.RLock()

	cfg, err := getKafkaConfig(c)
	if err != nil {
		return
	}

	name := cfg.Info.Name
	ki := s.infos[name]

	if s.cfg.Api.Search.Enabled {
		s.removeFromIndex(ki.Config)
	}
	delete(ki.configs, c.Info.Url.String())
	ki.update()

	if len(ki.configs) == 0 {
		s.m.RUnlock()
		s.m.Lock()
		delete(s.infos, name)
		s.events.ResetStores(events.NewTraits().WithNamespace("kafka").WithName(name))
		s.m.Unlock()
	} else {
		s.m.RUnlock()
	}
}

func (c *KafkaInfo) AddConfig(config *dynamic.Config) {
	key := config.Info.Url.String()
	c.configs[key] = config
	c.update()
}

func (c *KafkaInfo) update() {
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
		p, err := getKafkaConfig(c.configs[k])
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

func (c *KafkaInfo) Handler(kafka *monitor.Kafka) kafka.Handler {
	return &KafkaHandler{kafka: kafka, next: c.Store}
}

func (c *KafkaInfo) Configs() []*dynamic.Config {
	var r []*dynamic.Config
	for _, config := range c.configs {
		r = append(r, config)
	}
	return r
}

func (h *KafkaHandler) ServeMessage(rw kafka.ResponseWriter, req *kafka.Request) {
	ctx := monitor.NewKafkaContext(req.Context, h.kafka)

	req.WithContext(ctx)
	h.next.ServeMessage(rw, req)
}

func IsAsyncApiConfig(c *dynamic.Config) (*asyncapi3.Config, bool) {
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

	return cfg, true
}

func getKafkaConfig(c *dynamic.Config) (*asyncapi3.Config, error) {
	if _, ok := c.Data.(*asyncapi3.Config); ok {
		return c.Data.(*asyncapi3.Config), nil
	} else {
		old := c.Data.(*asyncApi.Config)
		return old.Convert()
	}
}

func (s *KafkaStore) updateEventStore(k *KafkaInfo) {
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
		s.monitor.Kafka.Messages.WithLabel(k.Config.Info.Name, topicName)
		s.monitor.Kafka.LastMessage.WithLabel(k.Config.Info.Name, topicName)
		traits := events.NewTraits().WithNamespace("kafka").WithName(k.Config.Info.Name).With("topic", topicName)
		s.events.SetStore(int(eventStore.Size), traits)
		k.seenTopics[topicName] = true
	}
}
