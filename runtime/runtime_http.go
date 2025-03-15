package runtime

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/providers/openapi"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"net/http"
	"path/filepath"
	"sort"
	"sync"
)

type HttpStore struct {
	infos map[string]*HttpInfo
	m     sync.RWMutex
}

type HttpInfo struct {
	*openapi.Config
	configs   map[string]*dynamic.Config
	seenPaths map[string]bool
}

type httpHandler struct {
	http *monitor.Http
	next http.Handler
}

func NewHttpInfo(c *dynamic.Config) *HttpInfo {
	hc := &HttpInfo{
		configs:   map[string]*dynamic.Config{},
		seenPaths: map[string]bool{},
	}
	hc.AddConfig(c)
	return hc
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
		hc = NewHttpInfo(c)
		s.infos[cfg.Info.Name] = hc

		events.ResetStores(events.NewTraits().WithNamespace("http").WithName(name))
		events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("http").WithName(name))
	} else {
		hc.AddConfig(c)
	}

	for path := range cfg.Paths {
		if _, ok := hc.seenPaths[path]; ok {
			continue
		}
		events.SetStore(sizeEventStore, events.NewTraits().WithNamespace("http").WithName(name).With("path", path))
		hc.seenPaths[path] = true
	}

	return hc
}

func (s *HttpStore) Remove(c *dynamic.Config) {
	s.m.RLock()

	cfg := c.Data.(*openapi.Config)
	name := cfg.Info.Name
	hc := s.infos[name]
	hc.Remove(c)
	if len(hc.configs) == 0 {
		s.m.RUnlock()
		s.m.Lock()
		delete(s.infos, name)
		events.ResetStores(events.NewTraits().WithNamespace("http").WithName(name))
		s.m.Unlock()
	} else {
		s.m.RUnlock()
	}
}

func (c *HttpInfo) AddConfig(config *dynamic.Config) {
	c.configs[config.Info.Url.String()] = config
	c.update()
}

func (c *HttpInfo) Handler(http *monitor.Http, emitter common.EventEmitter) http.Handler {
	cfg := c.Config
	h := openapi.NewHandler(cfg, emitter)
	return &httpHandler{http: http, next: h}
}

func (c *HttpInfo) update() {
	if len(c.configs) == 0 {
		c.Config = nil
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
	*r = *getHttpConfig(c.configs[keys[0]])
	for _, k := range keys[1:] {
		p := getHttpConfig(c.configs[k])
		log.Infof("applying patch for %s: %s", c.Info.Name, k)
		r.Patch(p)
	}

	if len(r.Servers) == 0 {
		r.Servers = append(r.Servers, &openapi.Server{Url: "/"})
	}

	c.Config = r
}

func (c *HttpInfo) Configs() []*dynamic.Config {
	var r []*dynamic.Config
	for _, config := range c.configs {
		r = append(r, config)
	}
	return r
}

func (c *HttpInfo) Remove(cfg *dynamic.Config) {
	delete(c.configs, cfg.Info.Url.String())
	c.update()
}

func (h *httpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := monitor.NewHttpContext(r.Context(), h.http)

	h.next.ServeHTTP(rw, r.WithContext(ctx))
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
