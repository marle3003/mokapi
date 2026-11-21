package runtime

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/providers/openapi"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/runtime/search"
	"net/http"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

type HttpStore struct {
	infos  map[string]*HttpInfo
	cfg    *static.Config
	m      sync.RWMutex
	index  search.Index
	events *events.StoreManager
	reader dynamic.Reader
}

// HttpInfo publishes the merged config as an immutable snapshot. A reload replaces the
// whole config, so a reader takes one snapshot via Config and uses only that.
type HttpInfo struct {
	config    atomic.Pointer[openapi.Config]
	m         sync.RWMutex // guards configs for Configs; writers also hold HttpStore.m, so update and Remove read configs without m
	configs   map[string]*dynamic.Config
	seenPaths map[string]bool
}

type httpHandler struct {
	http *monitor.Http
	next openapi.Handler
}

func (s *HttpStore) Get(name string) *HttpInfo {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.infos[name]
}

func (s *HttpStore) List() []*HttpInfo {
	if s == nil {
		return nil
	}

	s.m.RLock()
	defer s.m.RUnlock()

	var list []*HttpInfo
	for _, v := range s.infos {
		list = append(list, v)
	}
	return list
}

// NewHttpInfo returns an info that publishes cfg as is, without merging files and without
// the default server of Add. It is meant for tests of readers: adding a file to it through
// HttpStore.Add is not supported.
func NewHttpInfo(cfg *openapi.Config) *HttpInfo {
	hi := &HttpInfo{
		configs:   map[string]*dynamic.Config{},
		seenPaths: map[string]bool{},
	}
	hi.config.Store(cfg)
	return hi
}

func (s *HttpStore) Add(c *dynamic.Config) *HttpInfo {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*HttpInfo)
	}
	cfg := c.Data.(*openapi.Config)
	name := cfg.Info.Name
	hc, ok := s.infos[name]
	if !ok {
		hc = NewHttpInfo(nil)
		s.infos[name] = hc
	}

	store, hasStoreConfig := s.cfg.Event.Store[name]
	if !hasStoreConfig {
		store = s.cfg.Event.Store["default"]
	}

	if !ok {
		s.events.ResetStores(events.NewTraits().WithNamespace("http").WithName(name))
		s.events.SetStore(int(store.Size), events.NewTraits().WithNamespace("http").WithName(name))
	}
	if s.cfg.Api.Search.Enabled {
		s.removeFromIndex(cfg)
	}
	hc.addConfig(c, s.reader)

	if cfg.Paths != nil {
		for it := cfg.Paths.Iter(); it.Next(); {
			path := it.Key()
			if _, ok := hc.seenPaths[path]; ok {
				continue
			}
			s.events.SetStore(int(store.Size), events.NewTraits().WithNamespace("http").WithName(name).With("path", path))
			hc.seenPaths[path] = true
		}
	}

	if s.cfg.Api.Search.Enabled {
		s.addToIndex(hc.Config())
	}

	return hc
}

func (s *HttpStore) Set(name string, hi *HttpInfo) {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.infos) == 0 {
		s.infos = make(map[string]*HttpInfo)
	}

	s.infos[name] = hi
}

func (s *HttpStore) Remove(c *dynamic.Config) {
	s.m.Lock()
	defer s.m.Unlock()

	cfg := c.Data.(*openapi.Config)
	name := cfg.Info.Name
	hc := s.infos[name]
	if hc == nil {
		return
	}

	if old := hc.Config(); old != nil && s.cfg.Api.Search.Enabled {
		s.removeFromIndex(old)
	}
	hc.removeConfig(c, s.reader)
	if s.cfg.Api.Search.Enabled {
		s.addToIndex(hc.Config())
	}

	if hc.Config() == nil {
		delete(s.infos, name)
		s.events.ResetStores(events.NewTraits().WithNamespace("http").WithName(name))
	}
}

// Handler serves requests with cfg, the snapshot the caller already holds.
func (c *HttpInfo) Handler(cfg *openapi.Config, http *monitor.Http, emitter common.HttpEventEmitter, eh events.Handler) openapi.Handler {
	h := openapi.NewHandler(cfg, emitter, eh)
	return &httpHandler{http: http, next: h}
}

func (c *HttpInfo) addConfig(config *dynamic.Config, reader dynamic.Reader) {
	c.m.Lock()
	c.configs[config.Info.Url.String()] = config
	c.m.Unlock()
	c.update(reader)
}

func (c *HttpInfo) removeConfig(config *dynamic.Config, reader dynamic.Reader) {
	c.m.Lock()
	delete(c.configs, config.Info.Url.String())
	c.m.Unlock()
	c.update(reader)
}

// Config returns the current merged config, or nil after the last file was removed.
// The result must not be modified.
func (c *HttpInfo) Config() *openapi.Config {
	return c.config.Load()
}

func (c *HttpInfo) update(reader dynamic.Reader) {
	if len(c.configs) == 0 {
		c.config.Store(nil)
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

	r := &openapi.Config{}
	for i, k := range keys {
		if i > 0 {
			log.Infof("applying patch for %s: %s", r.Info.Name, k)
		}
		r.Patch(getHttpConfig(c.configs[k]))
	}

	if len(c.configs) > 1 {
		err := r.Parse(&dynamic.Config{Data: r}, reader)
		if err != nil {
			log.Errorf("failed to parse config: %s", err)
		}
	}

	if len(r.Servers) == 0 {
		r.Servers = append(r.Servers, &openapi.Server{Url: "/", IsDefault: true})
	}

	c.config.Store(r)
}

func (c *HttpInfo) Configs() []*dynamic.Config {
	c.m.RLock()
	defer c.m.RUnlock()

	var r []*dynamic.Config
	for _, config := range c.configs {
		r = append(r, config)
	}
	return r
}

func (h *httpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) *openapi.HttpError {
	ctx := monitor.NewHttpContext(r.Context(), h.http)

	return h.next.ServeHTTP(rw, r.WithContext(ctx))
}

func IsHttpConfig(c *dynamic.Config) (*openapi.Config, bool) {
	switch v := c.Data.(type) {
	case *openapi.Config:
		return v, true
	default:
		return nil, false
	}
}

func getHttpConfig(c *dynamic.Config) *openapi.Config {
	return c.Data.(*openapi.Config)
}

func (s *HttpStore) Len() int {
	if s == nil {
		return 0
	}

	s.m.RLock()
	defer s.m.RUnlock()
	return len(s.infos)
}

func (s *HttpStore) Webhook(name string, args common.WebhookArgs) (*openapi.Path, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	if args.Api == "" {
		var webhook *openapi.Path
		for _, info := range s.infos {
			cfg := info.Config()
			if cfg == nil {
				continue
			}
			if w, ok := cfg.Webhooks[name]; ok {
				if webhook != nil {
					return nil, fmt.Errorf("ambiguous webhook '%s': use args.api to refine", name)
				}
				webhook = w.Value
			}
		}
		if webhook != nil {
			return webhook, nil
		}
	} else {
		var cfg *openapi.Config
		if info, ok := s.infos[args.Api]; ok {
			cfg = info.Config()
		}
		if cfg == nil {
			return nil, fmt.Errorf("api not found: %s", args.Api)
		}
		if w, ok := cfg.Webhooks[name]; ok && w.Value != nil {
			return w.Value, nil
		}
	}
	return nil, fmt.Errorf("webhook not found: %s", name)
}
