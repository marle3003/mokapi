package runtime

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/websocket"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/runtime/search"
	"mokapi/sortedmap"
	"net/http"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

type WebsocketStore struct {
	infos   map[string]*WebsocketInfo
	monitor *monitor.Monitor
	cfg     *static.Config
	m       sync.RWMutex
	events  *events.StoreManager
	index   search.Index
	reader  dynamic.Reader
}

// WebsocketInfo publishes the merged config as an immutable snapshot. A reload replaces the
// whole config, so a reader takes one snapshot via Config and uses only that.
type WebsocketInfo struct {
	*websocket.Store
	config                atomic.Pointer[asyncapi3.Config]
	m                     sync.RWMutex // guards configs, which are only written under the store lock
	configs               map[string]*dynamic.Config
	seenTopics            map[string]bool
	updateEventAndMetrics func(k *WebsocketInfo)
}

type WebsocketHandler struct {
	Websocket *monitor.Websocket
	next      http.Handler
}

func newWebsocketInfo(store *websocket.Store, updateEventAndMetrics func(info *WebsocketInfo)) *WebsocketInfo {
	hc := &WebsocketInfo{
		configs:               map[string]*dynamic.Config{},
		Store:                 store,
		seenTopics:            map[string]bool{},
		updateEventAndMetrics: updateEventAndMetrics,
	}
	return hc
}

// NewWebsocketInfo returns an info that publishes cfg as is, without merging files and
// without the implicit default server of Add. It is meant for tests of readers.
func NewWebsocketInfo(cfg *asyncapi3.Config, store *websocket.Store) *WebsocketInfo {
	wi := newWebsocketInfo(store, nil)
	wi.config.Store(cfg)
	return wi
}

func (s *WebsocketStore) Get(name string) *WebsocketInfo {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.infos[name]
}

func (s *WebsocketStore) List() []*WebsocketInfo {
	if s == nil {
		return nil
	}

	s.m.RLock()
	defer s.m.RUnlock()

	var list []*WebsocketInfo
	for _, v := range s.infos {
		list = append(list, v)
	}
	return list
}

func (s *WebsocketStore) Add(c *dynamic.Config, emitter common.WebsocketEventEmitter) (*WebsocketInfo, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*WebsocketInfo)
	}
	cfg, err := getWebsocketConfig(c)
	if err != nil {
		return nil, err
	}

	name := cfg.Info.Name
	wi, ok := s.infos[name]

	eventStore, hasStoreConfig := s.cfg.Event.Store[name]
	if !hasStoreConfig {
		eventStore = s.cfg.Event.Store["default"]
	}

	if !ok {
		s.events.ResetStores(events.NewTraits().WithNamespace("mqtt").WithName(cfg.Info.Name))
		s.events.SetStore(int(eventStore.Size), events.NewTraits().WithNamespace("mqtt").WithName(cfg.Info.Name))

		wi = newWebsocketInfo(websocket.New(cfg, emitter, s.events, s.monitor.Websocket), s.updateEventStore)
		s.infos[cfg.Info.Name] = wi
	}
	wi.addConfig(c, s.reader)

	if s.cfg.Api.Search.Enabled {
		s.addToIndex(wi.Config())
	}

	return wi, nil
}

func (s *WebsocketStore) Set(name string, wi *WebsocketInfo) {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*WebsocketInfo)
	}

	s.infos[name] = wi
}

func (s *WebsocketStore) Remove(c *dynamic.Config) {
	s.m.Lock()
	defer s.m.Unlock()

	cfg, err := getWebsocketConfig(c)
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

	mi.removeConfig(c, s.reader)

	if len(mi.configs) == 0 {
		delete(s.infos, name)
		s.events.ResetStores(events.NewTraits().WithNamespace("Mqtt").WithName(name))
	}
}

func (c *WebsocketInfo) addConfig(config *dynamic.Config, reader dynamic.Reader) {
	c.m.Lock()
	c.configs[config.Info.Url.String()] = config
	c.m.Unlock()
	c.update(reader)
}

func (c *WebsocketInfo) removeConfig(config *dynamic.Config, reader dynamic.Reader) {
	c.m.Lock()
	delete(c.configs, config.Info.Url.String())
	c.m.Unlock()
	c.update(reader)
}

