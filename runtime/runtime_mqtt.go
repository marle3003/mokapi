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
	"sync"
)

type MqttStore struct {
	infos   map[string]*MqttInfo
	monitor *monitor.Monitor
	cfg     *static.Config
	m       sync.RWMutex
	events  *events.StoreManager
	index   search.Index
	reader  dynamic.Reader
}

// MqttInfo publishes the merged config as an immutable snapshot. A reload replaces the
// whole config, so a reader takes one snapshot via Config and uses only that.
type MqttInfo struct {
	*store.Store
	configSnapshot[asyncapi3.Config]
	seenTopics            map[string]bool
	updateEventAndMetrics func(k *MqttInfo)
}

type MqttHandler struct {
	Mqtt *monitor.Mqtt
	next mqtt.Handler
}

func newMqttInfo(store *store.Store, updateEventAndMetrics func(info *MqttInfo)) *MqttInfo {
	hc := &MqttInfo{
		Store:                 store,
		seenTopics:            map[string]bool{},
		updateEventAndMetrics: updateEventAndMetrics,
	}
	return hc
}

// NewMqttInfo returns an info that publishes cfg as is, without merging files and
// without the implicit default broker of Add. It is meant for tests of readers: it has no
// event updater, so adding a file to it through MqttStore.Add panics.
func NewMqttInfo(cfg *asyncapi3.Config, store *store.Store) *MqttInfo {
	mi := newMqttInfo(store, nil)
	mi.publish(cfg)
	return mi
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
		s.events.ResetStores(events.NewTraits().WithNamespace("mqtt").WithName(cfg.Info.Name))
		s.events.SetStore(int(eventStore.Size), events.NewTraits().WithNamespace("mqtt").WithName(cfg.Info.Name))

		ki = newMqttInfo(store.New(cfg, emitter, s.events, s.monitor.Mqtt), s.updateEventStore)
		s.infos[cfg.Info.Name] = ki
	}
	ki.afterPublish(ki.add(c, mergeMqttConfigs(s.reader)))

	if s.cfg.Api.Search.Enabled {
		s.addToIndex(ki.Config())
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
	s.m.Lock()
	defer s.m.Unlock()

	cfg, err := getMqttConfig(c)
	if err != nil {
		return
	}
	name := cfg.Info.Name
	mi := s.infos[name]
	if mi == nil {
		return
	}

	if s.cfg.Api.Search.Enabled {
		s.removeFromIndex(mi.Config())
	}

	mi.afterPublish(mi.remove(c, mergeMqttConfigs(s.reader)))

	if remaining := mi.Config(); remaining == nil {
		delete(s.infos, name)
		s.events.ResetStores(events.NewTraits().WithNamespace("mqtt").WithName(name))
	} else if s.cfg.Api.Search.Enabled {
		s.addToIndex(remaining)
	}
}

// mergeMqttConfigs merges the files of a service and adds the default server if the spec
// defines none.
func mergeMqttConfigs(reader dynamic.Reader) func(files []*dynamic.Config) *asyncapi3.Config {
	return func(files []*dynamic.Config) *asyncapi3.Config {
		cfg := mergeAsyncApiConfigs(files, reader)
		defaultAsyncApiServer(cfg, "mqtt", "mqtt: using implicit default broker", &asyncapi3.Server{
			Host:     ":1883",
			Protocol: "mqtt",
			Title:    "Mokapi Default Broker",
			Summary:  "Automatically added broker because no servers are defined in the AsyncAPI spec",
		})
		return cfg
	}
}

func (c *MqttInfo) afterPublish(cfg *asyncapi3.Config) {
	if cfg == nil {
		return
	}
	c.updateEventAndMetrics(c)
	c.Store.Update(cfg)
}

func (c *MqttInfo) Handler(Mqtt *monitor.Mqtt) mqtt.Handler {
	return &MqttHandler{Mqtt: Mqtt, next: c.Store}
}

func (h *MqttHandler) ServeMessage(rw mqtt.MessageWriter, req *mqtt.Message) {
	ctx := monitor.NewMqttContext(req.Context, h.Mqtt)

	req.WithContext(ctx)
	h.next.ServeMessage(rw, req)
}

func getMqttConfig(c *dynamic.Config) (*asyncapi3.Config, error) {
	data := c.Data()
	if cfg, ok := data.(*asyncapi3.Config); ok {
		return cfg, nil
	}
	return data.(*asyncApi.Config).Convert()
}

func (s *MqttStore) updateEventStore(k *MqttInfo) {
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
		s.monitor.Mqtt.Messages.WithLabel(cfg.Info.Name, topicName)
		s.monitor.Mqtt.LastMessage.WithLabel(cfg.Info.Name, topicName)
		traits := events.NewTraits().WithNamespace("mqtt").WithName(cfg.Info.Name).With("topic", topicName)
		s.events.SetStore(int(eventStore.Size), traits)
		k.seenTopics[topicName] = true
	}
}

func (s *MqttStore) Len() int {
	if s == nil {
		return 0
	}

	s.m.RLock()
	defer s.m.RUnlock()
	return len(s.infos)
}

type MqttMode int

const (
	MqttNone MqttMode = iota
	// MqttExplicit Server with protocol mqtt available
	MqttExplicit
	// MqttImplicit no servers, MQTT derived from bindings
	MqttImplicit
)

func ShouldStartMqtt(c *dynamic.Config) (*asyncapi3.Config, MqttMode) {
	cfg, ok := IsAsyncApiConfig(c)
	if !ok {
		return nil, MqttNone
	}
	a := cfg.Analyze()
	k := a.Usage("mqtt")
	switch {
	case len(k.Servers) > 0:
		return cfg, MqttExplicit
	case !a.HasServers && len(k.Channels) > 0:
		return cfg, MqttImplicit
	}
	return cfg, MqttNone
}
