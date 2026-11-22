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

// KafkaInfo publishes the merged config as an immutable snapshot. A reload replaces the
// whole config, so a reader takes one snapshot via Config and uses only that.
type KafkaInfo struct {
	*store.Store
	configSnapshot[asyncapi3.Config]
	seenTopics            map[string]bool
	updateEventAndMetrics func(k *KafkaInfo)
}

type KafkaHandler struct {
	kafka *monitor.Kafka
	next  kafka.Handler
}

func newKafkaInfo(store *store.Store, updateEventAndMetrics func(info *KafkaInfo)) *KafkaInfo {
	hc := &KafkaInfo{
		Store:                 store,
		seenTopics:            map[string]bool{},
		updateEventAndMetrics: updateEventAndMetrics,
	}
	return hc
}

// NewKafkaInfo returns an info that publishes cfg as is, without merging files and
// without the implicit default broker of Add. It is meant for tests of readers: it has no
// event updater, so adding a file to it through KafkaStore.Add panics.
func NewKafkaInfo(cfg *asyncapi3.Config, store *store.Store) *KafkaInfo {
	ki := newKafkaInfo(store, nil)
	ki.publish(cfg)
	return ki
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

func (s *KafkaStore) Add(c *dynamic.Config, emitter common.KafkaEventEmitter) (*KafkaInfo, error) {
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
		ki = newKafkaInfo(store.NewEmpty(emitter, s.events, s.monitor.Kafka), s.updateEventStore)
		s.infos[cfg.Info.Name] = ki
	}

	defer s.m.Unlock()

	eventStore, hasStoreConfig := s.cfg.Event.Store[name]
	if !hasStoreConfig {
		eventStore = s.cfg.Event.Store["default"]
	}

	if !ok {
		s.events.ResetStores(events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name))
		s.events.SetStore(int(eventStore.Size), events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name))
	}
	ki.afterPublish(ki.add(c, mergeKafkaConfigs(s.reader)))

	if s.cfg.Api.Search.Enabled {
		s.addToIndex(ki.Config())
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
	s.m.Lock()
	defer s.m.Unlock()

	cfg := getKafkaConfig(c)
	name := cfg.Info.Name
	ki := s.infos[name]
	if ki == nil {
		return
	}

	if s.cfg.Api.Search.Enabled {
		s.removeFromIndex(ki.Config())
	}
	ki.afterPublish(ki.remove(c, mergeKafkaConfigs(s.reader)))

	if remaining := ki.Config(); remaining == nil {
		delete(s.infos, name)
		s.events.ResetStores(events.NewTraits().WithNamespace("kafka").WithName(name))
	} else if s.cfg.Api.Search.Enabled {
		s.addToIndex(remaining)
	}
}

// mergeKafkaConfigs merges the files of a service and adds the default server if the spec
// defines none.
func mergeKafkaConfigs(reader dynamic.Reader) func(files []*dynamic.Config) *asyncapi3.Config {
	return func(files []*dynamic.Config) *asyncapi3.Config {
		cfg := mergeAsyncApiConfigs(files, reader)
		defaultAsyncApiServer(cfg, "kafka", "kafka: using implicit default broker", &asyncapi3.Server{
			Host:     ":9092",
			Protocol: "kafka",
			Title:    "Mokapi Default Broker",
			Summary:  "Automatically added broker because no Kafka servers are defined in the AsyncAPI spec",
		})
		return cfg
	}
}

func (c *KafkaInfo) afterPublish(cfg *asyncapi3.Config) {
	if cfg == nil {
		return
	}
	c.updateEventAndMetrics(c)
	c.Store.Update(cfg)
}

func (c *KafkaInfo) Handler(kafka *monitor.Kafka) kafka.Handler {
	return &KafkaHandler{kafka: kafka, next: c.Store}
}

func (h *KafkaHandler) ServeMessage(rw kafka.ResponseWriter, req *kafka.Request) {
	ctx := monitor.NewKafkaContext(req.Context, h.kafka)

	req.WithContext(ctx)
	h.next.ServeMessage(rw, req)
}

func IsAsyncApiConfig(c *dynamic.Config) (*asyncapi3.Config, bool) {
	switch v := c.Data().(type) {
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
	cfg := k.Config()
	eventStore, hasStoreConfig := s.cfg.Event.Store[cfg.Info.Name]
	if !hasStoreConfig {
		eventStore = s.cfg.Event.Store["default"]
	}

	for topicName, topic := range cfg.Channels {
		if topic.Value == nil {
			continue
		}
		if topic.Value.Address != "" {
			topicName = topic.Value.Address
		}
		if _, ok := k.seenTopics[topicName]; ok {
			continue
		}
		s.monitor.Kafka.Messages.WithLabel(cfg.Info.Name, topicName)
		s.monitor.Kafka.LastMessage.WithLabel(cfg.Info.Name, topicName)
		traits := events.NewTraits().WithNamespace("kafka").WithName(cfg.Info.Name).With("topic", topicName)
		s.events.SetStore(int(eventStore.Size), traits)
		k.seenTopics[topicName] = true
	}
}

func (s *KafkaStore) Len() int {
	if s == nil {
		return 0
	}

	s.m.RLock()
	defer s.m.RUnlock()
	return len(s.infos)
}

type KafkaMode int

const (
	KafkaNone KafkaMode = iota
	// KafkaExplicit Server with protocol kafka
	KafkaExplicit
	// KafkaImplicit no servers, Kafka derived from bindings
	KafkaImplicit
)

func ShouldStartKafka(c *dynamic.Config) (*asyncapi3.Config, KafkaMode) {
	cfg, ok := IsAsyncApiConfig(c)
	if !ok {
		return nil, KafkaNone
	}
	a := cfg.Analyze()
	k := a.Usage("kafka")
	switch {
	case len(k.Servers) > 0:
		return cfg, KafkaExplicit
	case !a.HasServers && len(k.Channels) > 0:
		return cfg, KafkaImplicit
	}
	return cfg, KafkaNone
}