// Config returns the current merged config, or nil after the last file was removed.
// The result must not be modified.
func (c *WebsocketInfo) Config() *asyncapi3.Config {
	return c.config.Load()
}

func (c *WebsocketInfo) update(reader dynamic.Reader) {
	if len(c.configs) == 0 {
		c.config.Store(nil)
		return
	}

	cfg := mergeAsyncApiConfigs(c.configs, reader)

	a := cfg.Analyze()
	k := a.Usage("ws")
	if len(k.Servers) == 0 {
		log.Infof("websocket: using implicit default server for '%s' (inferred from bindings)", cfg.Info.Name)
		if cfg.Servers == nil {
			cfg.Servers = &sortedmap.LinkedHashMap[string, *asyncapi3.ServerRef]{}
		}
		cfg.Servers.Set("mokapi", &asyncapi3.ServerRef{
			Value: &asyncapi3.Server{
				Host:     ":80",
				Protocol: "ws",
				Title:    "Mokapi Default Server",
				Summary:  "Automatically added server because no servers are defined in the AsyncAPI spec",
			},
		})
	}

	c.config.Store(cfg)
	c.updateEventAndMetrics(c)
	c.Store.Update(cfg)
}

func (c *WebsocketInfo) Handler(ws *monitor.Websocket) http.Handler {
	return &WebsocketHandler{Websocket: ws, next: c.Store}
}

func (c *WebsocketInfo) Configs() []*dynamic.Config {
	c.m.RLock()
	defer c.m.RUnlock()

	var r []*dynamic.Config
	for _, config := range c.configs {
		r = append(r, config)
	}
	return r
}

func (h *WebsocketHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := monitor.NewWebsocketContext(r.Context(), h.Websocket)

	h.next.ServeHTTP(rw, r.WithContext(ctx))
}

func getWebsocketConfig(c *dynamic.Config) (*asyncapi3.Config, error) {
	if _, ok := c.Data.(*asyncapi3.Config); ok {
		return c.Data.(*asyncapi3.Config), nil
	} else {
		old := c.Data.(*asyncApi.Config)
		return old.Convert()
	}
}

func (s *WebsocketStore) updateEventStore(k *WebsocketInfo) {
	cfg := k.Config()
	eventStore, hasStoreConfig := s.cfg.Event.Store[cfg.Info.Name]
	if !hasStoreConfig {
		eventStore = s.cfg.Event.Store["default"]
	}

	for channelName, topic := range cfg.Channels {
		if topic.Value == nil {
			continue
		}
		if topic.Value.Address != "" {
			channelName = topic.Value.Address
		}
		if _, ok := k.seenTopics[channelName]; ok {
			continue
		}
		s.monitor.Websocket.Messages.WithLabel(cfg.Info.Name, channelName)
		s.monitor.Websocket.LastMessage.WithLabel(cfg.Info.Name, channelName)
		traits := events.NewTraits().WithNamespace("websocket").WithName(cfg.Info.Name).With("channel", channelName)
		s.events.SetStore(int(eventStore.Size), traits)
		k.seenTopics[channelName] = true
	}
}

func (s *WebsocketStore) Len() int {
	if s == nil {
		return 0
	}

	s.m.RLock()
	defer s.m.RUnlock()
	return len(s.infos)
}

type WebsocketMode int

const (
	WebsocketNone WebsocketMode = iota
	// WebsocketExplicit Server with protocol ws available
	WebsocketExplicit
	// WebsocketImplicit no servers, Websocket derived from bindings
	WebsocketImplicit
)

func ShouldStartWebsocket(c *dynamic.Config) (*asyncapi3.Config, WebsocketMode) {
	cfg, ok := IsAsyncApiConfig(c)
	if !ok {
		return nil, WebsocketNone
	}
	a := cfg.Analyze()
	k := a.Usage("ws")
	switch {
	case len(k.Servers) > 0:
		return cfg, WebsocketExplicit
	case !a.HasServers && len(k.Channels) > 0:
		return cfg, WebsocketImplicit
	}
	return cfg, WebsocketNone
}
