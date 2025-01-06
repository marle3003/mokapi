package runtime

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/providers/openapi"
	"mokapi/runtime/monitor"
	"net/http"
	"path/filepath"
	"sort"
)

type HttpInfo struct {
	*openapi.Config
	configs map[string]*dynamic.Config
}

type httpHandler struct {
	http *monitor.Http
	next http.Handler
}

func NewHttpInfo(c *dynamic.Config) *HttpInfo {
	hc := &HttpInfo{
		configs: map[string]*dynamic.Config{},
	}
	hc.AddConfig(c)
	return hc
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
	var keys []string
	for k := range c.configs {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		x := keys[i]
		y := keys[j]
		return filepath.Base(x) < filepath.Base(y)
	})

	r := getHttpConfig(c.configs[keys[0]])
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

func (h *httpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := monitor.NewHttpContext(r.Context(), h.http)

	h.next.ServeHTTP(rw, r.WithContext(ctx))
}

func IsHttpConfig(c *dynamic.Config) bool {
	switch c.Data.(type) {
	case *openapi.Config:
		return true
	default:
		return false
	}
}

func getHttpConfig(c *dynamic.Config) *openapi.Config {
	return c.Data.(*openapi.Config)
}
