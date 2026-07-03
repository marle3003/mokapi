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
	"path/filepath"
	"sort"
	"strings"
	"sync"

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

type WebsocketInfo struct {
	*asyncapi3.Config
	*websocket.Store
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

func (s *WebsocketStore) Add(c *dynamic.Config, emitter common.EventEmitter) (*WebsocketInfo, error) {
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

		wi = newWebsocketInfo(websocket.New(cfg, emitter, s.events, s.monitor.Mqtt), s.updateEventStore)
		s.infos[cfg.Info.Name] = wi
	}
	wi.addConfig(c, s.reader)

	if s.cfg.Api.Search.Enabled {
		s.addToIndex(wi.Config)
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
	s.m.RLock()

	cfg, err := getWebsocketConfig(c)
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
		s.events.ResetStores(events.NewTraits().WithNamespace("Mqtt").WithName(name))
		s.m.Unlock()
	} else {
		s.m.RUnlock()
	}
}

func (c *WebsocketInfo) addConfig(config *dynamic.Config, reader dynamic.Reader) {
	key := config.Info.Url.String()
	c.configs[key] = config
	c.update(reader)
}

func (c *WebsocketInfo) update(reader dynamic.Reader) {
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
		log.Infof("no servers defined in AsyncAPI spec — using default Mokapi server for service '%s'", cfg.Info.Name)
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

	c.Config = cfg
	c.updateEventAndMetrics(c)
	c.Store.Update(cfg)
}

func (c *WebsocketInfo) Handler(ws *monitor.Websocket) http.Handler {
	return &WebsocketHandler{Websocket: ws, next: c.Store}
}

func (c *WebsocketInfo) Configs() []*dynamic.Config {
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

func HasWebsocketServer(c *dynamic.Config) (*asyncapi3.Config, bool) {
	cfg, ok := IsAsyncApiConfig(c)
	if !ok {
		return nil, false
	}
	for it := cfg.Servers.Iter(); it.Next(); {
		s := it.Value()
		if s.Value == nil {
			continue
		}
		if strings.ToLower(s.Value.Protocol) == "ws" {
			return cfg, true
		}
	}
	return cfg, false
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
	eventStore, hasStoreConfig := s.cfg.Event.Store[k.Config.Info.Name]
	if !hasStoreConfig {
		eventStore = s.cfg.Event.Store["default"]
	}

	for channelName, topic := range k.Config.Channels {
		if topic.Value == nil {
			continue
		}
		if topic.Value.Address != "" {
			channelName = topic.Value.Address
		}
		if _, ok := k.seenTopics[channelName]; ok {
			continue
		}
		s.monitor.Mqtt.Messages.WithLabel(k.Config.Info.Name, channelName)
		s.monitor.Mqtt.LastMessage.WithLabel(k.Config.Info.Name, channelName)
		traits := events.NewTraits().WithNamespace("websocket").WithName(k.Config.Info.Name).With("channel", channelName)
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
