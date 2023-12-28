package runtime

import (
	log "github.com/sirupsen/logrus"
	cfg "mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/swagger"
	"mokapi/engine/common"
	"mokapi/runtime/monitor"
	"net/http"
	"path/filepath"
	"sort"
)

type HttpInfo struct {
	*openapi.Config
	configs map[string]*cfg.Config
}

type httpHandler struct {
	http *monitor.Http
	next http.Handler
}

func NewHttpInfo(c *cfg.Config) *HttpInfo {
	hc := &HttpInfo{
		configs: map[string]*cfg.Config{},
	}
	hc.AddConfig(c)
	return hc
}

func (c *HttpInfo) AddConfig(config *cfg.Config) {
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

	r := getHttpConfig(c.configs[keys[0]])
	for _, k := range keys[1:] {
		p := getHttpConfig(c.configs[k])
		r.Patch(p)
	}

	if len(r.Servers) == 0 {
		r.Servers = append(r.Servers, &openapi.Server{Url: "/"})
	}

	c.Config = r
}

func (c *HttpInfo) Configs() []*cfg.Config {
	var r []*cfg.Config
	for _, config := range c.configs {
		r = append(r, config)
		r = append(r, config.Refs()...)
	}
	return r
}

func (h *httpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := monitor.NewHttpContext(r.Context(), h.http)

	h.next.ServeHTTP(rw, r.WithContext(ctx))
}

func IsHttpConfig(c *cfg.Config) bool {
	switch c.Data.(type) {
	case *openapi.Config:
		return true
	case *swagger.Config:
		return true
	default:
		return false
	}
}

func getHttpConfig(c *cfg.Config) *openapi.Config {
	if sw, ok := c.Data.(*swagger.Config); ok {
		oc, err := swagger.Convert(sw)
		if err != nil {
			log.Errorf("unable to convert swagger config to openapi: %v", err)
			return nil
		}
		return oc
	} else {
		return c.Data.(*openapi.Config)
	}
}
