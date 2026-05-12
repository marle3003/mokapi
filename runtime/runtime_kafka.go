package runtime

import (
	"errors"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/kafka"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/kafka/store"
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

type KafkaStore struct {
	infos   map[string]*KafkaInfo
	monitor *monitor.Monitor
	cfg     *static.Config
	events  *events.StoreManager
	index   search.Index
	m       sync.RWMutex
	reader  dynamic.Reader
}

type KafkaInfo struct {
	*asyncapi3.Config
	*store.Store
	configs               map[string]*dynamic.Config
	seenTopics            map[string]bool
	updateEventAndMetrics func(k *KafkaInfo)
	m                     sync.Mutex
}

type KafkaHandler struct {
	kafka *monitor.Kafka
	next  kafka.Handler
}

func newKafkaInfo(store *store.Store, updateEventAndMetrics func(info *KafkaInfo)) *KafkaInfo {
	hc := &KafkaInfo{
		configs:               map[string]*dynamic.Config{},
		Store:                 store,
		seenTopics:            map[string]bool{},
		updateEventAndMetrics: updateEventAndMetrics,
	}
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
	cfg := getKafkaConfig(c)
	if cfg == nil {
		return nil, errors.New("no Kafka config found")
	}

	s.m.Lock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*KafkaInfo)
	}
	name := cfg.Info.Name
	ki, ok := s.infos[name]
	if !ok {
		log.Debugf("starting Kafka Store with topics: %v", len(cfg.Channels))
		ki = newKafkaInfo(store.NewEmpty(emitter, s.events, s.monitor.Kafka), s.updateEventStore)
		log.Debugf("end Kafka Store with topics: %v", len(cfg.Channels))
		ki.Config = cfg
		s.infos[cfg.Info.Name] = ki
	}

	defer s.m.Unlock()

	ki.m.Lock()
	defer ki.m.Unlock()

	eventStore, hasStoreConfig := s.cfg.Event.Store[name]
	if !hasStoreConfig {
		eventStore = s.cfg.Event.Store["default"]
	}

	if !ok {
		s.events.ResetStores(events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name))
		s.events.SetStore(int(eventStore.Size), events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name))
	}
	ki.addConfig(c, s.reader)

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

	cfg := getKafkaConfig(c)
	name := cfg.Info.Name
	ki := s.infos[name]

	if s.cfg.Api.Search.Enabled {
		s.removeFromIndex(ki.Config)
	}
	delete(ki.configs, c.Info.Url.String())
	ki.update(s.reader)

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

func (c *KafkaInfo) addConfig(config *dynamic.Config, reader dynamic.Reader) {
	key := config.Info.Url.String()
	c.configs[key] = config
	c.update(reader)
}

func (c *KafkaInfo) update(reader dynamic.Reader) {
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
		p := getKafkaConfig(c.configs[k])
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
				Host:     ":9092",
				Protocol: "kafka",
				Title:    "Mokapi Default Broker",
				Summary:  "Automatically added broker because no servers are defined in the AsyncAPI spec",
			},
		})
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
	switch v := c.Data.(type) {
	case *asyncapi3.Config:
		return v, true
	case *asyncApi.Config:
		conv, err := v.Convert()
		if err != nil {
			log.Errorf("failed to convert asyncapi 2.0 config: %s", err)
			return nil, false
		}
		return conv, true
	default:
		return nil, false
	}
}

func getKafkaConfig(c *dynamic.Config) *asyncapi3.Config {
	cfg, ok := IsAsyncApiConfig(c)
	if !ok {
		return nil
	}
	return cfg
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

func HasKafkaBroker(c *dynamic.Config) (*asyncapi3.Config, bool) {
	cfg, ok := IsAsyncApiConfig(c)
	if !ok {
		return nil, false
	}
	for it := cfg.Servers.Iter(); it.Next(); {
		s := it.Value()
		if s.Value == nil {
			continue
		}
		if strings.ToLower(s.Value.Protocol) == "kafka" {
			return cfg, true
		}
	}
	return cfg, false
}

func (s *KafkaStore) Len() int {
	if s == nil {
		return 0
	}

	s.m.RLock()
	defer s.m.RUnlock()
	return len(s.infos)
}
