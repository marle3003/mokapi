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
	"net/http"
	"sync"
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
	configSnapshot[asyncapi3.Config]
	seenTopics            map[string]bool
	updateEventAndMetrics func(k *WebsocketInfo)
}

type WebsocketHandler struct {
	Websocket *monitor.Websocket
	next      http.Handler
}

func newWebsocketInfo(store *websocket.Store, updateEventAndMetrics func(info *WebsocketInfo)) *WebsocketInfo {
	hc := &WebsocketInfo{
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
	wi.publish(cfg)
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
		s.events.ResetStores(events.NewTraits().WithNamespace("websocket").WithName(cfg.Info.Name))
		s.events.SetStore(int(eventStore.Size), events.NewTraits().WithNamespace("websocket").WithName(cfg.Info.Name))

		wi = newWebsocketInfo(websocket.New(cfg, emitter, s.events, s.monitor.Websocket), s.updateEventStore)
		s.infos[cfg.Info.Name] = wi
	}
	wi.afterPublish(wi.add(c, mergeWebsocketConfigs(s.reader)))

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

	mi.afterPublish(mi.remove(c, mergeWebsocketConfigs(s.reader)))

	if remaining := mi.Config(); remaining == nil {
		delete(s.infos, name)
		s.events.ResetStores(events.NewTraits().WithNamespace("websocket").WithName(name))
	} else if s.cfg.Api.Search.Enabled {
		s.addToIndex(remaining)
	}
}

// mergeWebsocketConfigs merges the files of a service and adds the default server if the spec
// defines none.
func mergeWebsocketConfigs(reader dynamic.Reader) func(files []*dynamic.Config) *asyncapi3.Config {
	return func(files []*dynamic.Config) *asyncapi3.Config {
		cfg := mergeAsyncApiConfigs(files, reader)
		defaultAsyncApiServer(cfg, "ws", "websocket: using implicit default server", &asyncapi3.Server{
			Host:     ":80",
			Protocol: "ws",
			Title:    "Mokapi Default Server",
			Summary:  "Automatically added server because no servers are defined in the AsyncAPI spec",
		})
		return cfg
	}
}

func (c *WebsocketInfo) afterPublish(cfg *asyncapi3.Config) {
	if cfg == nil {
		return
	}
	c.updateEventAndMetrics(c)
	c.Store.Update(cfg)
}

func (c *WebsocketInfo) Handler(ws *monitor.Websocket) http.Handler {
	return &WebsocketHandler{Websocket: ws, next: c.Store}
}

func (h *WebsocketHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := monitor.NewWebsocketContext(r.Context(), h.Websocket)

	h.next.ServeHTTP(rw, r.WithContext(ctx))
}

func getWebsocketConfig(c *dynamic.Config) (*asyncapi3.Config, error) {
	data := c.Data()
	if cfg, ok := data.(*asyncapi3.Config); ok {
		return cfg, nil
	}
	return data.(*asyncApi.Config).Convert()
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
