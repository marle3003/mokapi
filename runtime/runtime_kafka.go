package runtime

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
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
	m       sync.RWMutex
}

type KafkaInfo struct {
	*asyncapi3.Config
	*store.Store
	configs    map[string]*dynamic.Config
	seenTopics map[string]bool
}

type KafkaHandler struct {
	kafka *monitor.Kafka
	next  kafka.Handler
}

func NewKafkaInfo(c *dynamic.Config, store *store.Store) *KafkaInfo {
	hc := &KafkaInfo{
		configs:    map[string]*dynamic.Config{},
		Store:      store,
		seenTopics: map[string]bool{},
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
	if !ok {
		ki = NewKafkaInfo(c, store.New(cfg, emitter))
		s.infos[cfg.Info.Name] = ki

		events.ResetStores(events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name))
		events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name))
	} else {
		ki.AddConfig(c)
	}

	for name := range cfg.Channels {
		if _, ok := ki.seenTopics[name]; ok {
			continue
		}
		s.monitor.Kafka.Messages.WithLabel(cfg.Info.Name, name)
		s.monitor.Kafka.LastMessage.WithLabel(cfg.Info.Name, name)
		traits := events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name).With("topic", name)
		events.SetStore(sizeEventStore, traits)
		ki.seenTopics[name] = true
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
	ki.Remove(c)
	if len(ki.configs) == 0 {
		s.m.RUnlock()
		s.m.Lock()
		delete(s.infos, name)
		events.ResetStores(events.NewTraits().WithNamespace("kafka").WithName(name))
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

func IsKafkaConfig(c *dynamic.Config) (*asyncapi3.Config, bool) {
	if cfg, ok := c.Data.(*asyncapi3.Config); ok {
		return cfg, true
	}
	if old, ok := c.Data.(*asyncApi.Config); ok {
		cfg, err := old.Convert()
		if err != nil {
			return nil, false
		}
		return cfg, true
	}
	return nil, false
}

func (c *KafkaInfo) Remove(cfg *dynamic.Config) {
	delete(c.configs, cfg.Info.Url.String())
	c.update()
}

func getKafkaConfig(c *dynamic.Config) (*asyncapi3.Config, error) {
	if _, ok := c.Data.(*asyncapi3.Config); ok {
		return c.Data.(*asyncapi3.Config), nil
	} else {
		old := c.Data.(*asyncApi.Config)
		return old.Convert()
	}
}
